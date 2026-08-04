package service

import (
	"context"
	"fmt"

	"example.com/internal/models"
	"example.com/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type PayrollService interface {
	ScheduleBatch(ctx context.Context, req models.PayrollBatchRequest) (*models.PayrollBatch, error)
	CreatePayroll(ctx context.Context, req models.PayrollRequest) (*models.Payroll, error)
	GetPayrollBatchById(ctx context.Context, batchID pgtype.UUID, companyAccountID pgtype.UUID) (*models.PayrollBatch, error)
}

type payrollService struct {
	payrollRepo repository.PayrollRepository
}

func NewPayrollService(payrollRepo repository.PayrollRepository) PayrollService {
	return &payrollService{payrollRepo: payrollRepo}
}

func (s *payrollService) ScheduleBatch(ctx context.Context, req models.PayrollBatchRequest) (*models.PayrollBatch, error) {
	batch, err := s.payrollRepo.CreatePayrollBatch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule payroll batch: %w", err)
	}
	return batch, nil
}

func (s *payrollService) CreatePayroll(ctx context.Context, req models.PayrollRequest) (*models.Payroll, error) {
	payroll, err := s.payrollRepo.CreatePayroll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payroll item: %w", err)
	}
	return payroll, nil
}

func (s *payrollService) GetPayrollBatchById(ctx context.Context, batchID pgtype.UUID, companyAccountID pgtype.UUID) (*models.PayrollBatch, error) {
	batch, err := s.payrollRepo.GetPayrollBatchById(ctx, batchID, companyAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to load payroll batch: %w", err)
	}
	return batch, nil
}
