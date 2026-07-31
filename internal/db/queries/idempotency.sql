-- db/queries/idempotency.sql
-- name: CreateIdempotencyKey :one
INSERT INTO
    idempotency_keys (idempotency_key, user_id, request_hash, locked_at)
VALUES
    ($1, $2, $3, now()) ON CONFLICT (idempotency_key)
DO NOTHING
RETURNING
    *;

-- name: CheckIfIdempotencyKeyExists :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            idempotency_keys
        WHERE
            idempotency_key = $1
    );

-- name: CompleteIdempotencyKey :one
UPDATE idempotency_keys
SET
    status_code = $2,
    response_body = $3,
    completed_at = now()
WHERE
    idempotency_key = $1
RETURNING
    *;

-- name: GetIdempotencyKey :one
SELECT
    *
FROM
    idempotency_keys
WHERE
    idempotency_key = $1
    AND user_id = $2;