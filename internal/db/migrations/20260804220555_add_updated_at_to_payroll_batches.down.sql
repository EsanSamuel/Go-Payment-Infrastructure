-- db/migrations/20260804220555_add_updated_at_to_payroll_batches.down.sql
ALTER TABLE payroll_batches
DROP COLUMN IF EXISTS updated_at;