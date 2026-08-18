package repository

import (
	"context"
	"fmt"

	"example.com/internal/db/sqlc"
	"example.com/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository interface {
	WithTx(tx pgx.Tx) NotificationRepository
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error)
}

type notificationRepository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *notificationRepository) WithTx(tx pgx.Tx) NotificationRepository {
	return &notificationRepository{
		Queries: r.Queries.WithTx(tx),
		pool:    r.pool,
	}
}

func (n *notificationRepository) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error) {
	notification, err := n.Queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		UserID:  req.UserID,
		Type:    req.Type,
		Message: req.Message,
		Title:   req.Title,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &models.NotificationResponse{
		ID:      notification.ID,
		Message: notification.Message,
		Title:   notification.Title,
		Type:    notification.Type,
		UserID:  notification.UserID,
	}, nil
}
