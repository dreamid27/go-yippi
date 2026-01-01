package entities

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodVA_BCA      PaymentMethod = "va_bca"
	PaymentMethodVA_Mandiri  PaymentMethod = "va_mandiri"
	PaymentMethodVA_BNI      PaymentMethod = "va_bni"
	PaymentMethodVA_BRI      PaymentMethod = "va_bri"
	PaymentMethodGoPay       PaymentMethod = "gopay"
	PaymentMethodOVO         PaymentMethod = "ovo"
	PaymentMethodDANA        PaymentMethod = "dana"
	PaymentMethodShopeePay   PaymentMethod = "shopeepay"
	PaymentMethodQRIS        PaymentMethod = "qris"
	PaymentMethodCreditCard  PaymentMethod = "credit_card"
	PaymentMethodCOD         PaymentMethod = "cod"
)

// PaymentStatus represents the payment status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment represents payment information for an order
type Payment struct {
	ID                      uuid.UUID
	OrderID                 uuid.UUID
	Method                  PaymentMethod
	Status                  PaymentStatus
	Amount                  float64
	VaNumber                string // Nullable - for VA payments
	VaExpirationDate        *time.Time
	PaymentURL              string // Nullable - Midtrans redirect URL
	PaymentToken            string // Nullable - Midtrans token
	TransactionID           string // Nullable - Midtrans transaction ID
	SignatureKey            string // Nullable - Midtrans signature for webhook verification
	RawMidtransResponse     map[string]interface{} // Nullable - for debugging
	PaidAt                  *time.Time
	ExpiresAt               *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// TableName returns the database table name
func (Payment) TableName() string {
	return "payments"
}

// IsPaid checks if the payment has been paid
func (p *Payment) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsPending checks if the payment is still pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// CanBePaid checks if the payment can still be completed
func (p *Payment) CanBePaid() bool {
	return p.Status == PaymentStatusPending &&
		(p.ExpiresAt == nil || p.ExpiresAt.After(time.Now()))
}