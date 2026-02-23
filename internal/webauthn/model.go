package webauthn

import "time"

// Credential represents a stored WebAuthn credential.
type Credential struct {
	ID              []byte    `json:"id"`
	UserID          string    `json:"user_id"`
	PublicKey       []byte    `json:"public_key"`
	AttestationType string    `json:"attestation_type"`
	AAGUID          []byte    `json:"aaguid"`
	SignCount       uint32    `json:"sign_count"`
	Transports      []string  `json:"transports"`
	FriendlyName    string    `json:"friendly_name"`
	CreatedAt       time.Time `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
}

// CredentialInfo is the public-facing representation returned by the list endpoint.
type CredentialInfo struct {
	ID           string     `json:"id"`
	FriendlyName string    `json:"friendly_name"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}
