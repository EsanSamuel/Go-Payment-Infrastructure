-- db/queries/entries.sql
-- name: CreateEntry :one
INSERT INTO
    entries (account_id, transfer_id, amount, entry_type)
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetEntriesByAccountID :many
SELECT
    *
FROM
    entries
WHERE
    account_id = $1;

-- name: GetAccountIdCreditEntry :one
SELECT
    *
FROM
    entries
WHERE
    account_id = $1
    AND entry_type = 'CREDIT';

-- name: GetAccountIdDEBITEntry :one
SELECT
    *
FROM
    entries
WHERE
    account_id = $1
    AND entry_type = 'DEBIT';