package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProductVariant represents a variant of a product with specific attributes
type ProductVariant struct {
	ID              uuid.UUID         `json:"id"`
	ProductID       uuid.UUID         `json:"product_id"`
	SKU             string            `json:"sku"`
	Attributes      map[string]string `json:"attributes"` // {"size": "M", "color": "Red"}
	StockQuantity   int               `json:"stock_quantity"`
	PriceAdjustment float64           `json:"price_adjustment"`
	IsActive        bool              `json:"is_active"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CalculateFinalPrice calculates the final price based on product base price
func (v *ProductVariant) CalculateFinalPrice(basePrice float64) float64 {
	return basePrice + v.PriceAdjustment
}

// IsInStock checks if the variant has stock available
func (v *ProductVariant) IsInStock() bool {
	return v.StockQuantity > 0
}

// IsAvailable checks if variant is active and in stock
func (v *ProductVariant) IsAvailable() bool {
	return v.IsActive && v.IsInStock()
}
