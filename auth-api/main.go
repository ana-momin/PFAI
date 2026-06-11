package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//go:embed web
var webFiles embed.FS

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
type Task struct {
	ID        string `json:"id"`
	UserID    string `json:"-"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	CreatedAt string `json:"createdAt"`
}
type Store struct {
	mu    sync.RWMutex
	users map[string]User
	tasks map[string][]Task
}
type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

var secret = []byte(env("JWT_SECRET", "pfai-local-development-secret-change-me"))

func main() {
	store := &Store{users: map[string]User{}, tasks: map[string][]Task{}}
	r := gin.Default()
	r.Use(cors())
	api := r.Group("/api")
	api.POST("/auth/register", store.register)
	api.POST("/auth/login", store.login)
	secured := api.Group("/")
	secured.Use(auth())
	secured.GET("/me", store.me)
	secured.GET("/tasks", store.listTasks)
	secured.POST("/tasks", store.createTask)
	secured.PATCH("/tasks/:id", store.updateTask)
	secured.DELETE("/tasks/:id", store.deleteTask)
	static, _ := fs.Sub(webFiles, "web")
	indexPage, _ := fs.ReadFile(static, "index.html")
	r.StaticFS("/assets", http.FS(static))
	serveIndex := func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", indexPage) }
	r.GET("/", serveIndex)
	r.NoRoute(serveIndex)
	r.Run(":" + env("PORT", "8082"))
}

func (s *Store) register(c *gin.Context) {
	var body struct{ Name, Email, Password string }
	if c.ShouldBindJSON(&body) != nil || len(strings.TrimSpace(body.Name)) < 2 || !strings.Contains(body.Email, "@") || len(body.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Use a name, valid email, and password of 6+ characters."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[email]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "That email is already registered."})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	user := User{ID: id(), Name: strings.TrimSpace(body.Name), Email: email, Password: string(hash)}
	s.users[email] = user
	c.JSON(http.StatusCreated, gin.H{"token": token(user.ID), "user": user})
}
func (s *Store) login(c *gin.Context) {
	var body struct{ Email, Password string }
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid request."})
		return
	}
	s.mu.RLock()
	user, ok := s.users[strings.ToLower(strings.TrimSpace(body.Email))]
	s.mu.RUnlock()
	if !ok || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		c.JSON(401, gin.H{"error": "Email or password is incorrect."})
		return
	}
	c.JSON(200, gin.H{"token": token(user.ID), "user": user})
}
func (s *Store) me(c *gin.Context) {
	uid := c.GetString("uid")
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == uid {
			c.JSON(200, u)
			return
		}
	}
	c.JSON(404, gin.H{"error": "User not found."})
}
func (s *Store) listTasks(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c.JSON(200, s.tasks[c.GetString("uid")])
}
func (s *Store) createTask(c *gin.Context) {
	var body struct{ Title string }
	if c.ShouldBindJSON(&body) != nil || strings.TrimSpace(body.Title) == "" {
		c.JSON(400, gin.H{"error": "Task title is required."})
		return
	}
	t := Task{ID: id(), UserID: c.GetString("uid"), Title: strings.TrimSpace(body.Title), CreatedAt: time.Now().Format(time.RFC3339)}
	s.mu.Lock()
	s.tasks[t.UserID] = append([]Task{t}, s.tasks[t.UserID]...)
	s.mu.Unlock()
	c.JSON(201, t)
}
func (s *Store) updateTask(c *gin.Context) {
	var body struct {
		Title     *string
		Completed *bool
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Invalid request."})
		return
	}
	uid := c.GetString("uid")
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks[uid] {
		if t.ID == c.Param("id") {
			if body.Title != nil {
				t.Title = strings.TrimSpace(*body.Title)
			}
			if body.Completed != nil {
				t.Completed = *body.Completed
			}
			s.tasks[uid][i] = t
			c.JSON(200, t)
			return
		}
	}
	c.JSON(404, gin.H{"error": "Task not found."})
}
func (s *Store) deleteTask(c *gin.Context) {
	uid := c.GetString("uid")
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks[uid] {
		if t.ID == c.Param("id") {
			s.tasks[uid] = append(s.tasks[uid][:i], s.tasks[uid][i+1:]...)
			c.Status(204)
			return
		}
	}
	c.JSON(404, gin.H{"error": "Task not found."})
}
func auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		claims := &Claims{}
		t, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) { return secret, nil })
		if err != nil || !t.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Valid bearer token required."})
			return
		}
		c.Set("uid", claims.UserID)
		c.Next()
	}
}
func token(uid string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: uid, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now())}})
	out, _ := t.SignedString(secret)
	return out
}
func id() string { b := make([]byte, 6); rand.Read(b); return hex.EncodeToString(b) }
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
