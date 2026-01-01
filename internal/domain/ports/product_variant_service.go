package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// ProductVariantService defines the interface for product variant business logic
type ProductVariantService interface {
	// CreateVariant creates a new product variant
	CreateVariant(ctx context.Context, productID uuid.UUID, sku string, attributes map[string]string, stockQuantity int, priceAdjustment float64, isActive *bool) (*entities.ProductVariant, error)

	// UpdateVariant updates an existing variant
	UpdateVariant(ctx context.Context, variantID uuid.UUID, sku *string, attributes *map[string]string, stockQuantity *int, priceAdjustment *float64, isActive *bool) (*entities.ProductVariant, error)

	// GetVariantsByProduct returns all variants for a product
	GetVariantsByProduct(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error)

	// UpdateStock updates stock quantity for a variant
	UpdateStock(ctx context.Context, variantID uuid.UUID, newStockQuantity int) (*entities.ProductVariant, error)

	// DeleteVariant deletes a variant (soft delete if in use)
	DeleteVariant(ctx context.Context, variantID uuid.UUID) error

	// GetVariantByID returns a variant by ID
	GetVariantByID(ctx context.Context, variantID uuid.UUID) (*entities.ProductVariant, error)
}
