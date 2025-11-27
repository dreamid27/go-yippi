package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// BrandService defines the interface for brand business logic operations
type BrandService interface {
	CreateBrand(ctx context.Context, brand *entities.Brand) error
	GetBrand(ctx context.Context, id uuid.UUID) (*entities.Brand, error)
	GetBrandByName(ctx context.Context, name string) (*entities.Brand, error)
	ListBrands(ctx context.Context) ([]*entities.Brand, error)
	UpdateBrand(ctx context.Context, brand *entities.Brand) error
	DeleteBrand(ctx context.Context, id uuid.UUID) error
}
