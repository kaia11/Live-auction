package ws

import (
	"sync"
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
	mu     sync.RWMutex
	buffer []Message
}

func NewHub() *Hub {
	return &Hub{
		buffer: make([]Message, 0),
	}
}

func (h *Hub) Publish(roomID string, event string, payload any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	version := int64(len(h.buffer) + 1)
	h.buffer = append(h.buffer, Message{
		RoomID:     roomID,
		Event:      event,
		Payload:    payload,
		Version:    version,
		ServerTime: time.Now().Format(time.RFC3339),
	})
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
