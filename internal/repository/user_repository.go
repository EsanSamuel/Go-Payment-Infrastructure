package repository

import (
	"context"
	"fmt"
	"log"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, email string, full_name string, password string) (*models.Users, error)
	GetUsers(ctx context.Context) ([]models.Users, error)
	GetUserByEmail(ctx context.Context, email string) (*models.Users, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	InsertVerificationToken(ctx context.Context, token pgtype.Text, user_id pgtype.UUID) error
	InsertRefreshToken(ctx context.Context, token pgtype.Text, user_id pgtype.UUID) error
}

type userRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (u *userRepository) Create(ctx context.Context, email string, full_name string, password string) (*models.Users, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:     full_name,
		Email:        email,
		PasswordHash: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Users{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (u *userRepository) GetUsers(ctx context.Context) ([]models.Users, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	users, err := qtx.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	result := make([]models.Users, len(users))

	for i, user := range users {
		result[i] = models.Users{
			ID:    user.ID,
			Email: user.Email,
		}
	}

	return result, nil
}

func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.Users, error) {
	user, err := u.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &models.Users{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
	}, nil
}

func (u *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := u.Queries.EmailExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

func (u *userRepository) InsertVerificationToken(ctx context.Context, token pgtype.Text, user_id pgtype.UUID) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	user, err := qtx.InsertVerificationToken(ctx, sqlc.InsertVerificationTokenParams{
		VerificationToken: token,
		ID:                user_id,
	})
	log.Println("verification token inserted", user.VerificationToken)

	if err != nil {
		return fmt.Errorf("failed to insert token to db: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (u *userRepository) InsertRefreshToken(ctx context.Context, token pgtype.Text, user_id pgtype.UUID) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	user, err := qtx.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
		RefreshToken: token,
		ID:           user_id,
	})
	log.Println("verification token inserted", user.VerificationToken)

	if err != nil {
		return fmt.Errorf("failed to insert token to db: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
