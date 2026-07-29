-- db/queries/users.sql
-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = $1;

-- name: ListUsers :many
SELECT
    *
FROM
    users
ORDER BY
    id;

-- name: CreateUser :one
INSERT INTO
    users (email, password_hash, full_name)
VALUES
    ($1, $2, $3)
RETURNING
    *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE
    id = $1;

-- name: VerifyUser :exec
UPDATE users
SET
    is_verified = TRUE
WHERE
    id = $1;