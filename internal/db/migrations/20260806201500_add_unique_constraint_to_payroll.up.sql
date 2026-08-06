-- db/migrations/20260806201500_add_unique_constraint_to_payroll.up.sql
ALTER TABLE payroll
ADD CONSTRAINT payroll_employee_batch_unique UNIQUE (batch_id, employee_account_id);