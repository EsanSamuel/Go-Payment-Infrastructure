package controller

import (
	"context"
	"net/http"
	"time"

	"example.com/internal/models"
	"example.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type PayrollController struct {
	payrollService service.PayrollService
	accountService service.AccountService
}

func NewPayrollController(payrollService service.PayrollService, accountService service.AccountService) *PayrollController {
	return &PayrollController{payrollService: payrollService, accountService: accountService}
}

func (pc *PayrollController) CreatePayrollBatch(c *gin.Context) error {
	// Implementation for creating payroll batch
	var req models.PayrollBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payroll batch payload", "details": err.Error()})
		return err
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userIDstr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "company not authenticated"})
		return nil
	}

	var userID pgtype.UUID
	if err := userID.Scan(userIDstr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return err
	}

	companyAccount, err := pc.accountService.GetAccountById(ctx, req.CompanyAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get company account", "details": err.Error()})
		return nil
	}

	if companyAccount == nil || companyAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "company account does not belong to authenticated user"})
		return nil
	}

	batch, err := pc.payrollService.ScheduleBatch(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule payroll batch", "details": err.Error()})
		return err
	}
	c.JSON(http.StatusCreated, batch)
	return nil
}

func (pc *PayrollController) CreatePayroll(c *gin.Context) error {
	var req models.PayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payroll payload", "details": err.Error()})
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	payroll, err := pc.payrollService.CreatePayroll(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payroll item", "details": err.Error()})
		return err
	}
	c.JSON(http.StatusCreated, payroll)
	return nil
}
