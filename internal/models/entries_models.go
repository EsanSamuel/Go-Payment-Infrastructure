package models

import "github.com/jackc/pgx/v5/pgtype"

type Entries struct {
	Amount        int64       `json:"amount"`
	AccountID     pgtype.UUID `json:"account_id"`
	TransactionID pgtype.UUID `json:"transaction_id"`
	EntryType     string      `json:"entry_type"`
}
