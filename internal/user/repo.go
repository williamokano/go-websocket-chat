package user

import "context"

// Repo defines the interface for user storage operations.
type Repo interface {
	Create(ctx context.Context, username, hashedPassword string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}
