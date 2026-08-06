package models

import (
	"example.com/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type PayrollBatchRequest struct {
	CompanyAccountID pgtype.UUID             `json:"company_account_id"`
	ScheduleDate     pgtype.Timestamptz      `json:"schedule_date"`
	Status           sqlc.PayrollBatchStatus `json:"status"`
	BatchName        pgtype.Text             `json:"batch_name"`
}

type PayrollRequest struct {
	BatchID           pgtype.UUID            `json:"batch_id"`
	EmployeeAccountID pgtype.UUID            `json:"employee_account_id"`
	Amount            int64                  `json:"amount"`
	Status            sqlc.PayrollItemStatus `json:"status"`
	Description       pgtype.Text            `json:"description"`
}

type PayrollBatch struct {
	ID               pgtype.UUID             `json:"id"`
	CompanyAccountID pgtype.UUID             `json:"company_account_id"`
	ScheduleDate     pgtype.Timestamptz      `json:"schedule_date"`
	Status           sqlc.PayrollBatchStatus `json:"status"`
	CreatedAt        pgtype.Timestamptz      `json:"created_at"`
	BatchName        pgtype.Text             `json:"batch_name"`
}

type Payroll struct {
	ID                pgtype.UUID            `json:"id"`
	BatchID           pgtype.UUID            `json:"batch_id"`
	EmployeeAccountID pgtype.UUID            `json:"employee_account_id"`
	Amount            int64                  `json:"amount"`
	Status            sqlc.PayrollItemStatus `json:"status"`
	CreatedAt         pgtype.Timestamptz     `json:"created_at"`
	Description       pgtype.Text            `json:"description"`
	CompanyAccountID  pgtype.UUID            `json:"company_account_id"`
	CompanyUserID     pgtype.UUID            `json:"company_user_id"`
}
