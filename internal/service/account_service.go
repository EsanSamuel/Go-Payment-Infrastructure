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
	GetAccountById(ctx context.Context, id pgtype.UUID) (*models.Account, error)
}

type accountService struct {
	pool                *pgxpool.Pool
	userRepo            repository.UserRepository
	accountRepo         repository.AccountRepository
	ledgerRepo          repository.LedgerRepository
	notificationService NotificationService
}

func NewAccountService(userRepo repository.UserRepository, pool *pgxpool.Pool, accountRepo repository.AccountRepository, ledgerRepo repository.LedgerRepository, notificationService NotificationService) AccountService {
	return &accountService{
		userRepo:            userRepo,
		accountRepo:         accountRepo,
		ledgerRepo:          ledgerRepo,
		pool:                pool,
		notificationService: notificationService,
	}
}

func (s *accountService) Transfer(ctx context.Context, req models.TransferRequest, UserID pgtype.UUID) (*models.Transfer, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	accountRepo := s.accountRepo.WithTx(tx)
	ledgerRepo := s.ledgerRepo.WithTx(tx)

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
		return nil, fmt.Errorf("Failed to add balance to receiver's account %w", err)
	}

	// Resolve the public account number to the internal UUID
	// before persisting the transfer.
	req.ToAccountID = toAccountID
	transcation, err := accountRepo.CreateTransfer(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create transfer %w", err)
	}

	// Sender Entry Ledger
	SenderEntry := models.Entries{
		AccountID:     sender.ID,
		Amount:        req.Amount,
		EntryType:     "DEBIT",
		TransactionID: transcation.ID,
	}
	_, err = ledgerRepo.CreateEntry(ctx, SenderEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender entry: %w", err)
	}

	// Receiver Entry Ledger
	ReceiverEntry := models.Entries{
		AccountID:     receiver.ID,
		Amount:        req.Amount,
		EntryType:     "CREDIT",
		TransactionID: transcation.ID,
	}
	_, err = ledgerRepo.CreateEntry(ctx, ReceiverEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to create receiver entry: %w", err)
	}

	transfer := &models.Transfer{
		ID:            transcation.ID,
		FromAccountID: sender.ID,
		ToAccountID:   receiver.ID,
		Amount:        transcation.Amount,
		Description:   transcation.Description,
		Reference:     transcation.Reference,
	}

	senderAccount, err := accountRepo.GetAccountForUpdate(ctx, req.FromAccountID)
	sender, err = accountRepo.GetAccountByAccountNumber(ctx, senderAccount.AccountNumber)

	notificationReq := models.CreateNotificationRequest{
		UserID:  receiver.UserID,
		Type:    "transfer_received",
		Title:   "Transfer Received",
		Message: "You received a transfer from" + sender.Fullname,
	}

	s.notificationService.CreateNotification(ctx, notificationReq)

	sender_notificationReq := models.CreateNotificationRequest{
		UserID:  sender.UserID,
		Type:    "transfer_sent",
		Title:   "Transfer Sent",
		Message: "You sent a transfer to" + receiverAccount.Fullname,
	}

	s.notificationService.CreateNotification(ctx, sender_notificationReq)

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

func (s *accountService) GetAccountById(ctx context.Context, id pgtype.UUID) (*models.Account, error) {
	account, err := s.accountRepo.GetAccountById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return account, nil
}
