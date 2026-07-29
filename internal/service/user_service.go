package service

import (
	"context"

	"example.com/internal/models"
	"example.com/internal/repository"
)

type UserService interface {
	Create(ctx context.Context, email string) (*models.Users, error)
	GetUsers(ctx context.Context) ([]models.Users, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *userService {
	return &userService{
		userRepo: userRepo,
	}
}

func (u *userService) Create(ctx context.Context, email string) (*models.Users, error) {
	user, err := u.userRepo.Create(ctx, email)
	if err != nil {
		return nil, err
	}
	return &models.Users{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (u *userService) GetUsers(ctx context.Context) ([]models.Users, error) {
	users, err := u.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
