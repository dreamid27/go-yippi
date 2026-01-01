package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// StockValidatorService defines the interface for stock validation
type StockValidatorService interface {
	// ValidateStock validates if requested quantity is available
	ValidateStock(ctx context.Context, variantID uuid.UUID, requestedQty int) error

	// ValidateCartStock validates stock for all items in a cart
	ValidateCartStock(ctx context.Context, cartItems []*entities.CartItem) ([]StockValidationError, error)
}

// StockValidationError represents a stock validation error for a cart item
type StockValidationError struct {
	CartItemID uuid.UUID
	VariantID  uuid.UUID
	Available  int
	InCart     int
	Message    string
}
