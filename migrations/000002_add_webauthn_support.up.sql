-- Make password nullable for passwordless accounts
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password SET DEFAULT '';

-- WebAuthn credentials table
CREATE TABLE webauthn_credentials (
    id              BYTEA PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_key      BYTEA NOT NULL,
    attestation_type VARCHAR(64) NOT NULL DEFAULT '',
    aaguid          BYTEA NOT NULL DEFAULT '\x00000000000000000000000000000000',
    sign_count      BIGINT NOT NULL DEFAULT 0,
    transports      TEXT[] NOT NULL DEFAULT '{}',
    friendly_name   VARCHAR(255) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_webauthn_credentials_user_id ON webauthn_credentials (user_id);
