package main

import (
	"fmt"
	"log"

	"example.com/internal/config"
	"example.com/internal/controller"
	"example.com/internal/db"
	"example.com/internal/pkg/jwt"
	"example.com/internal/repository"
	"example.com/internal/router"
	"example.com/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading app config", err)
	}

	s := cfg.Server
	serverAddr := fmt.Sprintf(":%s", s.Port)

	conn, err := db.InitDatabase(&cfg.Database)
	if err != nil {
		fmt.Println("Error creating connection pool:", err)
	}

	jwtConfig := cfg.JWT

	jwt := jwt.NewManager(jwtConfig.Secret, jwtConfig.AccessExpiry, jwtConfig.RefreshExpiry)

	userRepo := repository.NewUserRepository(conn)
	accountRepo := repository.NewAccountRepository(conn)
	authService := service.NewAuthService(userRepo, *jwt, conn, accountRepo)
	userController := controller.NewUserController(authService)

	fmt.Println(conn)

	r := gin.Default()
	router.RegisterUserRoutes(r, userController)
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Error starting Server", err)
	}

}
