package jobs

import (
	"context"
	"fmt"
	"time"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"example.com/internal/repository"
	"example.com/internal/service"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Scheduler struct {
	accountRepo    repository.AccountRepository
	payrollRepo    repository.PayrollRepository
	accountService service.AccountService
	*sqlc.Queries
	pool     *pgxpool.Pool
	enqueuer *work.Enqueuer
}

var RedisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	},
}

type Context struct{}

func NewSchedule(accountRepo repository.AccountRepository, payrollRepo repository.PayrollRepository, pool *pgxpool.Pool, enqueuer *work.Enqueuer, accountService service.AccountService) *Scheduler {
	return &Scheduler{
		accountRepo:    accountRepo,
		payrollRepo:    payrollRepo,
		Queries:        sqlc.New(pool),
		pool:           pool,
		enqueuer:       enqueuer,
		accountService: accountService,
	}
}

func (s *Scheduler) RegisterPerodicJobs(pool *work.WorkerPool) {
	pool.PeriodicallyEnqueue("*/5 * * * *", "check_due_payroll_batches") // Schedule the job to run every 5 minutes
	pool.Job("check_due_payroll_batches", s.CheckDuePayrollBatches)
	pool.Job("process_payroll_item", s.ProcessPayrollItem)
}

func (s *Scheduler) CheckDuePayrollBatches(job *work.Job) error {
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	payrollRepo := s.payrollRepo.WithTx(tx)
	dueBatches, err := payrollRepo.GetDuePayrollBatches(ctx)
	if err != nil {
		return err
	}
	for i, batch := range dueBatches {
		fmt.Println("Processing batch id:", i)
		err := payrollRepo.UpdateBatchStatusToProcessing(ctx, batch.ID)
		if err != nil {
			return err
		}

		items, err := payrollRepo.GetPendingPayrollItemByBatchId(ctx, batch.ID)
		fmt.Println("items:", items)
		if err != nil {
			return fmt.Errorf("failed to fetch items for batch %s: %w", batch.ID, err)
		}

		for _, item := range items {
			fmt.Println("Processing item id:", item.ID)
			_, err = s.enqueuer.Enqueue("process_payroll_item", work.Q{
				"payroll_item_id": item.ID.String(),
				"batch_timestamp": batch.ScheduleDate.Time.Format(time.RFC3339),
			})
			if err != nil {
				return fmt.Errorf("failed to enqueue item %s: %w", item.ID, err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)

	}

	return nil
}

func (s *Scheduler) ProcessPayrollItem(job *work.Job) error {
	ctx := context.Background()
	itemID := job.ArgString("payroll_item_id")
	batch_timestamp := job.ArgString("batch_timestamp")

	var itemUUID pgtype.UUID
	if err := itemUUID.Scan(itemID); err != nil {
		return err
	}

	item, err := s.payrollRepo.GetPayrollItemById(ctx, itemUUID)
	if err != nil {
		return fmt.Errorf("failed to fetch payroll item %s: %w", itemID, err)
	}

	// Already handled — don't reprocess on retry
	if item.Status == "COMPLETED" || item.Status == "FAILED" {
		return nil
	}

	receiverAccount, err := s.accountRepo.GetAccountById(ctx, item.EmployeeAccountID)

	key := fmt.Sprintf("%s:%s",
		item.ID.String(),
		batch_timestamp,
	)

	transferReq := models.TransferRequest{
		FromAccountID:   item.CompanyAccountID,
		ToAccountID:     item.EmployeeAccountID,
		Amount:          item.Amount,
		IdempotencyKey:  key,
		ToAccountNumber: receiverAccount.AccountNumber,
	}

	_, transferErr := s.accountService.Transfer(ctx, transferReq, item.CompanyUserID)
	if transferErr != nil {
		fmt.Printf("Transfer failed for payroll item %s: %v\n", item.ID, transferErr)
		s.payrollRepo.UpdatePayrollItemToFailed(ctx, item.ID)
	} else {
		s.payrollRepo.UpdatePayrollItemToCompleted(ctx, item.ID)
	}

	if err := s.FinalizeBatchIfDone(ctx, item.BatchID); err != nil {
		return fmt.Errorf("failed to finalize batch %s: %w", item.BatchID, err)
	}
	return nil
}

func nextMonthEndOrSameDay(current time.Time) time.Time {
	year, month, day := current.Date()
	nextMonth := time.Date(year, month+1, day, current.Hour(), current.Minute(), 0, 0, current.Location())
	// handle month-end overflow: if day doesn't exist in next month (e.g. 31st),
	// time.Date normalizes it forward — detect and clamp to last day of that month instead
	if nextMonth.Day() != day {
		nextMonth = time.Date(year, month+2, 0, current.Hour(), current.Minute(), 0, 0, current.Location())
	}
	return nextMonth
}

func (s *Scheduler) FinalizeBatchIfDone(ctx context.Context, batchID pgtype.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	fmt.Println("Finalizing batch id:", batchID)

	payrollRepo := s.payrollRepo.WithTx(tx)

	batch, err := payrollRepo.GetPayrollBatchForUpdate(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to lock batch: %w", err)
	}

	if batch.Status != "PROCESSING" {
		// already finalized by another concurrent call — nothing to do
		fmt.Println("Batch not in processing state, skipping finalization for batch id:", batchID)
		return nil
	}

	allDone, anyFailed, err := payrollRepo.CheckBatchCompletion(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to check batch completion: %w", err)
	}

	if !allDone {
		fmt.Println("Batch not completed yet, skipping finalization for batch id:", batchID)
		return nil // still waiting on other items
	}

	status := "COMPLETED"
	if anyFailed {
		status = "PARTIAL"
	}
	_ = status

	nextRun := nextMonthEndOrSameDay(batch.ScheduleDate.Time)

	var nextRunPg pgtype.Timestamptz
	if err := nextRunPg.Scan(nextRun); err != nil {
		return fmt.Errorf("failed to convert next run time: %w", err)
	}

	err = payrollRepo.FinalizeBatchStatus(ctx, batch.ID, "PENDING", nextRunPg)
	if err != nil {
		return fmt.Errorf("failed to finalize batch: %w", err)
	}

	return tx.Commit(ctx)
}
