package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"example.com/internal/models"
	"example.com/internal/pkg/hash"
	"example.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountService interface {
	Transfer(ctx context.Context, req models.TransferRequest, UserID pgtype.UUID) (*models.Transfer, error)
}

type accountService struct {
	pool        *pgxpool.Pool
	userRepo    repository.UserRepository
	accountRepo repository.AccountRepository
}

func NewAccountService(userRepo repository.UserRepository, pool *pgxpool.Pool, accountRepo repository.AccountRepository) AccountService {
	return &accountService{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		pool:        pool,
	}
}

func (s *accountService) Transfer(ctx context.Context, req models.TransferRequest, UserID pgtype.UUID) (*models.Transfer, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	accountRepo := s.accountRepo.WithTx(tx)

	request_hash := hash.HashRequest(req.Path, req.Method, req.Body)

	claimed, err := accountRepo.CreateIdempotency(ctx, models.IdempotentRequest{IdempotencyKey: req.IdempotencyKey, UserID: UserID, RequestHash: request_hash})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, err := accountRepo.GetIdempotency(ctx, req.IdempotencyKey, UserID)
			if err != nil {
				return nil, fmt.Errorf("Failed to get idempotency %w", err)
			}
			if existing.RequestHash != request_hash {
				return nil, fmt.Errorf("idempotency key reused with different payload")
			}
			if existing.CompletedAt.Valid {
				var cached models.Transfer
				if err := json.Unmarshal(existing.ResponseBody, &cached); err != nil {
					return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
				}
				return &cached, nil
			}
			return nil, fmt.Errorf("request already in progress")
		}
		return nil, fmt.Errorf("failed to claim idempotency key: %w", err)
	}
	_ = claimed

	receiverAccount, err := accountRepo.GetAccountByAccountNumber(ctx, req.ToAccountNumber)
	if err != nil {
		return nil, fmt.Errorf("receiver account not found: %w", err)
	}

	toAccountID := receiverAccount.ID

	if bytes.Equal(req.FromAccountID.Bytes[:], toAccountID.Bytes[:]) {
		return nil, fmt.Errorf("cannot transfer to the same account")
	}

	firstID, secondID := req.FromAccountID, toAccountID
	if bytes.Compare(firstID.Bytes[:], secondID.Bytes[:]) > 0 {
		firstID, secondID = secondID, firstID
	}

	firstAccount, err := accountRepo.GetAccountForUpdate(ctx, firstID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock account: %w", err)
	}
	secondAccount, err := accountRepo.GetAccountForUpdate(ctx, secondID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock account: %w", err)
	}

	var sender, receiver *models.Account
	if firstAccount.ID == req.FromAccountID {
		sender, receiver = firstAccount, secondAccount
	} else {
		sender, receiver = secondAccount, firstAccount
	}

	if sender.Balance < req.Amount {
		return nil, fmt.Errorf("Insufficient Balance")
	}

	err = accountRepo.SubtractBalance(ctx, req.Amount, sender.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to subtract balance from sender's account %w", err)
	}
	err = accountRepo.AddBalance(ctx, req.Amount, receiver.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to ass balance to receiver's account %w", err)
	}

	// Resolve the public account number to the internal UUID
	// before persisting the transfer.
	req.ToAccountID = toAccountID
	_, err = accountRepo.CreateTransfer(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create transfer %w", err)
	}

	transfer := &models.Transfer{
		FromAccountID: sender.ID,
		ToAccountID:   receiver.ID,
		Amount:        req.Amount,
	}

	responseBody, err := json.Marshal(transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	if err := accountRepo.CompleteIdempotency(ctx, req.IdempotencyKey, UserID, 200, responseBody); err != nil {
		return nil, fmt.Errorf("failed to complete idempotency key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return transfer, nil
}
