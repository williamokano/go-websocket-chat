package chat

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

const chatChannel = "chat:broadcast"

type RedisAdapter struct {
	client *redis.Client
	hub    *Hub
}

func NewRedisAdapter(redisURL string, hub *Hub) (*RedisAdapter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	return &RedisAdapter{client: client, hub: hub}, nil
}

func (r *RedisAdapter) Publish(ctx context.Context, message []byte) {
	if err := r.client.Publish(ctx, chatChannel, message).Err(); err != nil {
		slog.Error("redis publish error", "error", err)
	}
}

func (r *RedisAdapter) Subscribe(ctx context.Context) {
	pubsub := r.client.Subscribe(ctx, chatChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			r.hub.BroadcastLocal([]byte(msg.Payload))
		case <-ctx.Done():
			return
		}
	}
}

func (r *RedisAdapter) Close() error {
	return r.client.Close()
}
