package controller

import (
	"context"
	"net/http"
	"time"

	"example.com/internal/models"
	"example.com/internal/service"
	"github.com/gin-gonic/gin"
)

type userController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *userController {
	return &userController{
		userService: userService,
	}
}

func (h *userController) CreateUser(c *gin.Context) error {
	var user models.Users
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding payload request", "Details": err.Error()})
		return err
	}

	createdUser, err := h.userService.Create(ctx, user.Email)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating user", "Details": err.Error()})
		return err
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created", "user": createdUser})
	return err
}

func (h *userController) GetUsers(c *gin.Context) error {
	users, err := h.userService.GetUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return err
	}

	c.JSON(http.StatusOK, users)
	return nil
}
