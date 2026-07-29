package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Users struct {
	ID                pgtype.UUID        `json:"id"`
	Email             string             `json:"email"`
	PasswordHash      string             `json:"-"`
	FullName          string             `json:"full_name"`
	IsVerified        bool               `json:"is_verified"`
	VerificationToken pgtype.Text        `json:"verification_token"`
	RefreshToken      pgtype.Text        `json:"refresh_token"`
	CreatedAt         pgtype.Timestamptz `json:"created_at"`
	UpdatedAt         pgtype.Timestamptz `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *Users `json:"user"`
}

type JWTClaims struct {
	Full_name string
	Email     string
	UserID    string
	jwt.RegisteredClaims
}
