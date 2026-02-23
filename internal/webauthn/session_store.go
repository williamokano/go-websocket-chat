package webauthn

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "webauthn:session:"
	sessionTTL       = 5 * time.Minute
)

// SessionStore manages WebAuthn challenge sessions in Redis with a 5-minute TTL
// and one-time use (deleted after retrieval).
type SessionStore struct {
	client *redis.Client
}

func NewSessionStore(redisURL string) (*SessionStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	return &SessionStore{client: client}, nil
}

// Save stores session data and returns a session ID for later retrieval.
func (s *SessionStore) Save(ctx context.Context, session *webauthn.SessionData) (string, error) {
	sessionID := uuid.New().String()
	data, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("marshaling session: %w", err)
	}

	key := sessionKeyPrefix + sessionID
	if err := s.client.Set(ctx, key, data, sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("saving session: %w", err)
	}

	return sessionID, nil
}

// Get retrieves and deletes session data (one-time use).
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*webauthn.SessionData, error) {
	key := sessionKeyPrefix + sessionID

	data, err := s.client.GetDel(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("getting session: %w", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshaling session: %w", err)
	}

	return &session, nil
}

func (s *SessionStore) Close() error {
	return s.client.Close()
}
