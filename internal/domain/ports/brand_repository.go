package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// BrandRepository defines the interface for brand data operations
type BrandRepository interface {
	Create(ctx context.Context, brand *entities.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Brand, error)
	GetByName(ctx context.Context, name string) (*entities.Brand, error)
	List(ctx context.Context) ([]*entities.Brand, error)
	Update(ctx context.Context, brand *entities.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
}
