package router

import (
	"github.com/gin-gonic/gin"
)

type UserController interface {
	CreateUser(c *gin.Context) error
	GetUsers(c *gin.Context) error
}

func RegisterUserRoutes(r *gin.Engine, controller UserController) {
	users := r.Group("/users")
	users.POST("", func(c *gin.Context) {
		_ = controller.CreateUser(c)
	})
	users.GET("", func(c *gin.Context) {
		_ = controller.GetUsers(c)
	})
}
