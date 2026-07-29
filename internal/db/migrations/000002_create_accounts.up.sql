-- db/migrations/000002_create_accounts.up.sql
CREATE TYPE account_status AS ENUM ('ACTIVE', 'FROZEN', 'CLOSED');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    account_number VARCHAR NOT NULL UNIQUE,
    currency CHAR(3) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    status account_status NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    UNIQUE (user_id, currency)
);