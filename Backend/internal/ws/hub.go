package ws

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	RoomID     string `json:"roomId"`
	Event      string `json:"event"`
	Payload    any    `json:"payload"`
	Version    int64  `json:"version"`
	ServerTime string `json:"serverTime"`
}

type Hub struct {
	mu        sync.RWMutex
	buffer    []Message
	clients   map[string]map[int64]*Client
	clientSeq int64
}

func NewHub() *Hub {
	return &Hub{
		buffer:  make([]Message, 0),
		clients: make(map[string]map[int64]*Client),
	}
}

func (h *Hub) Publish(roomID string, event string, payload any) {
	h.mu.Lock()
	version := int64(len(h.buffer) + 1)
	message := Message{
		RoomID:     roomID,
		Event:      event,
		Payload:    payload,
		Version:    version,
		ServerTime: time.Now().Format(time.RFC3339),
	}
	h.buffer = append(h.buffer, message)

	roomClients := make([]*Client, 0)
	for _, client := range h.clients[roomID] {
		roomClients = append(roomClients, client)
	}
	h.mu.Unlock()

	for _, client := range roomClients {
		client.Send(message)
	}
}

func (h *Hub) List(roomID string, sinceVersion int64, limit int) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]Message, 0)
	for _, message := range h.buffer {
		if roomID != "" && message.RoomID != roomID {
			continue
		}
		if message.Version <= sinceVersion {
			continue
		}

		result = append(result, message)
		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

func (h *Hub) LatestVersion(roomID string) int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var latest int64
	for _, message := range h.buffer {
		if roomID != "" && message.RoomID != roomID {
			continue
		}
		latest = message.Version
	}

	return latest
}

func (h *Hub) Register(roomID string, client *Client) func() {
	id := atomic.AddInt64(&h.clientSeq, 1)

	h.mu.Lock()
	client.id = id
	if _, ok := h.clients[roomID]; !ok {
		h.clients[roomID] = make(map[int64]*Client)
	}
	h.clients[roomID][id] = client
	h.mu.Unlock()

	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		clients := h.clients[roomID]
		if clients == nil {
			return
		}

		delete(clients, id)
		if len(clients) == 0 {
			delete(h.clients, roomID)
		}
	}
}

type Client struct {
	id     int64
	roomID string
	send   chan Message
}

func NewClient(roomID string, queueSize int) *Client {
	if queueSize <= 0 {
		queueSize = 32
	}

	return &Client{
		roomID: roomID,
		send:   make(chan Message, queueSize),
	}
}

func (c *Client) RoomID() string {
	return c.roomID
}

func (c *Client) Messages() <-chan Message {
	return c.send
}

func (c *Client) Send(message Message) {
	select {
	case c.send <- message:
	default:
	}
}

func (c *Client) Close() {
}

var ErrClientClosed = errors.New("client closed")
