-- db/migrations/00004_create_link_visited.up.sql
CREATE TABLE
    link_visits (
        id BIGSERIAL PRIMARY KEY,
        link_id BIGINT NOT NULL REFERENCES links (id) ON DELETE CASCADE,
        visited_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    )