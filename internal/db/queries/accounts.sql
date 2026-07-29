-- db/queries/accounts.sql
-- name: CreateUserAccount :one
INSERT INTO
    accounts (account_number, user_id, currency)
VALUES
    ($1, $2, $3)
RETURNING
    *;

-- name: GetUserAccount :one
SELECT
    *
FROM
    accounts
    JOIN users ON accounts.user_id = users.id
WHERE
    user_id = $1;

-- name: GetAccount :one
SELECT
    *
FROM
    accounts
WHERE
    id = $1;

-- name: GetAccountByAccountNumber :one
SELECT
    *
FROM
    accounts
WHERE
    account_number = $1;

-- name: CheckIfAccountNumberExists :one 
SELECT
    EXISTS (
        SELECT
            1
        FROM
            accounts
        WHERE
            account_number = $1
    );

-- name: GetAccountForUpdate :one 
SELECT
    *
FROM
    accounts
WHERE
    id = $1 FOR
UPDATE;

-- name: AddBalance :exec
UPDATE accounts
SET
    balance = balance + $1
WHERE
    id = $2;

-- name: SubtractBalance :exec
UPDATE accounts
SET
    balance = balance - $1
WHERE
    id = $2;

-- name: ActivateAccount :exec 
UPDATE accounts
SET
    status = 'ACTIVE'
WHERE
    id = $1;