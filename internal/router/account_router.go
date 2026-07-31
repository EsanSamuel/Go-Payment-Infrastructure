package router

import (
	"example.com/internal/middleware"
	"example.com/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type AccountController interface {
	Transfer(c *gin.Context) error
}

func RegisterAccountRoutes(r *gin.Engine, controller AccountController, jwtManager *jwt.Manager) {
	accounts := r.Group("/accounts")
	accounts.Use(middleware.AuthMiddleware(jwtManager))
	accounts.POST("/transfer", func(c *gin.Context) {
		_ = controller.Transfer(c)
	})
}
