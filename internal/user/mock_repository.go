package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockRepository is an in-memory implementation of Repo for testing.
type MockRepository struct {
	mu    sync.RWMutex
	users map[string]*User // keyed by ID
}

func NewMockRepository() *MockRepository {
	return &MockRepository{users: make(map[string]*User)}
}

func (r *MockRepository) Create(ctx context.Context, username, hashedPassword string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Username == username {
			return nil, fmt.Errorf("creating user: duplicate username")
		}
	}

	u := &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}
	r.users[u.ID] = u
	copy := *u
	return &copy, nil
}

func (r *MockRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			copy := *u
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("finding user: not found")
}

func (r *MockRepository) FindByID(ctx context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("finding user: not found")
	}
	copy := *u
	return &copy, nil
}
