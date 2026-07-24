-- db/migrations/000003_create_tags.up.sql

CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
)