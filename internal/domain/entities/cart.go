package entities

import (
	"time"

	"github.com/google/uuid"
)

// Cart represents a user's shopping cart
type Cart struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItem represents an item in the cart
type CartItem struct {
	ID               uuid.UUID `json:"id"`
	CartID           uuid.UUID `json:"cart_id"`
	ProductVariantID uuid.UUID `json:"product_variant_id"`
	Quantity         int       `json:"quantity"`
	PriceSnapshot    float64   `json:"price_snapshot"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CalculateSubtotal calculates subtotal for this cart item
func (ci *CartItem) CalculateSubtotal() float64 {
	return ci.PriceSnapshot * float64(ci.Quantity)
}

// CartDetail represents cart with all items and computed fields
type CartDetail struct {
	Cart      *Cart       `json:"cart"`
	Items     []CartItem  `json:"items"`
	Subtotal  float64     `json:"subtotal"`
	ItemCount int         `json:"item_count"`
}

// CalculateTotals computes subtotal and item count
func (cd *CartDetail) CalculateTotals() {
	cd.Subtotal = 0
	cd.ItemCount = 0
	for _, item := range cd.Items {
		cd.Subtotal += item.CalculateSubtotal()
		cd.ItemCount += item.Quantity
	}
}
