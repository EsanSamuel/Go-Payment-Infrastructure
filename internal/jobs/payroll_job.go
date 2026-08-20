package jobs

import (
	"context"
	"fmt"
	"time"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"example.com/internal/repository"
	logClient "github.com/EsanSamuel/sensory/LogClient"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferService interface {
	Transfer(ctx context.Context, req models.TransferRequest, UserID pgtype.UUID) (*models.Transfer, error)
}

type Context struct{}

var RedisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	},
}

type Scheduler struct {
	accountRepo    repository.AccountRepository
	payrollRepo    repository.PayrollRepository
	accountService TransferService
	ledgerRepo     repository.LedgerRepository
	*sqlc.Queries
	pool     *pgxpool.Pool
	enqueuer *work.Enqueuer
	logger   *logClient.Client
}

func NewSchedule(accountRepo repository.AccountRepository, payrollRepo repository.PayrollRepository, pool *pgxpool.Pool, enqueuer *work.Enqueuer, accountService TransferService, ledgerRepo repository.LedgerRepository, logger *logClient.Client) *Scheduler {
	return &Scheduler{
		accountRepo:    accountRepo,
		payrollRepo:    payrollRepo,
		Queries:        sqlc.New(pool),
		pool:           pool,
		enqueuer:       enqueuer,
		accountService: accountService,
		ledgerRepo:     ledgerRepo,
		logger:         logger,
	}
}

func (s *Scheduler) RegisterPerodicJobs(pool *work.WorkerPool) {
	fmt.Println("REGISTERING PERIODIC JOBS")

	pool.PeriodicallyEnqueue("0 */5 * * * *", "check_due_payroll_batches") // Schedule the job to run every 5 minutes
	pool.Job("check_due_payroll_batches", s.CheckDuePayrollBatches)
	pool.Job("process_payroll_item", s.ProcessPayrollItem)
}

func (s *Scheduler) CheckDuePayrollBatches(job *work.Job) error {
	fmt.Println("===== CRON FIRED =====", time.Now())
	s.logger.INFO(fmt.Sprintf("===== CRON FIRED at %v =====", time.Now()))

	ctx := context.Background()

	fmt.Println("Starting transaction")

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	fmt.Println("Fetching due batches")
	s.logger.INFO("Fetching due batches")

	payrollRepo := s.payrollRepo.WithTx(tx)

	dueBatches, err := payrollRepo.GetDuePayrollBatches(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Due batches found:", len(dueBatches))
	s.logger.INFO(fmt.Sprint("Due batches found:", len(dueBatches)))

	for i, batch := range dueBatches {
		fmt.Println("Processing batch:", i, batch.ID)
		s.logger.INFO(fmt.Sprint("Processing batch:", i, batch.ID))

		fmt.Println("Updating status")
		err := payrollRepo.UpdateBatchStatusToProcessing(ctx, batch.ID)
		if err != nil {
			return err
		}

		fmt.Println("Getting items")
		items, err := payrollRepo.GetPendingPayrollItemByBatchId(ctx, batch.ID)

		if err != nil {
			return err
		}

		fmt.Println("Items:", len(items))
		if len(items) == 0 {
			fmt.Println("No payroll items found, reverting batch")
			s.logger.INFO("No payroll items found, reverting batch")

			err := payrollRepo.FinalizeBatchStatus(
				ctx,
				batch.ID,
				"PENDING",
				batch.ScheduleDate,
			)

			if err != nil {
				return err
			}

			continue
		}

		for _, item := range items {
			fmt.Println("Processing item id:", item.ID)
			_, err = s.enqueuer.Enqueue("process_payroll_item", work.Q{
				"payroll_item_id": item.ID.String(),
				"batch_timestamp": batch.ScheduleDate.Time.Format(time.RFC3339),
			})
			if err != nil {
				s.logger.ERROR(fmt.Sprintf("failed to enqueue item %s: %w", item.ID, err))
				return fmt.Errorf("failed to enqueue item %s: %w", item.ID, err)

			}
		}
	}

	fmt.Println("Committing transaction")

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Println("Cron completed")
	s.logger.INFO("Cron completed")
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
		s.logger.ERROR(fmt.Sprintf("failed to fetch payroll item %s: %w", itemID, err))
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

	description := fmt.Sprintf("Schedule payment to %s for %s", receiverAccount.AccountNumber, item.Description.String)

	transferReq := models.TransferRequest{
		FromAccountID:   item.CompanyAccountID,
		ToAccountID:     item.EmployeeAccountID,
		Amount:          item.Amount,
		IdempotencyKey:  key,
		ToAccountNumber: receiverAccount.AccountNumber,
		Narration:       pgtype.Text{String: description, Valid: true},
	}

	_, transferErr := s.accountService.Transfer(ctx, transferReq, item.CompanyUserID)
	if transferErr != nil {
		s.logger.ERROR(fmt.Sprintf("Transfer failed for payroll item %s: %v\n", item.ID, transferErr))
		fmt.Printf("Transfer failed for payroll item %s: %v\n", item.ID, transferErr)
		s.payrollRepo.UpdatePayrollItemToFailed(ctx, item.ID)
		return nil
	}

	if err := s.payrollRepo.UpdatePayrollItemToCompleted(ctx, item.ID); err != nil {
		return err
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

	fmt.Println("========== FinalizeBatchIfDone ==========")
	fmt.Println("Batch ID:", batchID)
	s.logger.INFO("========== FinalizeBatchIfDone ==========")
	s.logger.INFO(fmt.Sprintf("Batch ID:", batchID))

	payrollRepo := s.payrollRepo.WithTx(tx)

	batch, err := payrollRepo.GetPayrollBatchForUpdate(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to lock batch: %w", err)
	}

	if batch.Status != "PROCESSING" {
		return nil
	}

	allDone, anyFailed, err := payrollRepo.CheckBatchCompletion(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to check batch completion: %w", err)
	}

	if !allDone {
		return nil
	}

	status := "COMPLETED"
	if anyFailed {
		status = "PARTIAL"
	}

	fmt.Println("Final batch result:", status)

	nextRun := nextMonthEndOrSameDay(batch.ScheduleDate.Time)
	fmt.Println("Next run:", nextRun)

	var nextRunPg pgtype.Timestamptz
	if err := nextRunPg.Scan(nextRun); err != nil {
		fmt.Println("Failed to scan nextRun:", err)
		return fmt.Errorf("failed to convert next run time: %w", err)
	}

	err = payrollRepo.FinalizeBatchStatus(ctx, batch.ID, "PENDING", nextRunPg)
	if err != nil {
		return fmt.Errorf("failed to finalize batch: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Batch finalized successfully!")
	fmt.Println("=========================================")

	return nil
}
