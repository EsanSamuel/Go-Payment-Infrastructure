package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"example.com/internal/models"
	"example.com/internal/paystack"
	"example.com/internal/pkg/jwt"
	"example.com/internal/pkg/token"
	"example.com/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email string, full_name string, password string) (*models.AuthResponse, error)
	Login(ctx context.Context, email, password string) (*models.AuthResponse, error)
	GetUsers(ctx context.Context) ([]models.Users, error)
}

type authService struct {
	pool        *pgxpool.Pool
	userRepo    repository.UserRepository
	accountRepo repository.AccountRepository
	jwt         jwt.Manager
	httpClient  *http.Client
}

func NewAuthService(userRepo repository.UserRepository, jwt jwt.Manager, pool *pgxpool.Pool, accountRepo repository.AccountRepository, httpClient *http.Client) AuthService {
	return &authService{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		jwt:         jwt,
		pool:        pool,
		httpClient:  httpClient,
	}
}

func (s *authService) RegisterUser(
	ctx context.Context,
	email, fullName, password string,
) (*models.AuthResponse, error) {

	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	verificationToken, err := token.GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	refreshToken := s.jwt.GenerateRefreshToken()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	userRepo := s.userRepo.WithTx(tx)
	accountRepo := s.accountRepo.WithTx(tx)

	user, err := userRepo.Create(ctx, email, fullName, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	customerReq := paystack.CustomerRequest{
		Firstname: user.FullName,
		Lastname:  "",
		Phone:     "+2348064515733",
		Email:     user.Email,
	}

	customer, err := paystack.CreateCustomer(customerReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to create customer %w:", err)
	}

	dedicatedAcct, err := paystack.CreateDedicatedAccount(customer.Data.CustomerCode)
	if err != nil {
		return nil, fmt.Errorf("Failed to create customer dedicated account %w:", err)
	}

	fmt.Printf("User dedicated Account %+v\n", dedicatedAcct)
	accountNumber := dedicatedAcct.Data.AccountNumber

	_, err = accountRepo.CreateAccount(ctx, accountNumber, "NGN", user.ID)
	if err != nil {
		return nil, err
	}

	err = userRepo.InsertVerificationToken(
		ctx,
		pgtype.Text{
			String: verificationToken,
			Valid:  true,
		},
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	err = userRepo.InsertRefreshToken(
		ctx,
		pgtype.Text{
			String: refreshToken,
			Valid:  true,
		},
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GetAccessToken(
		user.FullName,
		user.Email,
		user.ID,
	)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (u *authService) Login(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("Password do not match!")
	}

	accessToken, err := u.jwt.GetAccessToken(user.FullName, user.Email, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken := u.jwt.GenerateRefreshToken()

	err = u.userRepo.InsertRefreshToken(ctx, pgtype.Text{String: refreshToken, Valid: true}, user.ID)
	if err != nil {
		log.Println("Error inserting refresh token to db", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &models.Users{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
	}, nil
}

func (u *authService) GetUsers(ctx context.Context) ([]models.Users, error) {

	users, err := u.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
