package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"example.com/internal/models"
	"example.com/internal/pkg/jwt"
	"example.com/internal/pkg/token"
	"example.com/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email string, full_name string, password string) (*models.AuthResponse, error)
	Login(ctx context.Context, email, password string) (*models.AuthResponse, error)
	GetUsers(ctx context.Context) ([]models.Users, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwt      jwt.Manager
}

func NewAuthService(userRepo repository.UserRepository, jwt jwt.Manager) *authService {
	return &authService{
		userRepo: userRepo,
		jwt:      jwt,
	}
}

func (u *authService) RegisterUser(ctx context.Context, email string, full_name string, password string) (*models.AuthResponse, error) {
	emailExists, err := u.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		log.Println("Email already exists!")
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password!")
	}
	user, err := u.userRepo.Create(ctx, email, full_name, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	verificationToken, err := token.GenerateVerificationToken()
	if err != nil {
		log.Println("Error generating token", err)
	}

	err = u.userRepo.InsertVerificationToken(ctx, pgtype.Text{String: verificationToken, Valid: true}, user.ID)
	if err != nil {
		log.Println("Error inserting verification token to db", err)
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
