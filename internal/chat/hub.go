package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	redis      *RedisAdapter
}

func NewHub(redisURL string) (*Hub, error) {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	redis, err := NewRedisAdapter(redisURL, h)
	if err != nil {
		return nil, err
	}
	h.redis = redis

	return h, nil
}

func (h *Hub) Run(ctx context.Context) {
	go h.redis.Subscribe(ctx)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()

			msg := NewUserJoinedMessage(client.userID, client.username, count)
			data, _ := json.Marshal(msg)
			h.redis.Publish(ctx, data)

			slog.Info("client connected", "user", client.username, "online", count)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			count := len(h.clients)
			h.mu.Unlock()

			msg := NewUserLeftMessage(client.userID, client.username, count)
			data, _ := json.Marshal(msg)
			h.redis.Publish(ctx, data)

			slog.Info("client disconnected", "user", client.username, "online", count)

		case message := <-h.broadcast:
			h.redis.Publish(ctx, message)

		case <-ctx.Done():
			return
		}
	}
}

func (h *Hub) BroadcastLocal(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
