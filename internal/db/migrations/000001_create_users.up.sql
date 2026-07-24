-- db/migrations/000001_create_users.up.sql
CREATE TABLE
  users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now ()
  );