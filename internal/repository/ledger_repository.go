package repository

import (
	"context"
	"fmt"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerRepository interface {
	WithTx(tx pgx.Tx) LedgerRepository
	CreateEntry(ctx context.Context, req models.Entries) (*models.Entries, error)
}

type ledgerRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewLedgerRepository(pool *pgxpool.Pool) LedgerRepository {
	return &ledgerRepository{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (l *ledgerRepository) WithTx(tx pgx.Tx) LedgerRepository {
	return &ledgerRepository{
		Queries: l.Queries.WithTx(tx),
		pool:    l.pool,
	}
}

func (l *ledgerRepository) CreateEntry(ctx context.Context, req models.Entries) (*models.Entries, error) {
	entry, err := l.Queries.CreateEntry(ctx, sqlc.CreateEntryParams{
		Amount:     req.Amount,
		AccountID:  req.AccountID,
		EntryType:  sqlc.EntryTypeEnum(req.EntryType),
		TransferID: req.TransactionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}
	return &models.Entries{
		Amount:        entry.Amount,
		AccountID:     entry.AccountID,
		TransactionID: entry.TransferID,
		EntryType:     string(entry.EntryType),
	}, nil
}
