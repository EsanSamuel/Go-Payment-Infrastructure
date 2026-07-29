package repository

import (
	"context"
	"fmt"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account_number string, currency string, user_id pgtype.UUID) (*models.Account, error)
}

type accountRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) AccountRepository {
	return &accountRepository{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (a *accountRepository) CreateAccount(ctx context.Context, account_number string, currency string, user_id pgtype.UUID) (*models.Account, error) {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := a.Queries.WithTx(tx)

	account, err := qtx.CreateUserAccount(ctx, sqlc.CreateUserAccountParams{
		AccountNumber: account_number,
		Currency:      currency,
		UserID:        user_id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user account: %w", err)

	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Account{
		ID:            account.ID,
		UserID:        account.UserID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
	}, nil
}
