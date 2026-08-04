-- db/migrations/000006_create_payroll.up.sql
CREATE TYPE payroll_batch_status AS ENUM('PENDING', 'PROCESSING', 'COMPLETED', 'PARTIAL');

CREATE TYPE payroll_item_status AS ENUM('PENDING', 'COMPLETED', 'FAILED');

CREATE TYPE payroll_payment_interval AS ENUM('MONTHLY', 'WEEKLY', 'BIWEEKLY');

CREATE TABLE
    payroll_batches (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        company_account_id UUID NOT NULL REFERENCES accounts (id),
        schedule_date TIMESTAMPTZ NOT NULL,
        status payroll_batch_status NOT NULL DEFAULT 'PENDING',
        payment_interval payroll_payment_interval NOT NULL DEFAULT 'MONTHLY',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE TABLE
    payroll (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        batch_id UUID NOT NULL REFERENCES payroll_batches (id),
        employee_account_id UUID NOT NULL REFERENCES accounts (id),
        amount BIGINT NOT NULL,
        status payroll_item_status NOT NULL DEFAULT 'PENDING',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

ALTER TABLE payroll_batches
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();