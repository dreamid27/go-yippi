package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
)

// PriceCalculatorService defines the interface for price calculations
type PriceCalculatorService interface {
	// CalculateVariantPrice calculates final price for a variant
	CalculateVariantPrice(ctx context.Context, basePrice float64, priceAdjustment float64) float64

	// CalculateCartSubtotal calculates total price for cart items
	CalculateCartSubtotal(ctx context.Context, cartItems []*entities.CartItem) float64

	// CalculateItemSubtotal calculates subtotal for a single item
	CalculateItemSubtotal(ctx context.Context, priceSnapshot float64, quantity int) float64

	// GetPriceRange returns min and max prices from variants
	GetPriceRange(ctx context.Context, variants []*entities.ProductVariant, basePrice float64) (minPrice float64, maxPrice float64)
}
