-- db/migrations/20260804220555_add_updated_at_to_payroll_batches.up.sql
ALTER TABLE payroll_batches
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();