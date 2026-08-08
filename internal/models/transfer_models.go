package models

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type TransferRequest struct {
	FromAccountID   pgtype.UUID     `json:"from_account_id"`
	ToAccountID     pgtype.UUID     `json:"to_account_id"`
	ToAccountNumber string          `json:"to_account_number"`
	Amount          int64           `json:"amount"`
	Narration       pgtype.Text     `json:"narration"`
	IdempotencyKey  string          `json:"idempotency_key"`
	Method          string          `json:"-"`
	Path            string          `json:"-"`
	Body            json.RawMessage `json:"-"`
}

type Transfer struct {
	ID            pgtype.UUID `json:"id"`
	FromAccountID pgtype.UUID `json:"from_account_id"`
	ToAccountID   pgtype.UUID `json:"to_account_id"`
	Amount        int64       `json:"amount"`
	Reference     string      `json:"reference"`
	Status        string      `json:"status"`
	Description   pgtype.Text `json:"description"`
	CreatedAt     time.Time   `json:"created_at"`
}

type IdempotentRequest struct {
	IdempotencyKey string
	UserID         pgtype.UUID
	RequestHash    string
}

type IdempotentResponse struct {
	IdempotencyKey string             `json:"idempotency_key"`
	UserID         pgtype.UUID        `json:"user_id"`
	RequestHash    string             `json:"request_hash"`
	StatusCode     pgtype.Int4        `json:"status_code"`
	ResponseBody   []byte             `json:"response_body,omitempty"`
	LockedAt       pgtype.Timestamptz `json:"locked_at,omitempty"`
	CompletedAt    pgtype.Timestamptz `json:"completed_at,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	ExpiresAt      time.Time          `json:"expires_at"`
}

func (r *IdempotentResponse) IsCompleted() bool {
	return r.CompletedAt.Valid
}

func (r *IdempotentResponse) IsStale(staleAfter time.Duration) bool {
	if !r.LockedAt.Valid {
		return false
	}
	return time.Since(r.LockedAt.Time) > staleAfter
}
