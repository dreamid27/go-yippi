package entities

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPendingPayment   OrderStatus = "pending_payment"
	OrderStatusPaymentVerified OrderStatus = "payment_verified"
	OrderStatusProcessing      OrderStatus = "processing"
	OrderStatusShipped         OrderStatus = "shipped"
	OrderStatusDelivered       OrderStatus = "delivered"
	OrderStatusCompleted       OrderStatus = "completed"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// Order represents a customer order
type Order struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	ShippingAddressID   uuid.UUID
	Status              OrderStatus
	Subtotal            float64
	ShippingCost        float64
	DiscountAmount      float64
	VoucherCode         string // Nullable - for future voucher implementation
	VoucherDiscount     float64
	TotalAmount         float64
	Notes               string // Nullable - customer notes
	AdminNotes          string // Nullable - internal notes
	CancelledAt         *time.Time
	CancellationReason  string // Nullable
	PaidAt              *time.Time
	ShippedAt           *time.Time
	DeliveredAt         *time.Time
	CompletedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName returns the database table name
func (Order) TableName() string {
	return "orders"
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPendingPayment
}

// IsPaid checks if the order has been paid
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaymentVerified ||
		o.Status == OrderStatusProcessing ||
		o.Status == OrderStatusShipped ||
		o.Status == OrderStatusDelivered ||
		o.Status == OrderStatusCompleted
}