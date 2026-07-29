package jwt

import (
	"errors"
	"log"
	"time"

	"example.com/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secret             string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewManager(secret string, accessTokenExpiry time.Duration, refreshTokenExpiry time.Duration) *Manager {
	return &Manager{
		secret:             secret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (m *Manager) GetAccessToken(full_name, user_id, email string) (string, error) {
	claims := &models.JWTClaims{
		Full_name: full_name,
		Email:     email,
		UserID:    user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "FinTech",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		log.Println("Error creating access token", err.Error())
	}

	return tokenString, nil
}

func (m *Manager) GenerateRefreshToken() string {
	return uuid.New().String()
}

func (m *Manager) GetRefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}

func (m *Manager) VerifyAccessToken(tokenString string) (*models.JWTClaims, error) {
	claims := &models.JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}
	return claims, nil
}
