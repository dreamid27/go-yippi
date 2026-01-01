package entities

import (
	"time"

	"github.com/google/uuid"
)

// OrderItem represents an item within an order
type OrderItem struct {
	ID                           uuid.UUID
	OrderID                      uuid.UUID
	ProductID                    uuid.UUID
	ProductVariantID             uuid.UUID
	ProductName                  string
	ProductVariantAttributes     map[string]string // e.g., {"size": "M", "color": "Black"}
	SKU                          string
	Quantity                     int
	UnitPrice                    float64 // Price at time of order (snapshot)
	TotalPrice                   float64 // unitPrice * quantity
	ProductWeight                int     // in grams, for shipping calculation
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

// TableName returns the database table name
func (OrderItem) TableName() string {
	return "order_items"
}

// GetSubtotal calculates the subtotal for this order item
func (oi *OrderItem) GetSubtotal() float64 {
	return oi.UnitPrice * float64(oi.Quantity)
}