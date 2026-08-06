-- db/queries/payroll.sql
-- name: CreatePayrollBatch :one
INSERT INTO
    payroll_batches (
        batch_name,
        company_account_id,
        schedule_date,
        status
    )
VALUES
    ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreatePayroll :one
INSERT INTO
    payroll (
        batch_id,
        employee_account_id,
        amount,
        status,
        description
    )
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetPayrollBatchById :one
SELECT
    *
FROM
    payroll_batches
WHERE
    id = $1
    AND company_account_id = $2;

-- name: GetBatchById :one
SELECT
    *
FROM
    payroll_batches
WHERE
    id = $1;

-- name: GetDuePayrollBatches :many
SELECT
    *
FROM
    payroll_batches
WHERE
    schedule_date <= now()
    AND status = 'PENDING' FOR
UPDATE SKIP LOCKED;

-- name: UpdateBatchToProcessing :one 
UPDATE payroll_batches
SET
    status = 'PROCESSING'
WHERE
    id = $1
RETURNING
    *;

-- name: GetPendingBatchPayrollItem :many
SELECT
    *
FROM
    payroll
WHERE
    batch_id = $1
    AND status = 'PENDING';

-- name: GetPayrollItemByID :one
SELECT
    payroll.*,
    payroll_batches.*,
    accounts.*
FROM
    payroll
    JOIN payroll_batches ON payroll_batches.id = payroll.batch_id
    JOIN accounts ON accounts.id = payroll_batches.company_account_id
WHERE
    payroll.id = $1;

-- name: UpdatePayrollToFailed :one 
UPDATE payroll
SET
    status = 'FAILED'
WHERE
    id = $1
RETURNING
    *;

-- name: UpdatePayrollToCompleted :one 
UPDATE payroll
SET
    status = 'COMPLETED'
WHERE
    id = $1
RETURNING
    *;

-- name: CheckBatchCompleted :one
SELECT
    COUNT(*) AS TOTAL,
    COUNT(*) FILTER (
        WHERE
            status = 'COMPLETED'
    ) AS completed,
    COUNT(*) FILTER (
        WHERE
            status = 'FAILED'
    ) AS failed
FROM
    payroll
WHERE
    batch_id = $1;

-- name: GetBatchForUpdate :one
SELECT
    *
FROM
    payroll_batches
WHERE
    id = $1 FOR
UPDATE;

-- name: FinalizeBatch :one
UPDATE payroll_batches
SET
    status = $2,
    schedule_date = $3,
    updated_at = now()
WHERE
    id = $1
RETURNING
    *;