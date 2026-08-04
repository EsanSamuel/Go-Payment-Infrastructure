package router

import (
	"example.com/internal/middleware"
	"example.com/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type PayrollController interface {
	CreatePayrollBatch(c *gin.Context) error
	CreatePayroll(c *gin.Context) error
}

func RegisterPayrollRoutes(r *gin.Engine, controller PayrollController, jwtManager *jwt.Manager) {
	payrolls := r.Group("/payrolls")
	payrolls.Use(middleware.AuthMiddleware(jwtManager))
	payrolls.POST("/items", func(c *gin.Context) {
		_ = controller.CreatePayroll(c)
	})
	payrolls.POST("/batches", func(c *gin.Context) {
		_ = controller.CreatePayrollBatch(c)
	})
}
