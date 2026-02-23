package webauthn

import "context"

// Repo defines storage operations for WebAuthn credentials.
type Repo interface {
	Create(ctx context.Context, cred *Credential) error
	FindByCredentialID(ctx context.Context, credentialID []byte) (*Credential, error)
	FindByUserID(ctx context.Context, userID string) ([]*Credential, error)
	UpdateSignCount(ctx context.Context, credentialID []byte, signCount uint32) error
	Delete(ctx context.Context, credentialID []byte, userID string) error
}
