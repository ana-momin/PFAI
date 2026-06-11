package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed web
var webFiles embed.FS

type Event struct {
	Type    string   `json:"type"`
	User    string   `json:"user,omitempty"`
	Text    string   `json:"text,omitempty"`
	Time    string   `json:"time,omitempty"`
	Users   []string `json:"users,omitempty"`
	UserID  string   `json:"userId,omitempty"`
	Message string   `json:"message,omitempty"`
}

type Client struct {
	id   string
	name string
	conn *websocket.Conn
	send chan Event
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	history []Event
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) broadcast(event Event) {
	h.mu.Lock()
	if event.Type == "message" {
		h.history = append(h.history, event)
		if len(h.history) > 40 {
			h.history = h.history[len(h.history)-40:]
		}
	}
	for client := range h.clients {
		select {
		case client.send <- event:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) presence() {
	h.mu.RLock()
	users := make([]string, 0, len(h.clients))
	for client := range h.clients {
		users = append(users, client.name)
	}
	h.mu.RUnlock()
	h.broadcast(Event{Type: "presence", Users: users})
}

func (h *Hub) serveWS(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "Guest"
	}
	if len(name) > 24 {
		name = name[:24]
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{id: time.Now().Format("150405.000"), name: name, conn: conn, send: make(chan Event, 32)}
	h.mu.Lock()
	h.clients[client] = true
	history := append([]Event(nil), h.history...)
	h.mu.Unlock()

	go client.writePump()
	for _, event := range history {
		client.send <- event
	}
	h.broadcast(Event{Type: "system", Text: name + " joined the room", Time: clock()})
	h.presence()

	defer func() {
		h.mu.Lock()
		if h.clients[client] {
			delete(h.clients, client)
			close(client.send)
		}
		h.mu.Unlock()
		h.broadcast(Event{Type: "system", Text: name + " left the room", Time: clock()})
		h.presence()
		conn.Close()
	}()

	conn.SetReadLimit(2048)
	for {
		var incoming Event
		if err := conn.ReadJSON(&incoming); err != nil {
			break
		}
		text := strings.TrimSpace(incoming.Text)
		if text == "" {
			continue
		}
		if len(text) > 500 {
			text = text[:500]
		}
		h.broadcast(Event{Type: "message", User: name, UserID: client.id, Text: text, Time: clock()})
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(event); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func clock() string { return time.Now().Format("15:04") }

func main() {
	hub := &Hub{clients: make(map[*Client]bool)}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.serveWS)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		hub.mu.RLock()
		count := len(hub.clients)
		hub.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok", "connections": count})
	})
	static, _ := fs.Sub(webFiles, "web")
	mux.Handle("/", http.FileServer(http.FS(static)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Signal chat listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
