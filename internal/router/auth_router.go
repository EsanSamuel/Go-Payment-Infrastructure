package router

import "github.com/gin-gonic/gin"

type AuthController interface {
	Register(c *gin.Context) error
	Login(c *gin.Context) error
}

func RegisterAuthRoutes(r *gin.Engine, controller AuthController) {
	auth := r.Group("/auth")
	auth.POST("/register", func(c *gin.Context) {
		_ = controller.Register(c)
	})
	auth.POST("/login", func(c *gin.Context) {
		_ = controller.Login(c)
	})
}
