package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/internal/config"
	"example.com/internal/controller"
	"example.com/internal/db"
	"example.com/internal/jobs"
	"example.com/internal/pkg/jwt"
	"example.com/internal/repository"
	"example.com/internal/router"
	"example.com/internal/service"
	"example.com/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gocraft/work"
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
		log.Fatal("Error creating connection pool:", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	jwtConfig := cfg.JWT

	jwt := jwt.NewManager(jwtConfig.Secret, jwtConfig.AccessExpiry, jwtConfig.RefreshExpiry)
	h := ws.NewHub()

	// Repository
	userRepo := repository.NewUserRepository(conn)
	accountRepo := repository.NewAccountRepository(conn)
	ledgerRepo := repository.NewLedgerRepository(conn)
	payrollRepo := repository.NewPayrollRepository(conn)
	notificationRepo := repository.NewNotificationRepository(conn)

	// Services
	notificationService := service.NewNotificationService(notificationRepo, h)
	authService := service.NewAuthService(userRepo, *jwt, conn, accountRepo, httpClient)
	accountService := service.NewAccountService(userRepo, conn, accountRepo, ledgerRepo, notificationService)
	payrollService := service.NewPayrollService(payrollRepo)

	// Handlers
	authController := controller.NewAuthController(authService)
	accountController := controller.NewAccountController(accountService)
	userController := controller.NewUserController(authService)
	payrollController := controller.NewPayrollController(payrollService, accountService)

	fmt.Println("Database connection initialized:", conn)

	logger := config.InitLogger()
	fmt.Print(logger)

	var enqueuer = work.NewEnqueuer("payroll", jobs.RedisPool)

	scheduler := jobs.NewSchedule(accountRepo, payrollRepo, conn, enqueuer, accountService, ledgerRepo, logger, notificationService)
	fmt.Println("Scheduler initialized:", scheduler)

	workerPool := work.NewWorkerPool(jobs.Context{}, 10, "payroll", jobs.RedisPool)

	scheduler.RegisterPerodicJobs(workerPool)

	workerPool.Start()
	defer workerPool.Stop()

	r := gin.Default()
	router.RegisterAuthRoutes(r, authController)
	router.RegisterAccountRoutes(r, accountController, jwt)
	router.RegisterUserRoutes(r, userController)
	router.RegisterPayrollRoutes(r, payrollController, jwt)
	router.RegisterWSRoutes(r, h, jwt)
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Error starting Server", err)
	}

}

//migrate -path internal/db/migrations -database "$DATABASE_URL" up
