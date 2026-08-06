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

type PayrollRepository interface {
	WithTx(tx pgx.Tx) PayrollRepository
	CreatePayrollBatch(ctx context.Context, req models.PayrollBatchRequest) (*models.PayrollBatch, error)
	CreatePayroll(ctx context.Context, req models.PayrollRequest) (*models.Payroll, error)
	GetPayrollBatchById(ctx context.Context, batchID pgtype.UUID, companyAccountID pgtype.UUID) (*models.PayrollBatch, error)
	GetDuePayrollBatches(ctx context.Context) ([]models.PayrollBatch, error)
	UpdateBatchStatusToProcessing(ctx context.Context, batchId pgtype.UUID) error
	GetPendingPayrollItemByBatchId(ctx context.Context, batchID pgtype.UUID) ([]models.Payroll, error)
	GetPayrollItemById(ctx context.Context, itemID pgtype.UUID) (*models.Payroll, error)
	UpdatePayrollItemToFailed(ctx context.Context, itemID pgtype.UUID) error
	UpdatePayrollItemToCompleted(ctx context.Context, itemID pgtype.UUID) error
	GetPayrollBatchForUpdate(ctx context.Context, batchID pgtype.UUID) (*models.PayrollBatch, error)
	CheckBatchCompletion(ctx context.Context, batchID pgtype.UUID) (bool, bool, error)
	FinalizeBatchStatus(ctx context.Context, batchId pgtype.UUID, status sqlc.PayrollBatchStatus, scheduleDate pgtype.Timestamptz) error
}

type payrollRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewPayrollRepository(pool *pgxpool.Pool) PayrollRepository {
	return &payrollRepository{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *payrollRepository) WithTx(tx pgx.Tx) PayrollRepository {
	return &payrollRepository{
		Queries: r.Queries.WithTx(tx),
		pool:    r.pool,
	}
}

func (r *payrollRepository) CreatePayrollBatch(ctx context.Context, req models.PayrollBatchRequest) (*models.PayrollBatch, error) {
	batch, err := r.Queries.CreatePayrollBatch(ctx, sqlc.CreatePayrollBatchParams{
		CompanyAccountID: req.CompanyAccountID,
		ScheduleDate:     req.ScheduleDate,
		Status:           req.Status,
		BatchName:        req.BatchName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payroll batch: %w", err)
	}

	return &models.PayrollBatch{
		ID:               batch.ID,
		CompanyAccountID: batch.CompanyAccountID,
		ScheduleDate:     batch.ScheduleDate,
		Status:           batch.Status,
		CreatedAt:        batch.CreatedAt,
		BatchName:        batch.BatchName,
	}, nil
}

func (r *payrollRepository) CreatePayroll(ctx context.Context, req models.PayrollRequest) (*models.Payroll, error) {
	payroll, err := r.Queries.CreatePayroll(ctx, sqlc.CreatePayrollParams{
		BatchID:           req.BatchID,
		EmployeeAccountID: req.EmployeeAccountID,
		Amount:            req.Amount,
		Status:            req.Status,
		Description:       req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payroll record: %w", err)
	}

	return &models.Payroll{
		ID:                payroll.ID,
		BatchID:           payroll.BatchID,
		EmployeeAccountID: payroll.EmployeeAccountID,
		Amount:            payroll.Amount,
		Status:            payroll.Status,
		CreatedAt:         payroll.CreatedAt,
		Description:       payroll.Description,
	}, nil
}

func (r *payrollRepository) GetPayrollBatchById(ctx context.Context, batchID pgtype.UUID, companyAccountID pgtype.UUID) (*models.PayrollBatch, error) {
	batch, err := r.Queries.GetPayrollBatchById(ctx, sqlc.GetPayrollBatchByIdParams{
		ID:               batchID,
		CompanyAccountID: companyAccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll batch: %w", err)
	}

	return &models.PayrollBatch{
		ID:               batch.ID,
		CompanyAccountID: batch.CompanyAccountID,
		ScheduleDate:     batch.ScheduleDate,
		Status:           batch.Status,
		CreatedAt:        batch.CreatedAt,
	}, nil
}

func (r *payrollRepository) GetDuePayrollBatches(ctx context.Context) ([]models.PayrollBatch, error) {
	batches, err := r.Queries.GetDuePayrollBatches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get due payroll batches: %w", err)
	}

	result := make([]models.PayrollBatch, len(batches))
	for i := range batches {
		result[i] = models.PayrollBatch{
			ID:               batches[i].ID,
			CompanyAccountID: batches[i].CompanyAccountID,
			ScheduleDate:     batches[i].ScheduleDate,
			Status:           batches[i].Status,
			CreatedAt:        batches[i].CreatedAt,
		}
	}
	return result, nil
}

func (r *payrollRepository) UpdateBatchStatusToProcessing(ctx context.Context, batchId pgtype.UUID) error {
	_, err := r.Queries.UpdateBatchToProcessing(ctx, batchId)
	if err != nil {
		return fmt.Errorf("failed to get update payroll batche status to processing: %w", err)
	}
	return nil
}

func (r *payrollRepository) GetPendingPayrollItemByBatchId(ctx context.Context, batchID pgtype.UUID) ([]models.Payroll, error) {
	items, err := r.Queries.GetPendingBatchPayrollItem(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payroll batch items: %w", err)
	}

	result := make([]models.Payroll, len(items))
	for i := range items {
		result[i] = models.Payroll{
			ID:                items[i].ID,
			EmployeeAccountID: items[i].EmployeeAccountID,
			BatchID:           items[i].BatchID,
			Status:            items[i].Status,
			Amount:            items[i].Amount,
			CreatedAt:         items[i].CreatedAt,
		}
	}
	return result, nil
}

func (r *payrollRepository) GetPayrollItemById(ctx context.Context, itemID pgtype.UUID) (*models.Payroll, error) {
	item, err := r.Queries.GetPayrollItemByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll item: %w", err)
	}
	return &models.Payroll{
		ID:                item.ID,
		EmployeeAccountID: item.EmployeeAccountID,
		BatchID:           item.BatchID,
		Status:            item.Status,
		Amount:            item.Amount,
		CreatedAt:         item.CreatedAt,
		CompanyAccountID:  item.CompanyAccountID,
		CompanyUserID:     item.UserID,
	}, nil
}

func (r *payrollRepository) UpdatePayrollItemToFailed(ctx context.Context, itemID pgtype.UUID) error {
	_, err := r.Queries.UpdatePayrollToFailed(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to update payroll item status to failed: %w", err)
	}

	return nil
}

func (r *payrollRepository) UpdatePayrollItemToCompleted(ctx context.Context, itemID pgtype.UUID) error {
	_, err := r.Queries.UpdatePayrollToCompleted(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to update payroll item status to completed: %w", err)
	}

	return nil
}

func (r *payrollRepository) CheckBatchCompletion(ctx context.Context, batchID pgtype.UUID) (bool, bool, error) {
	row, err := r.Queries.CheckBatchCompleted(ctx, batchID)
	if err != nil {
		return false, false, fmt.Errorf("failed to check batch completion: %w", err)
	}
	resolved := row.Completed + row.Failed
	allDone := resolved >= row.Total
	anyFailed := row.Failed > 0

	fmt.Printf(
		"Resolved: %d, AllDone: %v, AnyFailed: %v\n",
		resolved,
		allDone,
		anyFailed,
	)
	return allDone, anyFailed, nil
}

func (r *payrollRepository) GetPayrollBatchForUpdate(ctx context.Context, batchID pgtype.UUID) (*models.PayrollBatch, error) {
	batch, err := r.Queries.GetBatchForUpdate(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll batch for update: %w", err)
	}

	return &models.PayrollBatch{
		ID:               batch.ID,
		CompanyAccountID: batch.CompanyAccountID,
		ScheduleDate:     batch.ScheduleDate,
		Status:           batch.Status,
		CreatedAt:        batch.CreatedAt,
	}, nil
}

func (r *payrollRepository) FinalizeBatchStatus(ctx context.Context, batchId pgtype.UUID, status sqlc.PayrollBatchStatus, scheduleDate pgtype.Timestamptz) error {
	_, err := r.Queries.FinalizeBatch(ctx, sqlc.FinalizeBatchParams{
		Status:       status,
		ScheduleDate: scheduleDate,
		ID:           batchId,
	})
	if err != nil {
		return fmt.Errorf("failed to finalize batch status: %w", err)
	}
	return nil
}
