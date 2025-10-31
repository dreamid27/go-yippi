package services

import (
	"context"
	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/ports"
)

// UserService handles business logic for users
type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
	// Add your business logic here
	return s.repo.Create(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, id int) (*entities.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context) ([]*entities.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) error {
	return s.repo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
