package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeAccountService struct {
	transferFn func(ctx context.Context, req models.TransferRequest, userID pgtype.UUID) (*models.Transfer, error)
}

func (f *fakeAccountService) Transfer(ctx context.Context, req models.TransferRequest, userID pgtype.UUID) (*models.Transfer, error) {
	return f.transferFn(ctx, req, userID)
}

func TestTransferHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeAccountService{
		transferFn: func(ctx context.Context, req models.TransferRequest, userID pgtype.UUID) (*models.Transfer, error) {
			return &models.Transfer{Amount: req.Amount}, nil
		},
	}

	controller := NewAccountController(service)
	router := gin.New()
	RegisterAccountRoutes(router, controller)

	payload := []byte(`{
		"from_account_id": "00000000-0000-0000-0000-000000000001",
		"to_account_id": "00000000-0000-0000-0000-000000000002",
		"amount": 50,
		"idempotency_key": "abc-123",
		"method": "POST",
		"path": "/accounts/transfer"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/accounts/transfer", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if _, ok := body["transfer"]; !ok {
		t.Fatalf("expected transfer in response, got %v", body)
	}
}
