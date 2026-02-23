package webauthn

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is a PostgreSQL-backed implementation of Repo.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, cred *Credential) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webauthn_credentials (id, user_id, public_key, attestation_type, aaguid, sign_count, transports, friendly_name, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		cred.ID, cred.UserID, cred.PublicKey, cred.AttestationType,
		cred.AAGUID, cred.SignCount, cred.Transports, cred.FriendlyName, cred.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating webauthn credential: %w", err)
	}
	return nil
}

func (r *Repository) FindByCredentialID(ctx context.Context, credentialID []byte) (*Credential, error) {
	var cred Credential
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, public_key, attestation_type, aaguid, sign_count, transports, friendly_name, created_at, last_used_at
		 FROM webauthn_credentials WHERE id = $1`,
		credentialID,
	).Scan(&cred.ID, &cred.UserID, &cred.PublicKey, &cred.AttestationType,
		&cred.AAGUID, &cred.SignCount, &cred.Transports, &cred.FriendlyName, &cred.CreatedAt, &cred.LastUsedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("credential not found")
		}
		return nil, fmt.Errorf("finding credential: %w", err)
	}
	return &cred, nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) ([]*Credential, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, public_key, attestation_type, aaguid, sign_count, transports, friendly_name, created_at, last_used_at
		 FROM webauthn_credentials WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying credentials: %w", err)
	}
	defer rows.Close()

	var creds []*Credential
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.ID, &c.UserID, &c.PublicKey, &c.AttestationType,
			&c.AAGUID, &c.SignCount, &c.Transports, &c.FriendlyName, &c.CreatedAt, &c.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scanning credential: %w", err)
		}
		creds = append(creds, &c)
	}
	return creds, rows.Err()
}

func (r *Repository) UpdateSignCount(ctx context.Context, credentialID []byte, signCount uint32) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE webauthn_credentials SET sign_count = $1, last_used_at = $2 WHERE id = $3`,
		signCount, now, credentialID,
	)
	if err != nil {
		return fmt.Errorf("updating sign count: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, credentialID []byte, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM webauthn_credentials WHERE id = $1 AND user_id = $2`,
		credentialID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting credential: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("credential not found")
	}
	return nil
}
