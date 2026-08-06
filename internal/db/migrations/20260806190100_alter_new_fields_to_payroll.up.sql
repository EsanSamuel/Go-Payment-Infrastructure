-- db/migrations/20260806190100_alter_new_fields_to_payroll.up.sql
ALTER TABLE payroll
ADD COLUMN description TEXT;

ALTER TABLE payroll_batches
ADD COLUMN batch_name TEXT;