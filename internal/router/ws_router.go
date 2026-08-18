package router

import (
	"example.com/internal/middleware"
	"example.com/internal/pkg/jwt"
	"example.com/internal/ws"

	"github.com/gin-gonic/gin"
)

func RegisterWSRoutes(
	r *gin.Engine,
	hub *ws.Hub,
	jwtManager *jwt.Manager,
) {
	r.GET("/ws", middleware.AuthMiddleware(jwtManager), func(c *gin.Context) {
		hub.HandleWS(c)
	})
}
