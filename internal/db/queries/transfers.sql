-- db/queries/transfers.sql
-- name: CreateTransfer :one
INSERT INTO
    transfers (
        from_account_id,
        to_account_id,
        amount,
        description,
        status
    )
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetTransferById :one 
SELECT
    *
FROM
    transfers
WHERE
    id = $1;

-- name: UpdateTransferStatus :exec
UPDATE transfers
SET
    status = $1
WHERE
    id = $2;

-- name: GetOutgoingTransfers :many 
SELECT
    *
FROM
    transfers
WHERE
    from_account_id = $1
ORDER BY
    created_at DESC
LIMIT
    $2
OFFSET
    $3;

-- name: GetIncomingTransfers :many 
SELECT
    *
FROM
    transfers
WHERE
    to_account_id = $1
ORDER BY
    created_at DESC
LIMIT
    $2
OFFSET
    $3;