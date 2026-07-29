package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Users struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	FullName     string    `json:"full_name"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type JWTClaims struct {
	Full_name string
	Email     string
	UserID    string
	jwt.RegisteredClaims
}
