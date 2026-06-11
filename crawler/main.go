package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/time/rate"
)

//go:embed web
var webFiles embed.FS

type Result struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Links    int    `json:"links"`
	Bytes    int64  `json:"bytes"`
	Duration int64  `json:"durationMs"`
	Error    string `json:"error,omitempty"`
}
type CrawlRequest struct {
	URL         string `json:"url"`
	MaxPages    int    `json:"maxPages"`
	Concurrency int    `json:"concurrency"`
}
type Crawler struct {
	client  *http.Client
	limiter *rate.Limiter
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/crawl", crawlHandler)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	static, _ := fs.Sub(webFiles, "web")
	mux.Handle("/", http.FileServer(http.FS(static)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("Orbit crawler listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
func crawlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req CrawlRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req) != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if req.MaxPages < 1 || req.MaxPages > 30 {
		req.MaxPages = 12
	}
	if req.Concurrency < 1 || req.Concurrency > 8 {
		req.Concurrency = 4
	}
	start, err := normalize(req.URL)
	if err != nil || !publicURL(start) {
		http.Error(w, "Enter a public http or https URL.", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	results := crawl(ctx, start, req.MaxPages, req.Concurrency)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"startUrl": start.String(), "pages": results, "count": len(results)})
}
func crawl(ctx context.Context, start *url.URL, maxPages, concurrency int) []Result {
	c := Crawler{client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 5 || !publicURL(req.URL) {
			return errors.New("redirect blocked")
		}
		return nil
	}}, limiter: rate.NewLimiter(rate.Limit(4), 1)}
	jobs := make(chan *url.URL)
	out := make(chan struct {
		r     Result
		links []*url.URL
	})
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range jobs {
				r, links := c.fetch(ctx, u, start.Hostname())
				select {
				case out <- struct {
					r     Result
					links []*url.URL
				}{r, links}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	seen := map[string]bool{start.String(): true}
	queue := []*url.URL{start}
	results := []Result{}
	active := 0
	for len(results) < maxPages && (len(queue) > 0 || active > 0) {
		for len(queue) > 0 && active < concurrency && len(results)+active < maxPages {
			u := queue[0]
			queue = queue[1:]
			active++
			jobs <- u
		}
		select {
		case got := <-out:
			active--
			results = append(results, got.r)
			for _, u := range got.links {
				if !seen[u.String()] && len(seen) < maxPages*4 {
					seen[u.String()] = true
					queue = append(queue, u)
				}
			}
		case <-ctx.Done():
			queue = nil
			active = 0
		}
	}
	close(jobs)
	wg.Wait()
	sort.SliceStable(results, func(i, j int) bool { return i < j })
	return results
}
func (c *Crawler) fetch(ctx context.Context, u *url.URL, host string) (Result, []*url.URL) {
	started := time.Now()
	r := Result{URL: u.String()}
	if err := c.limiter.Wait(ctx); err != nil {
		r.Error = err.Error()
		return r, nil
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	req.Header.Set("User-Agent", "OrbitCourseCrawler/1.0")
	resp, err := c.client.Do(req)
	if err != nil {
		r.Error = err.Error()
		r.Duration = time.Since(started).Milliseconds()
		return r, nil
	}
	defer resp.Body.Close()
	r.Status = resp.StatusCode
	doc, err := html.Parse(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		r.Error = "Could not parse HTML"
		return r, nil
	}
	links := []*url.URL{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil && r.Title == "" {
			r.Title = strings.TrimSpace(n.FirstChild.Data)
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					x, err := u.Parse(a.Val)
					if err == nil && x.Hostname() == host && (x.Scheme == "http" || x.Scheme == "https") {
						x.Fragment = ""
						links = append(links, x)
						r.Links++
					}
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)
	if r.Title == "" {
		r.Title = "Untitled page"
	}
	r.Duration = time.Since(started).Milliseconds()
	r.Bytes = resp.ContentLength
	return r, links
}
func normalize(s string) (*url.URL, error) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Hostname() == "" {
		return nil, fmt.Errorf("bad url")
	}
	u.Fragment = ""
	return u, nil
}
func publicURL(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			return false
		}
	}
	return true
}
