-- db/migrations/20260806201500_add_unique_constraint_to_payroll.down.sql

ALTER TABLE payroll
DROP CONSTRAINT payroll_batch_employee_unique;