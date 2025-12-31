package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// ProductVariantRepository defines the interface for product variant persistence
type ProductVariantRepository interface {
	// Create creates a new product variant
	Create(ctx context.Context, variant *entities.ProductVariant) error

	// FindByID finds a variant by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error)

	// FindBySKU finds a variant by SKU
	FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error)

	// FindByProductID finds all variants for a product
	FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error)

	// Update updates an existing variant
	Update(ctx context.Context, variant *entities.ProductVariant) error

	// UpdateStock updates stock quantity for a variant
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error

	// Delete deletes a variant
	Delete(ctx context.Context, id uuid.UUID) error
}
