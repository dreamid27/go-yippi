package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.PaymentStatus) error
	MarkAsPaid(ctx context.Context, id uuid.UUID) error
}