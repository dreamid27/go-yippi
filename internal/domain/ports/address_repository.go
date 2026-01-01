package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// AddressRepository defines the interface for address data operations
type AddressRepository interface {
	Create(ctx context.Context, address *entities.Address) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Address, error)
	Update(ctx context.Context, address *entities.Address) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SetDefaultAddress(ctx context.Context, id uuid.UUID) error
	GetDefaultAddress(ctx context.Context, userID uuid.UUID) (*entities.Address, error)
}