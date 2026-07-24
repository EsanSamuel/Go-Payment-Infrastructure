-- db/queries/users.sql

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (email) VALUES ($1) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;