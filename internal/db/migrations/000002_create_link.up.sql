-- db/migrations/000002_create_links.up.sql
CREATE TABLE
    links (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        url TEXT NOT NULL,
        title TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    )