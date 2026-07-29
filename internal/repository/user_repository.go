package repository

import (
	"context"
	"fmt"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, email string, full_name string, password string) (*models.Users, error)
	GetUsers(ctx context.Context) ([]models.Users, error)
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

	user, err := qtx.CreateUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Users{
		ID:    int(user.ID),
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
			ID:    int(user.ID),
			Email: user.Email,
		}
	}

	return result, nil
}

func (u *userRepository) CreateLink(ctx context.Context, user_id int, title, url string) (*models.Links, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	link, err := qtx.CreateLink(ctx, sqlc.CreateLinkParams{
		Title:  title,
		Url:    url,
		UserID: int64(user_id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Links{
		ID:      int(link.ID),
		Title:   link.Title,
		Url:     link.Url,
		User_ID: int(link.UserID),
	}, nil
}

func (u *userRepository) CreateTags(ctx context.Context, name string) (*models.Tags, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := u.Queries.WithTx(tx)

	tag, err := qtx.CreateTag(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Tags{
		ID:   int(tag.ID),
		Name: tag.Name,
	}, nil
}
