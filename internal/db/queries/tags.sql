-- db/migrations/tags.sql 

-- name: GetTag :one 
SELECT * FROM tags WHERE id = $1;

-- name: GetTags :many
SELECT * FROM tags ORDER BY id;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES ($1) RETURNING *;