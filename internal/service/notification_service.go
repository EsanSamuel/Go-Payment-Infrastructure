package service

import (
	"context"
	"fmt"

	"example.com/internal/models"
	"example.com/internal/repository"
	"example.com/internal/ws"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
	hub              *ws.Hub
}

func NewNotificationService(notificationRepo repository.NotificationRepository, hub *ws.Hub) NotificationService {
	return &notificationService{notificationRepo: notificationRepo, hub: hub}
}

func (s *notificationService) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error) {
	notification, err := s.notificationRepo.CreateNotification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.hub.SendNotification(notification.UserID.String(), *notification)
	return notification, nil
}
