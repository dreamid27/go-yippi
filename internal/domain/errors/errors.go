package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates that the provided input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateEntry indicates that the resource already exists
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrUnauthorized indicates that the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the user does not have permission
	ErrForbidden = errors.New("forbidden")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal error")

	// ErrInsufficientStock indicates that there is not enough stock
	ErrInsufficientStock = errors.New("insufficient stock")

	// ErrVariantInactive indicates that the variant is not active
	ErrVariantInactive = errors.New("variant is inactive")
)

// NotFoundError represents a resource not found error with additional context
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %v not found", e.Resource, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource string, id interface{}) error {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// DuplicateError represents a duplicate entry error
type DuplicateError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s with %s '%v' already exists", e.Resource, e.Field, e.Value)
}

func (e *DuplicateError) Is(target error) bool {
	return target == ErrDuplicateEntry
}

// NewDuplicateError creates a new DuplicateError
func NewDuplicateError(resource, field string, value interface{}) error {
	return &DuplicateError{
		Resource: resource,
		Field:    field,
		Value:    value,
	}
}

// InsufficientStockError represents insufficient stock error
type InsufficientStockError struct {
	VariantID interface{}
	Available int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for variant %v: available %d, requested %d", e.VariantID, e.Available, e.Requested)
}

func (e *InsufficientStockError) Is(target error) bool {
	return target == ErrInsufficientStock
}

// NewInsufficientStockError creates a new InsufficientStockError
func NewInsufficientStockError(variantID interface{}, available, requested int) error {
	return &InsufficientStockError{
		VariantID: variantID,
		Available: available,
		Requested: requested,
	}
}

// VariantInactiveError represents variant inactive error
type VariantInactiveError struct {
	VariantID interface{}
}

func (e *VariantInactiveError) Error() string {
	return fmt.Sprintf("variant %v is inactive", e.VariantID)
}

func (e *VariantInactiveError) Is(target error) bool {
	return target == ErrVariantInactive
}

// NewVariantInactiveError creates a new VariantInactiveError
func NewVariantInactiveError(variantID interface{}) error {
	return &VariantInactiveError{
		VariantID: variantID,
	}
}

// Address errors
var (
	// ErrAddressNotFound indicates address not found
	ErrAddressNotFound = errors.New("address not found")

	// ErrAddressNotOwned indicates address doesn't belong to user
	ErrAddressNotOwned = errors.New("address does not belong to user")

	// ErrAddressDeleted indicates address has been soft-deleted
	ErrAddressDeleted = errors.New("address has been deleted")

	// ErrAddressInUse indicates address is used in active orders
	ErrAddressInUse = errors.New("address is used in active orders")

	// ErrInvalidPhone indicates invalid phone number format
	ErrInvalidPhone = errors.New("invalid phone number format")

	// ErrInvalidPostalCode indicates invalid postal code
	ErrInvalidPostalCode = errors.New("invalid postal code format (must be 5 digits)")

	// ErrOnlyOneDefault indicates only one default address allowed
	ErrOnlyOneDefault = errors.New("only one default address allowed per user")
)

// Order errors
var (
	// ErrOrderNotFound indicates order not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrOrderNotOwned indicates order doesn't belong to user
	ErrOrderNotOwned = errors.New("order does not belong to user")

	// ErrCartEmpty indicates cart is empty
	ErrCartEmpty = errors.New("cart is empty")

	// ErrOrderCannotCancel indicates order cannot be cancelled in current status
	ErrOrderCannotCancel = errors.New("order cannot be cancelled in current status")

	// ErrOrderNotPendingPayment indicates order is not in pending_payment status
	ErrOrderNotPendingPayment = errors.New("order is not in pending_payment status")

	// ErrShippingCostMismatch indicates shipping cost doesn't match
	ErrShippingCostMismatch = errors.New("shipping cost does not match calculated cost")
)

// Payment errors
var (
	// ErrPaymentNotFound indicates payment not found
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrInvalidSignature indicates invalid webhook signature
	ErrInvalidSignature = errors.New("invalid webhook signature")

	// ErrPaymentExpired indicates payment has expired
	ErrPaymentExpired = errors.New("payment has expired")
)

// Voucher errors (for future implementation)
var (
	// ErrVoucherInvalid indicates voucher code is invalid
	ErrVoucherInvalid = errors.New("voucher code is invalid")

	// ErrVoucherExpired indicates voucher has expired
	ErrVoucherExpired = errors.New("voucher has expired")

	// ErrVoucherExhausted indicates voucher has reached max uses
	ErrVoucherExhausted = errors.New("voucher has reached maximum uses")

	// ErrVoucherMinPurchase indicates minimum purchase not met
	ErrVoucherMinPurchase = errors.New("minimum purchase amount not met")
)

// RajaOngkir errors
var (
	// ErrProvinceNotFound indicates province not found
	ErrProvinceNotFound = errors.New("province not found")

	// ErrCityNotFound indicates city not found
	ErrCityNotFound = errors.New("city not found")

	// ErrRajaongkirAPI indicates RajaOngkir API error
	ErrRajaongkirAPI = errors.New("rajaongkir API error")
)

// OrderCannotCancelError represents order cannot cancel error with status
type OrderCannotCancelError struct {
	Status string
}

func (e *OrderCannotCancelError) Error() string {
	return fmt.Sprintf("order cannot be cancelled in status: %s", e.Status)
}

func (e *OrderCannotCancelError) Is(target error) bool {
	return target == ErrOrderCannotCancel
}

// NewOrderCannotCancelError creates new OrderCannotCancelError
func NewOrderCannotCancelError(status string) error {
	return &OrderCannotCancelError{Status: status}
}

// ShippingCostMismatchError represents shipping cost mismatch
type ShippingCostMismatchError struct {
	Provided   float64
	Calculated float64
}

func (e *ShippingCostMismatchError) Error() string {
	return fmt.Sprintf("shipping cost mismatch: provided %.2f, calculated %.2f", e.Provided, e.Calculated)
}

func (e *ShippingCostMismatchError) Is(target error) bool {
	return target == ErrShippingCostMismatch
}

// NewShippingCostMismatchError creates new ShippingCostMismatchError
func NewShippingCostMismatchError(provided, calculated float64) error {
	return &ShippingCostMismatchError{
		Provided:   provided,
		Calculated: calculated,
	}
}

// VoucherMinPurchaseError represents minimum purchase not met
type VoucherMinPurchaseError struct {
	Min float64
}

func (e *VoucherMinPurchaseError) Error() string {
	return fmt.Sprintf("minimum purchase amount %.2f not met", e.Min)
}

func (e *VoucherMinPurchaseError) Is(target error) bool {
	return target == ErrVoucherMinPurchase
}

// NewVoucherMinPurchaseError creates new VoucherMinPurchaseError
func NewVoucherMinPurchaseError(min float64) error {
	return &VoucherMinPurchaseError{Min: min}
}
