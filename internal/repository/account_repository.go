package repository

import (
	"context"
	"fmt"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	WithTx(tx pgx.Tx) AccountRepository
	CreateAccount(ctx context.Context, account_number string, currency string, user_id pgtype.UUID) (*models.Account, error)
	GetAccountById(ctx context.Context, id pgtype.UUID) (*models.Account, error)
	GetAccountByAccountNumber(ctx context.Context, account_number string) (*models.Account, error)
	CheckIfAccountExists(ctx context.Context, account_number string) (bool, error)
	AddBalance(ctx context.Context, amount int64, account_id pgtype.UUID) error
	SubtractBalance(ctx context.Context, amount int64, account_id pgtype.UUID) error
	GetAccountForUpdate(ctx context.Context, account_id pgtype.UUID) (*models.Account, error)
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

func (a *accountRepository) WithTx(tx pgx.Tx) AccountRepository {
	return &accountRepository{
		Queries: a.Queries.WithTx(tx),
		pool:    a.pool,
	}
}

func (a *accountRepository) CreateAccount(ctx context.Context, account_number string, currency string, user_id pgtype.UUID) (*models.Account, error) {
	account, err := a.Queries.CreateUserAccount(ctx, sqlc.CreateUserAccountParams{
		AccountNumber: account_number,
		Currency:      currency,
		UserID:        user_id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user account: %w", err)

	}
	return &models.Account{
		ID:            account.ID,
		UserID:        account.UserID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
	}, nil
}

func (a *accountRepository) GetAccountById(ctx context.Context, id pgtype.UUID) (*models.Account, error) {
	account, err := a.Queries.GetAccount(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &models.Account{
		ID:            account.ID,
		UserID:        account.UserID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
	}, nil
}

func (a *accountRepository) GetAccountByAccountNumber(ctx context.Context, account_number string) (*models.Account, error) {
	account, err := a.Queries.GetAccountByAccountNumber(ctx, account_number)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by account number: %w", err)
	}

	return &models.Account{
		ID:            account.ID,
		UserID:        account.UserID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
	}, nil
}

func (a *accountRepository) CheckIfAccountExists(ctx context.Context, account_number string) (bool, error) {
	account, err := a.Queries.CheckIfAccountNumberExists(ctx, account_number)

	if err != nil {
		return false, fmt.Errorf("failed to get user by account number: %w", err)
	}

	return account, nil
}

func (a *accountRepository) AddBalance(ctx context.Context, amount int64, account_id pgtype.UUID) error {
	err := a.Queries.AddBalance(ctx, sqlc.AddBalanceParams{
		Balance: amount,
		ID:      account_id,
	})

	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	return nil
}

func (a *accountRepository) SubtractBalance(ctx context.Context, amount int64, account_id pgtype.UUID) error {
	err := a.Queries.SubtractBalance(ctx, sqlc.SubtractBalanceParams{
		Balance: amount,
		ID:      account_id,
	})

	if err != nil {
		return fmt.Errorf("failed to subtract balance: %w", err)
	}

	return nil
}

func (a *accountRepository) GetAccountForUpdate(ctx context.Context, account_id pgtype.UUID) (*models.Account, error) {
	account, err := a.Queries.GetAccountForUpdate(ctx, account_id)

	if err != nil {
		return nil, fmt.Errorf("failed to get account for update: %w", err)
	}

	return &models.Account{
		ID:            account.ID,
		UserID:        account.UserID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
		Balance:       account.Balance,
	}, nil
}
