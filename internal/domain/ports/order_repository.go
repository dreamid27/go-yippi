package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *entities.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Order, error)
	GetByUserIDWithStatus(ctx context.Context, userID uuid.UUID, status entities.OrderStatus) ([]*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OrderStatus) error
	Cancel(ctx context.Context, id uuid.UUID, reason string) error
}