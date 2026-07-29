-- db/migrations/00004_create_entries.up.sql
CREATE TABLE
    entries (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        account_id UUID NOT NULL REFERENCES accounts (id),
        transfer_id UUID NOT NULL REFERENCES transfers (id),
        amount BIGINT NOT NULL,
        entry_type TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )