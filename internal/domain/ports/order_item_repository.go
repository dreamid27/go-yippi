package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// OrderItemRepository defines the interface for order item data operations
type OrderItemRepository interface {
	Create(ctx context.Context, orderItem *entities.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.OrderItem, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entities.OrderItem, error)
	Update(ctx context.Context, orderItem *entities.OrderItem) error
}