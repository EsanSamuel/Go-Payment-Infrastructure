-- db/migrations/20260806190100_alter_new_fields_to_payroll.down.sql

ALTER TABLE payroll_batches
DROP COLUMN batch_name;

ALTER TABLE payroll
DROP COLUMN description;