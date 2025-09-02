package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Client struct {
	id     string
	groups map[string]bool
	ch     chan string
	done   chan struct{}
}

type Hub struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	groups   map[string]map[string]bool // group -> clientID set
	interval time.Duration
	retryMs  int
}

func NewHub(interval time.Duration) *Hub {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Hub{clients: make(map[string]*Client), groups: make(map[string]map[string]bool), interval: interval, retryMs: 5000}
}

func (h *Hub) AddClient(id string) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	c := &Client{id: id, groups: make(map[string]bool), ch: make(chan string, 64), done: make(chan struct{})}
	h.clients[id] = c
	return c
}

func (h *Hub) RemoveClient(id string) {
	h.mu.Lock()
	if c, ok := h.clients[id]; ok {
		close(c.done)
		for g := range c.groups {
			delete(h.groups[g], id)
		}
		delete(h.clients, id)
	}
	h.mu.Unlock()
}

func (h *Hub) Join(id, group string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[id]
	if !ok {
		return
	}
	c.groups[group] = true
	if h.groups[group] == nil {
		h.groups[group] = make(map[string]bool)
	}
	h.groups[group][id] = true
}

func (h *Hub) Leave(id, group string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[id]
	if !ok {
		return
	}
	delete(c.groups, group)
	if h.groups[group] != nil {
		delete(h.groups[group], id)
	}
}

func (h *Hub) Broadcast(data string)       { h.sendAll(formatData(data)) }
func (h *Hub) BroadcastJSON(v interface{}) { b, _ := json.Marshal(v); h.sendAll(formatData(string(b))) }
func (h *Hub) SendTo(id, data string) {
	h.mu.RLock()
	if c := h.clients[id]; c != nil {
		select {
		case c.ch <- formatData(data):
		default:
		}
	}
	h.mu.RUnlock()
}
func (h *Hub) SendToJSON(id string, v interface{}) { b, _ := json.Marshal(v); h.SendTo(id, string(b)) }
func (h *Hub) SendToGroup(group, data string) {
	h.mu.RLock()
	ids := h.groups[group]
	for id := range ids {
		if c := h.clients[id]; c != nil {
			select {
			case c.ch <- formatData(data):
			default:
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) sendAll(msg string) {
	h.mu.RLock()
	for _, c := range h.clients {
		select {
		case c.ch <- msg:
		default:
		}
	}
	h.mu.RUnlock()
}

func formatData(s string) string { return fmt.Sprintf("data: %s\n\n", s) }

func (h *Hub) Serve(c *gin.Context, clientID string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	fmt.Fprintf(c.Writer, "retry: %d\n\n", h.retryMs)

	client := h.AddClient(clientID)
	defer h.RemoveClient(clientID)
	if gid := c.Query("group"); gid != "" {
		h.Join(clientID, gid)
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}
	ping := time.NewTicker(h.interval)
	defer ping.Stop()
	c.Stream(func(w io.Writer) bool { return true })

	lastEventID := c.GetHeader("Last-Event-ID")
	_ = lastEventID // 留接口：可接入历史缓存重放

	for {
		select {
		case <-client.done:
			return
		case <-c.Request.Context().Done():
			return
		case <-ping.C:
			fmt.Fprintf(c.Writer, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		case msg := <-client.ch:
			c.Writer.Write([]byte(msg))
			flusher.Flush()
		}
	}
}
