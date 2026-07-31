package controller

import (
	"context"
	"net/http"
	"time"

	"example.com/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthController interface {
	Register(c *gin.Context) error
	Login(c *gin.Context) error
}

type authController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) AuthController {
	return &authController{authService: authService}
}

func (h *authController) Register(c *gin.Context) error {
	var payload struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid register payload", "details": err.Error()})
		return err
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response, err := h.authService.RegisterUser(ctx, payload.Email, payload.FullName, payload.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusCreated, response)
	return nil
}

func (h *authController) Login(c *gin.Context) error {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload", "details": err.Error()})
		return err
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response, err := h.authService.Login(ctx, payload.Email, payload.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, response)
	return nil
}
