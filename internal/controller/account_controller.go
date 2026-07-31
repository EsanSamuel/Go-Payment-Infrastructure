package controller

import (
	"context"
	"io"
	"net/http"
	"time"

	"example.com/internal/models"
	"example.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type AccountController interface {
	Transfer(c *gin.Context) error
}

type accountController struct {
	accountService service.AccountService
}

func NewAccountController(accountService service.AccountService) AccountController {
	return &accountController{accountService: accountService}
}

func (h *accountController) Transfer(c *gin.Context) error {
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer payload", "details": err.Error()})
		return err
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return nil
	}

	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return err
	}

	req.Method = c.Request.Method
	req.Path = c.FullPath()
	body, _ := io.ReadAll(c.Request.Body)
	req.Body = body

	transfer, err := h.accountService.Transfer(ctx, req, userUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"transfer": transfer})
	return nil
}
