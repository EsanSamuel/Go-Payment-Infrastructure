-- db/migrations/000005_create_idempotency_key.up.sql
CREATE TABLE
    idempotency_keys (
        idempotency_key TEXT PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        request_hash TEXT NOT NULL, -- hash of method+path+body to detect key reuse with different payload
        status_code INT,
        response_body JSONB,
        locked_at TIMESTAMPTZ, -- non-null while request is in-flight
        completed_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '24 hours'
    );

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

CREATE INDEX idx_idempotency_keys_user_id ON idempotency_keys (user_id);