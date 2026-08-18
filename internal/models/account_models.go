package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusFrozen AccountStatus = "FROZEN"
	AccountStatusClosed AccountStatus = "CLOSED"
)

type Account struct {
	ID            pgtype.UUID        `json:"id"`
	UserID        pgtype.UUID        `json:"user_id"`
	AccountNumber string             `json:"account_number"`
	Currency      string             `json:"currency"`
	Balance       int64              `json:"balance"`
	Status        AccountStatus      `json:"status"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Fullname      string             `json:"full_name"`
}
