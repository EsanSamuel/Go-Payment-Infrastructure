-- db/migrations/000003_create_transfers.up.sql
CREATE TABLE
    transfers (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        from_account_id UUID NOT NULL REFERENCES accounts (id),
        to_account_id UUID NOT NULL REFERENCES accounts (id),
        amount BIGINT NOT NULL,
        status TEXT NOT NULL DEFAULT 'PENDING',
        reference UUID NOT NULL UNIQUE DEFAULT gen_random_uuid (),
        description TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )