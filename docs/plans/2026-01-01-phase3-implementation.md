# Phase 3 Backend Implementation Plan - Checkout & Payment System

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build complete checkout and payment system with address management, RajaOngkir shipping integration, Midtrans payment processing, order lifecycle management, and automated notifications.

**Architecture:** Hexagonal architecture with strict dependency flow. New schemas (Address, Order, OrderItem, Payment) integrate with existing Product/Cart entities. RajaOngkir and Midtrans are isolated in infrastructure layer with in-memory caching. Service layer orchestrates order creation, payment handling, and status transitions with stock deduction after payment verification.

**Tech Stack:** Go 1.23, Ent ORM, PostgreSQL, RajaOngkir API, Midtrans Snap SDK, Huma v2, OpenTelemetry, Zap, robfig/cron, SMTP

---

## Task 1: Add Phase 3 Domain Errors

**Files:**
- Modify: `internal/domain/errors/errors.go`

**Step 1: Add Phase 3 error definitions**

Append to `internal/domain/errors/errors.go`:

```go
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
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/errors`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/errors/errors.go
git commit -m "feat(domain): add Phase 3 error definitions for address, order, payment"
```

---

## Task 2: Create Address Schema

**Files:**
- Create: `internal/adapters/persistence/db/schema/address.go`

**Step 1: Create Address schema**

Create file `internal/adapters/persistence/db/schema/address.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Address holds the schema definition for the Address entity.
type Address struct {
	ent.Schema
}

// Fields of the Address.
func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("User ID - address owner"),

		field.String("label").
			NotEmpty().
			Comment("Address label: Home, Office, Parents, etc."),

		field.String("recipient_name").
			NotEmpty().
			Comment("Full name of recipient"),

		field.String("phone").
			NotEmpty().
			Comment("Recipient phone number (Indonesian format)"),

		field.String("address_line1").
			NotEmpty().
			Comment("Street address, building, floor, unit number"),

		field.String("address_line2").
			Optional().
			Comment("Additional address information"),

		field.String("province_id").
			NotEmpty().
			Comment("RajaOngkir province ID"),

		field.String("province_name").
			NotEmpty().
			Comment("Province name (cached from RajaOngkir)"),

		field.String("city_id").
			NotEmpty().
			Comment("RajaOngkir city ID"),

		field.String("city_name").
			NotEmpty().
			Comment("City name (cached from RajaOngkir)"),

		field.String("district").
			Optional().
			Comment("District/Sub-district (Kecamatan)"),

		field.String("postal_code").
			NotEmpty().
			Comment("Postal code"),

		field.Bool("is_default").
			Default(false).
			Comment("Is this default shipping address"),

		field.Bool("is_deleted").
			Default(false).
			Comment("Soft delete flag"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Address.
func (Address) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("addresses").
			Unique().
			Required().
			Field("user_id"),

		edge.To("orders", Order.Type),
	}
}

// Indexes of the Address.
func (Address) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_default").Unique(), // Only one default per user
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/address.go
git commit -m "feat(schema): add Address schema with soft delete and default address"
```

---

## Task 3: Create Order Schema

**Files:**
- Create: `internal/adapters/persistence/db/schema/order.go`

**Step 1: Create Order schema**

Create file `internal/adapters/persistence/db/schema/order.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Order holds the schema definition for the Order entity.
type Order struct {
	ent.Schema
}

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("User ID - customer who placed the order"),

		field.UUID("shipping_address_id", uuid.UUID{}).
			Comment("Shipping address ID (reference to Address)"),

		field.String("order_number").
			Unique().
			NotEmpty().
			Comment("Human-readable order number: ORD-20250101-12345"),

		field.Enum("status").
			Values(
				"pending_payment",
				"payment_verified",
				"processing",
				"shipped",
				"delivered",
				"completed",
				"cancelled",
			).
			Default("pending_payment").
			Comment("Order status workflow"),

		// Shipping details
		field.String("courier").
			Optional().
			Comment("Courier name: JNE, J&T, SiCepat, Anteraja"),

		field.String("courier_service").
			Optional().
			Comment("Service type: OKE, REG, YES, CARGO"),

		field.Float("shipping_cost").
			Default(0).
			Comment("Shipping cost in IDR"),

		field.String("tracking_number").
			Optional().
			Comment("Courier tracking number (resi)"),

		// Order totals
		field.Float("subtotal").
			Positive().
			Comment("Sum of all items (price_snapshot * quantity)"),

		field.Float("discount").
			Default(0).
			Comment("Total discount amount (from vouchers/promos)"),

		field.Float("tax").
			Default(0).
			Comment("Tax amount (11% PPN in Indonesia)"),

		field.Float("total").
			Positive().
			Comment("Final total: subtotal + shipping_cost + tax - discount"),

		// Metadata
		field.String("notes").
			Optional().
			Comment("Customer order notes"),

		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional metadata: source_device, referrer, etc."),

		field.Time("cancelled_at").
			Optional().
			Comment("Order cancellation timestamp"),

		field.String("cancellation_reason").
			Optional().
			Comment("Reason for cancellation"),

		field.Time("shipped_at").
			Optional().
			Comment("Shipment timestamp"),

		field.Time("delivered_at").
			Optional().
			Comment("Delivery timestamp"),

		field.Time("completed_at").
			Optional().
			Comment("Order completion timestamp"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("orders").
			Unique().
			Required().
			Field("user_id"),

		edge.From("shipping_address", Address.Type).
			Ref("orders").
			Unique().
			Required().
			Field("shipping_address_id"),

		edge.To("items", OrderItem.Type),

		edge.To("payment", Payment.Type).
			Unique(),
	}
}

// Indexes of the Order.
func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("order_number").Unique(),
		index.Fields("status"),
		index.Fields("created_at"),
		// Composite index for user's pending orders
		index.Fields("user_id", "status", "created_at"),
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors (will have errors about OrderItem and Payment - that's OK for now)

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/order.go
git commit -m "feat(schema): add Order schema with status workflow and price snapshots"
```

---

## Task 4: Create OrderItem Schema

**Files:**
- Create: `internal/adapters/persistence/db/schema/order_item.go`

**Step 1: Create OrderItem schema**

Create file `internal/adapters/persistence/db/schema/order_item.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrderItem holds the schema definition for the OrderItem entity.
type OrderItem struct {
	ent.Schema
}

// Fields of the OrderItem.
func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("order_id", uuid.UUID{}).
			Comment("Order ID"),

		field.UUID("product_id", uuid.UUID{}).
			Comment("Product ID (reference to Product)"),

		field.UUID("product_variant_id", uuid.UUID{}).
			Comment("Product Variant ID (reference to ProductVariant)"),

		field.String("product_name").
			NotEmpty().
			Comment("Product name at time of order (snapshot)"),

		field.JSON("product_variant_attributes", map[string]string{}).
			Comment("Variant attributes at time of order: {\"size\": \"M\", \"color\": \"Black\"}"),

		field.String("product_image_url").
			Optional().
			Comment("Main product image URL at time of order"),

		field.String("sku").
			NotEmpty().
			Comment("Variant SKU at time of order"),

		field.Int("quantity").
			Positive().
			Comment("Quantity ordered"),

		field.Float("price").
			Positive().
			Comment("Unit price at time of order (snapshot)"),

		field.Float("subtotal").
			Positive().
			Comment("Item subtotal: price * quantity"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OrderItem.
func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("items").
			Unique().
			Required().
			Field("order_id"),

		edge.From("product", Product.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("product_id"),

		edge.From("product_variant", ProductVariant.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("product_variant_id"),
	}
}

// Indexes of the OrderItem.
func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("product_id"),
		index.Fields("product_variant_id"),
		index.Fields("order_id", "product_variant_id"),
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors (will have errors about Payment - that's OK)

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/order_item.go
git commit -m "feat(schema): add OrderItem schema with price snapshots"
```

---

## Task 5: Create Payment Schema

**Files:**
- Create: `internal/adapters/persistence/db/schema/payment.go`

**Step 1: Create Payment schema**

Create file `internal/adapters/persistence/db/schema/payment.go`:

```go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Payment holds the schema definition for the Payment entity.
type Payment struct {
	ent.Schema
}

// Fields of the Payment.
func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("order_id", uuid.UUID{}).
			Comment("Order ID (one-to-one with Order)"),

		field.Enum("method").
			Values(
				"va_bca",
				"va_mandiri",
				"va_bni",
				"va_bri",
				"gopay",
				"ovo",
				"dana",
				"shopeepay",
				"qris",
				"credit_card",
				"cod",
			).
			NotEmpty().
			Comment("Payment method"),

		field.Enum("status").
			Values(
				"pending",
				"processing",
				"success",
				"failed",
				"expired",
				"cancelled",
			).
			Default("pending").
			Comment("Payment status from Midtrans"),

		field.Float("amount").
			Positive().
			Comment("Payment amount (should match order.total)"),

		field.String("midtrans_transaction_id").
			Optional().
			Comment("Midtrans transaction ID"),

		field.String("midtrans_payment_type").
			Optional().
			Comment("Midtrans payment type: bank_transfer, e_wallet, qris, credit_card"),

		field.String("va_number").
			Optional().
			Comment("Virtual Account number (for VA payments)"),

		field.String("va_bank").
			Optional().
			Comment("Virtual Account bank: BCA, Mandiri, BNI, BRI"),

		field.String("payment_url").
			Optional().
			Comment("Midtrans Snap payment URL or redirect URL"),

		field.Time("expires_at").
			Optional().
			Comment("Payment expiration timestamp (24h from creation)"),

		field.Time("paid_at").
			Optional().
			Comment("Successful payment timestamp"),

		field.String("failure_reason").
			Optional().
			Comment("Reason for payment failure (from Midtrans)"),

		field.JSON("raw_midtrans_response", map[string]interface{}{}).
			Optional().
			Comment("Raw Midtrans webhook/response data for debugging"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Payment.
func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("payment").
			Unique().
			Required().
			Field("order_id"),
	}
}

// Indexes of the Payment.
func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id").Unique(),
		index.Fields("midtrans_transaction_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/payment.go
git commit -m "feat(schema): add Payment schema with Midtrans integration"
```

---

## Task 6: Update User Schema with New Edges

**Files:**
- Modify: `internal/adapters/persistence/db/schema/user.go`

**Step 1: Add addresses and orders edges to User**

Update the `Edges()` function in `internal/adapters/persistence/db/schema/user.go` (around line 49):

Replace:
```go
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),

		// One cart per user
		edge.To("cart", Cart.Type).
			Unique(),
	}
}
```

With:
```go
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),

		// One cart per user
		edge.To("cart", Cart.Type).
			Unique(),

		// NEW: Address relationship
		edge.To("addresses", Address.Type),

		// NEW: Orders relationship
		edge.To("orders", Order.Type),
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/user.go
git commit -m "feat(schema): add addresses and orders edges to User schema"
```

---

## Task 7: Update Product Schema with Order Items Edge

**Files:**
- Modify: `internal/adapters/persistence/db/schema/product.go`

**Step 1: Add order_items edge to Product**

Update the `Edges()` function in `internal/adapters/persistence/db/schema/product.go` (around line 105):

Replace:
```go
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		// Many-to-one ke Category (pakai kolom category_id, bukan join table category_products)
		edge.From("category", Category.Type).
			Ref("products").
			Unique().
			Field("category_id"),

		// Many-to-one ke Brand (pakai kolom brand_id)
		edge.From("brand", Brand.Type).
			Ref("products").
			Unique().
			Field("brand_id"),

		// One-to-many to ProductVariant
		edge.To("variants", ProductVariant.Type),
	}
}
```

With:
```go
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		// Many-to-one ke Category (pakai kolom category_id, bukan join table category_products)
		edge.From("category", Category.Type).
			Ref("products").
			Unique().
			Field("category_id"),

		// Many-to-one ke Brand (pakai kolom brand_id)
		edge.From("brand", Brand.Type).
			Ref("products").
			Unique().
			Field("brand_id"),

		// One-to-many to ProductVariant
		edge.To("variants", ProductVariant.Type),

		// NEW: Order items relationship (historical data)
		edge.To("order_items", OrderItem.Type),
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/product.go
git commit -m "feat(schema): add order_items edge to Product schema"
```

---

## Task 8: Update ProductVariant Schema with Order Items Edge

**Files:**
- Modify: `internal/adapters/persistence/db/schema/product_variant.go`

**Step 1: Add order_items edge to ProductVariant**

Find the `Edges()` function in `internal/adapters/persistence/db/schema/product_variant.go` and update it:

If it looks like:
```go
edge.To("cart_items", CartItem.Type),
```

Add after it:
```go
// NEW: Order items relationship (historical data)
edge.To("order_items", OrderItem.Type),
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/schema`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/adapters/persistence/db/schema/product_variant.go
git commit -m "feat(schema): add order_items edge to ProductVariant schema"
```

---

## Task 9: Generate Ent Code

**Files:**
- Generate: `internal/adapters/persistence/db/ent/` (all generated files)

**Step 1: Generate Ent code**

Run: `make generate`

Or directly:
```bash
go run -mod=mod entgo.io/ent/cmd/ent generate --target ./internal/adapters/persistence/db/ent ./internal/adapters/persistence/db/schema
```

Expected output: Generated files in `internal/adapters/persistence/db/ent/` including:
- `address.go`
- `address_create.go`
- `order.go`
- `order_create.go`
- `order_item.go`
- `order_item_create.go`
- `payment.go`
- `payment_create.go`
- Updated `user.go`, `product.go`, `product_variant.go` with new edges

**Step 2: Verify compilation**

Run: `go build ./internal/adapters/persistence/db/ent`
Expected: No errors

**Step 3: Commit generated code**

```bash
git add internal/adapters/persistence/db/ent/
git commit -m "chore: generate Ent code for Phase 3 schemas (Address, Order, OrderItem, Payment)"
```

---

## Task 10: Create Database Migration

**Files:**
- Create: `internal/adapters/persistence/db/migrations/20260101_phase3_checkout_payment.up.sql`
- Create: `internal/adapters/persistence/db/migrations/20260101_phase3_checkout_payment.down.sql`

**Step 1: Create up migration**

Create file `internal/adapters/persistence/db/migrations/20260101_phase3_checkout_payment.up.sql`:

```sql
-- Migration: Phase 3 - Checkout & Payment System
-- Created: 2026-01-01

-- Create addresses table
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    recipient_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    address_line1 VARCHAR(500) NOT NULL,
    address_line2 VARCHAR(500),
    province_id VARCHAR(10) NOT NULL,
    province_name VARCHAR(100) NOT NULL,
    city_id VARCHAR(10) NOT NULL,
    city_name VARCHAR(100) NOT NULL,
    district VARCHAR(100),
    postal_code VARCHAR(10) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for addresses
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_addresses_user_default ON addresses(user_id, is_default) WHERE is_default = TRUE;

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shipping_address_id UUID NOT NULL REFERENCES addresses(id),
    order_number VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_payment',
    courier VARCHAR(50),
    courier_service VARCHAR(50),
    shipping_cost FLOAT DEFAULT 0 NOT NULL,
    tracking_number VARCHAR(100),
    subtotal FLOAT NOT NULL CHECK (subtotal > 0),
    discount FLOAT DEFAULT 0 NOT NULL,
    tax FLOAT DEFAULT 0 NOT NULL,
    total FLOAT NOT NULL CHECK (total > 0),
    notes TEXT,
    metadata JSONB,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for orders
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_user_status_created ON orders(user_id, status, created_at);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_variant_id UUID NOT NULL REFERENCES product_variants(id),
    product_name VARCHAR(255) NOT NULL,
    product_variant_attributes JSONB NOT NULL,
    product_image_url TEXT,
    sku VARCHAR(100) NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    price FLOAT NOT NULL CHECK (price > 0),
    subtotal FLOAT NOT NULL CHECK (subtotal > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for order_items
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_variant_id ON order_items(product_variant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_variant ON order_items(order_id, product_variant_id);

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    amount FLOAT NOT NULL CHECK (amount > 0),
    midtrans_transaction_id VARCHAR(100),
    midtrans_payment_type VARCHAR(50),
    va_number VARCHAR(50),
    va_bank VARCHAR(20),
    payment_url TEXT,
    expires_at TIMESTAMP,
    paid_at TIMESTAMP,
    failure_reason TEXT,
    raw_midtrans_response JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for payments
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_midtrans_transaction_id ON payments(midtrans_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_expires_at ON payments(expires_at);

-- Create updated_at trigger function (if not exists)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_addresses_updated_at BEFORE UPDATE ON addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_order_items_updated_at BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Step 2: Create down migration**

Create file `internal/adapters/persistence/db/migrations/20260101_phase3_checkout_payment.down.sql`:

```sql
-- Rollback: Phase 3 - Checkout & Payment System

-- Drop triggers
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- Drop tables (in correct order due to foreign keys)
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS addresses;
```

**Step 3: Test migration on dev database**

Run: `psql -h localhost -U admin -d go-yippi -f internal/adapters/persistence/db/migrations/20260101_phase3_checkout_payment.up.sql`

Expected: All tables created successfully

Verify: Check tables exist
```bash
psql -h localhost -U admin -d go-yippi -c "\dt"
```

Expected output: Tables `addresses`, `orders`, `order_items`, `payments` listed

**Step 4: Commit migration files**

```bash
git add internal/adapters/persistence/db/migrations/
git commit -m "feat(db): add Phase 3 migration for checkout and payment system"
```

---

## Task 11: Create Domain Entities - Address

**Files:**
- Create: `internal/domain/entities/address.go`

**Step 1: Create Address entity**

Create file `internal/domain/entities/address.go`:

```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Address represents a user's shipping address
type Address struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Label              string
	RecipientName      string
	Phone              string
	AddressLine1       string
	AddressLine2       *string
	ProvinceID         string
	ProvinceName       string
	CityID             string
	CityName           string
	District           *string
	PostalCode         string
	IsDefault          bool
	IsDeleted          bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// AddressStatus represents address status
type AddressStatus string

const (
	AddressStatusActive    AddressStatus = "active"
	AddressStatusDeleted   AddressStatus = "deleted"
)
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/entities`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/entities/address.go
git commit -m "feat(domain): add Address entity"
```

---

## Task 12: Create Domain Entities - Order

**Files:**
- Create: `internal/domain/entities/order.go`

**Step 1: Create Order entity**

Create file `internal/domain/entities/order.go`:

```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a customer order
type Order struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ShippingAddressID  uuid.UUID
	OrderNumber        string
	Status             OrderStatus
	Courier            *string
	CourierService     *string
	ShippingCost       float64
	TrackingNumber     *string
	Subtotal           float64
	Discount           float64
	Tax                float64
	Total              float64
	Notes              *string
	Metadata           map[string]interface{}
	CancelledAt        *time.Time
	CancellationReason *string
	ShippedAt          *time.Time
	DeliveredAt        *time.Time
	CompletedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// Relations (loaded via eager loading)
	ShippingAddress *Address
	Items           []*OrderItem
	Payment         *Payment
}

// OrderStatus represents order status workflow
type OrderStatus string

const (
	OrderStatusPendingPayment  OrderStatus = "pending_payment"
	OrderStatusPaymentVerified OrderStatus = "payment_verified"
	OrderStatusProcessing      OrderStatus = "processing"
	OrderStatusShipped         OrderStatus = "shipped"
	OrderStatusDelivered       OrderStatus = "delivered"
	OrderStatusCompleted       OrderStatus = "completed"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// CanCancel checks if order can be cancelled
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPendingPayment
}

// IsPaid checks if order is paid
func (o *Order) IsPaid() bool {
	return o.Status != OrderStatusPendingPayment &&
		o.Status != OrderStatusCancelled
}

// GetStatusTransitions returns valid status transitions
func (o *Order) GetStatusTransitions() []OrderStatus {
	switch o.Status {
	case OrderStatusPendingPayment:
		return []OrderStatus{OrderStatusPaymentVerified, OrderStatusCancelled}
	case OrderStatusPaymentVerified:
		return []OrderStatus{OrderStatusProcessing}
	case OrderStatusProcessing:
		return []OrderStatus{OrderStatusShipped}
	case OrderStatusShipped:
		return []OrderStatus{OrderStatusDelivered}
	case OrderStatusDelivered:
		return []OrderStatus{OrderStatusCompleted}
	case OrderStatusCompleted:
		return []OrderStatus{} // Final state
	case OrderStatusCancelled:
		return []OrderStatus{} // Final state
	default:
		return []OrderStatus{}
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/entities`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/entities/order.go
git commit -m "feat(domain): add Order entity with status workflow"
```

---

## Task 13: Create Domain Entities - OrderItem

**Files:**
- Create: `internal/domain/entities/order_item.go`

**Step 1: Create OrderItem entity**

Create file `internal/domain/entities/order_item.go`:

```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// OrderItem represents an item in an order (price snapshot)
type OrderItem struct {
	ID                       uuid.UUID
	OrderID                  uuid.UUID
	ProductID                uuid.UUID
	ProductVariantID         uuid.UUID
	ProductName              string
	ProductVariantAttributes map[string]string
	ProductImageURL          *string
	SKU                      string
	Quantity                 int
	Price                    float64
	Subtotal                 float64
	CreatedAt                time.Time
	UpdatedAt                time.Time

	// Relations (loaded via eager loading)
	Product        *Product
	ProductVariant *ProductVariant
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/entities`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/entities/order_item.go
git commit -m "feat(domain): add OrderItem entity with price snapshots"
```

---

## Task 14: Create Domain Entities - Payment

**Files:**
- Create: `internal/domain/entities/payment.go`

**Step 1: Create Payment entity**

Create file `internal/domain/entities/payment.go`:

```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a payment for an order
type Payment struct {
	ID                     uuid.UUID
	OrderID                uuid.UUID
	Method                 PaymentMethod
	Status                 PaymentStatus
	Amount                 float64
	MidtransTransactionID  *string
	MidtransPaymentType    *string
	VANumber               *string
	VABank                 *string
	PaymentURL             *string
	ExpiresAt              *time.Time
	PaidAt                 *time.Time
	FailureReason          *string
	RawMidtransResponse    map[string]interface{}
	CreatedAt              time.Time
	UpdatedAt              time.Time

	// Relations (loaded via eager loading)
	Order *Order
}

// PaymentMethod represents payment method types
type PaymentMethod string

const (
	PaymentMethodVABCA      PaymentMethod = "va_bca"
	PaymentMethodVAMandiri  PaymentMethod = "va_mandiri"
	PaymentMethodVABNI      PaymentMethod = "va_bni"
	PaymentMethodVABRI      PaymentMethod = "va_bri"
	PaymentMethodGoPay      PaymentMethod = "gopay"
	PaymentMethodOVO        PaymentMethod = "ovo"
	PaymentMethodDANA       PaymentMethod = "dana"
	PaymentMethodShopeePay  PaymentMethod = "shopeepay"
	PaymentMethodQRIS       PaymentMethod = "qris"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodCOD        PaymentMethod = "cod"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSuccess    PaymentStatus = "success"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// IsPending checks if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsSuccessful checks if payment was successful
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusSuccess
}

// IsFailed checks if payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed ||
		p.Status == PaymentStatusExpired ||
		p.Status == PaymentStatusCancelled
}

// IsExpired checks if payment is expired
func (p *Payment) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/entities`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/entities/payment.go
git commit -m "feat(domain): add Payment entity with Midtrans integration"
```

---

## Task 15: Create RajaOngkir Types

**Files:**
- Create: `internal/infrastructure/rajaongkir/types.go`

**Step 1: Create RajaOngkir DTOs**

Create file `internal/infrastructure/rajaongkir/types.go`:

```go
package rajaongkir

// Province represents RajaOngkir province data
type Province struct {
	ID   string `json:"province_id"`
	Name string `json:"province"`
}

// City represents RajaOngkir city data
type City struct {
	ID         string `json:"city_id"`
	Name       string `json:"city_name"`
	Type       string `json:"type"`
	PostalCode string `json:"postal_code"`
	ProvinceID string `json:"province_id"`
}

// CostRequest represents RajaOngkir cost calculation request
type CostRequest struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	Weight      int     `json:"weight"`
	Courier     string  `json:"courier"`
}

// CostResponse represents RajaOngkir cost calculation response
type CostResponse struct {
	RajaOngkir struct {
		Results []CostResult `json:"results"`
	} `json:"rajaongkir"`
}

// CostResult represents shipping cost result for a courier
type CostResult struct {
	Code   string         `json:"code"`
	Name   string         `json:"name"`
	Costs  []CostService  `json:"costs"`
}

// CostService represents shipping cost for a service type
type CostService struct {
	Service     string  `json:"service"`
	Description string  `json:"description"`
	Cost        []CostDetail `json:"cost"`
}

// CostDetail represents cost breakdown
type CostDetail struct {
	Value int64  `json:"value"`
	Etd   string `json:"etd"`
	Note  string `json:"note"`
}

// APIResponse represents generic RajaOngkir API response
type APIResponse struct {
	RajaOngkir struct {
		Status struct {
			Code    int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
	} `json:"rajaongkir"`
}

// ProvincesResponse represents provinces API response
type ProvincesResponse struct {
	RajaOngkir struct {
		Results []Province `json:"results"`
	} `json:"rajaongkir"`
}

// CitiesResponse represents cities API response
type CitiesResponse struct {
	RajaOngkir struct {
		Results []City `json:"results"`
	} `json:"rajaongkir"`
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/types.go
git commit -m "feat(infrastructure): add RajaOngkir DTO types"
```

---

## Task 16: Create RajaOngkir Cache

**Files:**
- Create: `internal/infrastructure/rajaongkir/cache.go`

**Step 1: Create in-memory cache implementation**

Create file `internal/infrastructure/rajaongkir/cache.go`:

```go
package rajaongkir

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if cache item is expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// Cache provides in-memory caching for RajaOngkir data
type Cache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
	}

	// Start background cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves item from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// Set stores item in cache with TTL
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes item from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear clears all cache items
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// cleanupExpired removes expired items periodically
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Cache key generators
func ProvinceCacheKey() string {
	return "rajaongkir:provinces"
}

func CitiesCacheKey(provinceID string) string {
	return "rajaongkir:cities:" + provinceID
}

func CostCacheKey(origin, destination string, weight int, courier string) string {
	return "rajaongkir:cost:" + origin + ":" + destination + ":" + string(rune(weight)) + ":" + courier
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/cache.go
git commit -m "feat(infrastructure): add RajaOngkir in-memory cache with TTL"
```

---

## Task 17: Create RajaOngkir HTTP Client

**Files:**
- Create: `internal/infrastructure/rajaongkir/client.go`

**Step 1: Create RajaOngkir HTTP client**

Create file `internal/infrastructure/rajaongkir/client.go`:

```go
package rajaongkir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client represents RajaOngkir API client
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	cache      *Cache
}

// NewClient creates a new RajaOngkir client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  os.Getenv("RAJAONGKIR_API_KEY"),
		baseURL: "https://api.rajaongkir.com/starter",
		cache:   NewCache(),
	}
}

// doRequest performs HTTP request to RajaOngkir API
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rajaongkir API error: status %d", resp.StatusCode)
	}

	return resp, nil
}

// GetCache returns cache instance
func (c *Client) GetCache() *Cache {
	return c.cache
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/client.go
git commit -m "feat(infrastructure): add RajaOngkir HTTP client with API key"
```

---

## Task 18: Create RajaOngkir Province Service

**Files:**
- Create: `internal/infrastructure/rajaongkir/province_service.go`

**Step 1: Create province service**

Create file `internal/infrastructure/rajaongkir/province_service.go`:

```go
package rajaongkir

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// GetProvinces retrieves all provinces (with caching)
func (c *Client) GetProvinces(ctx context.Context) ([]Province, error) {
	// Check cache first
	if cached, found := c.cache.Get(ProvinceCacheKey()); found {
		if provinces, ok := cached.([]Province); ok {
			return provinces, nil
		}
	}

	// Fetch from API
	resp, err := c.doRequest(ctx, http.MethodGet, "/province", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provinces: %w", err)
	}
	defer resp.Body.Close()

	var response ProvincesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache for 24 hours
	c.cache.Set(ProvinceCacheKey(), response.RajaOngkir.Results, 24*time.Hour)

	return response.RajaOngkir.Results, nil
}

// GetProvinceByID retrieves a single province by ID
func (c *Client) GetProvinceByID(ctx context.Context, id string) (*Province, error) {
	provinces, err := c.GetProvinces(ctx)
	if err != nil {
		return nil, err
	}

	for _, province := range provinces {
		if province.ID == id {
			return &province, nil
		}
	}

	return nil, fmt.Errorf("province with ID %s not found", id)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/province_service.go
git commit -m "feat(infrastructure): add RajaOngkir province service with 24h cache"
```

---

## Task 19: Create RajaOngkir City Service

**Files:**
- Create: `internal/infrastructure/rajaongkir/city_service.go`

**Step 1: Create city service**

Create file `internal/infrastructure/rajaongkir/city_service.go`:

```go
package rajaongkir

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// GetCities retrieves all cities (with caching)
func (c *Client) GetCities(ctx context.Context) ([]City, error) {
	// Check cache first
	if cached, found := c.cache.Get(ProvinceCacheKey() + ":all"); found {
		if cities, ok := cached.([]City); ok {
			return cities, nil
		}
	}

	// Fetch from API
	resp, err := c.doRequest(ctx, http.MethodGet, "/city", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cities: %w", err)
	}
	defer resp.Body.Close()

	var response CitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache for 24 hours
	c.cache.Set(ProvinceCacheKey()+":all", response.RajaOngkir.Results, 24*time.Hour)

	return response.RajaOngkir.Results, nil
}

// GetCitiesByProvince retrieves cities for a specific province (with caching)
func (c *Client) GetCitiesByProvince(ctx context.Context, provinceID string) ([]City, error) {
	// Check cache first
	cacheKey := CitiesCacheKey(provinceID)
	if cached, found := c.cache.Get(cacheKey); found {
		if cities, ok := cached.([]City); ok {
			return cities, nil
		}
	}

	// Fetch from API
	resp, err := c.doRequest(ctx, http.MethodGet, "/city?province="+provinceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cities: %w", err)
	}
	defer resp.Body.Close()

	var response CitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache for 24 hours
	c.cache.Set(cacheKey, response.RajaOngkir.Results, 24*time.Hour)

	return response.RajaOngkir.Results, nil
}

// GetCityByID retrieves a single city by ID
func (c *Client) GetCityByID(ctx context.Context, id string) (*City, error) {
	cities, err := c.GetCities(ctx)
	if err != nil {
		return nil, err
	}

	for _, city := range cities {
		if city.ID == id {
			return &city, nil
		}
	}

	return nil, fmt.Errorf("city with ID %s not found", id)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/city_service.go
git commit -m "feat(infrastructure): add RajaOngkir city service with 24h cache"
```

---

## Task 20: Create RajaOngkir Cost Service

**Files:**
- Create: `internal/infrastructure/rajaongkir/cost_service.go`

**Step 1: Create cost calculation service**

Create file `internal/infrastructure/rajaongkir/cost_service.go`:

```go
package rajaongkir

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ShippingCost represents formatted shipping cost result
type ShippingCost struct {
	CourierCode string
	CourierName string
	Service     string
	Description string
	Cost        float64
	ETD         string
}

// CalculateCost calculates shipping costs (with caching)
func (c *Client) CalculateCost(ctx context.Context, req CostRequest) ([]ShippingCost, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("rajaongkir:cost:%s:%s:%d:%s", req.Origin, req.Destination, req.Weight, req.Courier)
	if cached, found := c.cache.Get(cacheKey); found {
		if costs, ok := cached.([]ShippingCost); ok {
			return costs, nil
		}
	}

	// Fetch from API
	resp, err := c.doRequest(ctx, http.MethodPost, "/cost", req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}
	defer resp.Body.Close()

	var response CostResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Format results
	results := make([]ShippingCost, 0)
	for _, result := range response.RajaOngkir.Results {
		for _, cost := range result.Costs {
			for _, detail := range cost.Cost {
				results = append(results, ShippingCost{
					CourierCode: result.Code,
					CourierName: result.Name,
					Service:     cost.Service,
					Description: cost.Description,
					Cost:        float64(detail.Value),
					ETD:         detail.ETD,
				})
			}
		}
	}

	// Cache for 1 hour
	c.cache.Set(cacheKey, results, 1*time.Hour)

	return results, nil
}

// CalculateCostMultipleCouriers calculates shipping costs for multiple couriers
func (c *Client) CalculateCostMultipleCouriers(ctx context.Context, req CostRequest, couriers []string) ([]ShippingCost, error) {
	var allResults []ShippingCost

	for _, courier := range couriers {
		req.Courier = courier
		results, err := c.CalculateCost(ctx, req)
		if err != nil {
			// Log error but continue with other couriers
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// ParseCouriers parses colon-separated courier string (e.g., "jne:jnt:sicepat")
func ParseCouriers(courierStr string) []string {
	if courierStr == "" {
		return []string{"jne"} // Default to JNE
	}

	couriers := strings.Split(courierStr, ":")
	// Validate courier codes
	validCouriers := map[string]bool{
		"jne":       true,
		"jnt":       true,
		"sicepat":   true,
		"anteraja":  true,
		"pos":       true,
	}

	var result []string
	for _, courier := range couriers {
		if validCouriers[strings.ToLower(courier)] {
			result = append(result, strings.ToLower(courier))
		}
	}

	if len(result) == 0 {
		return []string{"jne"}
	}

	return result
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/infrastructure/rajaongkir`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/cost_service.go
git commit -m "feat(infrastructure): add RajaOngkir cost service with 1h cache"
```

---

## Task 21: Create RajaOngkir Client Tests

**Files:**
- Create: `internal/infrastructure/rajaongkir/client_test.go`

**Step 1: Write failing tests for RajaOngkir client**

Create file `internal/infrastructure/rajaongkir/client_test.go`:

```go
package rajaongkir

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.cache == nil {
		t.Error("cache is nil")
	}
}

func TestCache(t *testing.T) {
	cache := NewCache()
	ctx := context.Background()

	// Test Set and Get
	testData := []Province{
		{ID: "1", Name: "Test Province"},
	}

	cache.Set(ProvinceCacheKey(), testData, 1*time.Hour)

	// Verify Get works
	cached, found := cache.Get(ProvinceCacheKey())
	if !found {
		t.Fatal("Cache Get returned not found")
	}

	provinces, ok := cached.([]Province)
	if !ok {
		t.Fatal("Cache data type mismatch")
	}

	if len(provinces) != 1 {
		t.Errorf("Expected 1 province, got %d", len(provinces))
	}

	// Test expiration
	cache.Set("expired_key", testData, 1*time.Nanosecond)
	time.Sleep(10 * time.Millisecond)

	_, found = cache.Get("expired_key")
	if found {
		t.Error("Expired cache item should not be found")
	}
}

func TestCacheConcurrent(t *testing.T) {
	cache := NewCache()
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(n int) {
			cache.Set("key", n, 1*time.Hour)
			cache.Get("key")
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// If we got here without panic, test passed
}

func TestParseCouriers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single courier",
			input:    "jne",
			expected: []string{"jne"},
		},
		{
			name:     "multiple couriers",
			input:    "jne:jnt:sicepat",
			expected: []string{"jne", "jnt", "sicepat"},
		},
		{
			name:     "empty string defaults to jne",
			input:    "",
			expected: []string{"jne"},
		},
		{
			name:     "invalid couriers filtered out",
			input:    "jne:invalid:jnt",
			expected: []string{"jne", "jnt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCouriers(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d couriers, got %d", len(tt.expected), len(result))
			}

			for i, courier := range result {
				if courier != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, courier)
				}
			}
		})
	}
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./internal/infrastructure/rajaongkir -v`

Expected: All tests PASS

**Step 3: Commit**

```bash
git add internal/infrastructure/rajaongkir/client_test.go
git commit -m "test(infrastructure): add RajaOngkir client unit tests"
```

---

## Task 22: Create Repository Ports - Address

**Files:**
- Modify: `internal/domain/ports/repository.go`

**Step 1: Add Address repository port**

Append to `internal/domain/ports/repository.go`:

```go
// AddressRepository defines the interface for address persistence
type AddressRepository interface {
	// Create creates a new address
	Create(ctx context.Context, address *entities.Address) error

	// FindByID retrieves address by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Address, error)

	// FindByUserID retrieves all addresses for a user (excluding deleted)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Address, error)

	// FindDefaultByUserID retrieves user's default address
	FindDefaultByUserID(ctx context.Context, userID uuid.UUID) (*entities.Address, error)

	// Update updates an existing address
	Update(ctx context.Context, address *entities.Address) error

	// SoftDelete soft deletes an address (sets is_deleted = true)
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// SetDefault sets address as default and unsets others
	SetDefault(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) error

	// CountActiveOrders checks if address is used in active orders
	CountActiveOrders(ctx context.Context, addressID uuid.UUID) (int, error)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/domain/ports`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/domain/ports/repository.go
git commit -m "feat(ports): add Address repository port"
```

---

## Task 23: Create Repository Ports - Order, OrderItem, Payment

**Files:**
- Modify: `internal/domain/ports/repository.go`

**Step 1: Add Order, OrderItem, Payment repository ports**

Append to `internal/domain/ports/repository.go`:

```go
// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *entities.Order) error

	// FindByID retrieves order by ID with relations
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)

	// FindByOrderNumber retrieves order by order number
	FindByOrderNumber(ctx context.Context, orderNumber string) (*entities.Order, error)

	// FindByUserID retrieves orders for a user with pagination
	FindByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*entities.Order, error)

	// CountByUserID counts orders for a user
	CountByUserID(ctx context.Context, userID uuid.UUID, status string) (int, error)

	// Update updates an existing order
	Update(ctx context.Context, order *entities.Order) error

	// Delete deletes an order
	Delete(ctx context.Context, id uuid.UUID) error

	// CountOrdersByDate counts orders created on a specific date (for order number generation)
	CountOrdersByDate(ctx context.Context, date string) (int, error)
}

// OrderItemRepository defines the interface for order item persistence
type OrderItemRepository interface {
	// Create creates a new order item
	Create(ctx context.Context, item *entities.OrderItem) error

	// FindByOrderID retrieves all items for an order
	FindByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entities.OrderItem, error)

	// FindByID retrieves order item by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.OrderItem, error)

	// DeleteByOrderID deletes all items for an order
	DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error
}

// PaymentRepository defines the interface for payment persistence
type PaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *entities.Payment) error

	// FindByID retrieves payment by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)

	// FindByOrderID retrieves payment by order ID
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*entities.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *entities.Payment) error

	// FindExpiredPayments retrieves payments that have expired
	FindExpiredPayments(ctx context.Context, expiryTime time.Time) ([]*entities.Payment, error)

	// FindPendingPayments retrieves pending payments older than specified duration
	FindPendingPayments(ctx context.Context, olderThan time.Time) ([]*entities.Payment, error)
}
```

**Step 4: Verify compilation**

Run: `go build ./internal/domain/ports`
Expected: No errors

**Step 5: Commit**

```bash
git add internal/domain/ports/repository.go
git commit -m "feat(ports): add Order, OrderItem, Payment repository ports"
```

---

## Task 24: Implement Address Repository

**Files:**
- Create: `internal/adapters/persistence/address_repository.go`

**Step 1: Write failing test for Address repository**

Create file `internal/adapters/persistence/address_repository_test.go`:

```go
package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
)

func TestAddressRepository_Create(t *testing.T) {
	// This test requires a test database
	// Skip if not available
	t.Skip("requires test database")

	repo := NewAddressRepository(client)
	ctx := context.Background()

	address := &entities.Address{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Label:         "Home",
		RecipientName: "John Doe",
		Phone:         "08123456789",
		AddressLine1:  "Jl. Test No. 123",
		ProvinceID:    "1",
		ProvinceName:  "DKI Jakarta",
		CityID:        "17",
		CityName:      "Jakarta Selatan",
		PostalCode:    "12345",
		IsDefault:     false,
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, address)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, address.ID)
}

func TestAddressRepository_FindByUserID(t *testing.T) {
	t.Skip("requires test database")
	// TODO: Implement test
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/persistence -v -run TestAddressRepository_Create`

Expected: FAIL with undefined: NewAddressRepository

**Step 3: Implement Address repository**

Create file `internal/adapters/persistence/address_repository.go`:

```go
package persistence

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/address"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
)

// AddressRepository implements ports.AddressRepository interface
type AddressRepository struct {
	client *ent.Client
}

// NewAddressRepository creates a new Address repository
func NewAddressRepository(client *ent.Client) *AddressRepository {
	return &AddressRepository{
		client: client,
	}
}

// Create creates a new address
func (r *AddressRepository) Create(ctx context.Context, addr *entities.Address) error {
	// If setting as default, unset other defaults
	if addr.IsDefault {
		_, err := r.client.Address.
			Update().
			Where(address.UserID(addr.UserID), address.IsDefault(true)).
			SetIsDefault(false).
			Exec(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return fmt.Errorf("failed to unset existing default: %w", err)
		}
	}

	// Create address
	create := r.client.Address.Create().
		SetID(addr.ID).
		SetUserID(addr.UserID).
		SetLabel(addr.Label).
		SetRecipientName(addr.RecipientName).
		SetPhone(addr.Phone).
		SetAddressLine1(addr.AddressLine1).
		SetProvinceID(addr.ProvinceID).
		SetProvinceName(addr.ProvinceName).
		SetCityID(addr.CityID).
		SetCityName(addr.CityName).
		SetPostalCode(addr.PostalCode).
		SetIsDefault(addr.IsDefault).
		SetIsDeleted(false).
		SetNillableAddressLine2(addr.AddressLine2).
		SetNillableDistrict(addr.District)

	err := create.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create address: %w", err)
	}

	return nil
}

// FindByID retrieves address by ID
func (r *AddressRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	entAddr, err := r.client.Address.
		Query().
		Where(address.ID(id), address.IsDeleted(false)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to find address: %w", err)
	}

	return r.toEntity(entAddr), nil
}

// FindByUserID retrieves all addresses for a user (excluding deleted)
func (r *AddressRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Address, error) {
	entAddrs, err := r.client.Address.
		Query().
		Where(address.UserID(userID), address.IsDeleted(false)).
		Order(address.ByIsDefault(sql.OrderDesc())).
		Order(address.ByCreatedAt(sql.OrderDesc())).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find addresses: %w", err)
	}

	addresses := make([]*entities.Address, len(entAddrs))
	for i, entAddr := range entAddrs {
		addresses[i] = r.toEntity(entAddr)
	}

	return addresses, nil
}

// FindDefaultByUserID retrieves user's default address
func (r *AddressRepository) FindDefaultByUserID(ctx context.Context, userID uuid.UUID) (*entities.Address, error) {
	entAddr, err := r.client.Address.
		Query().
		Where(address.UserID(userID), address.IsDefault(true), address.IsDeleted(false)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to find default address: %w", err)
	}

	return r.toEntity(entAddr), nil
}

// Update updates an existing address
func (r *AddressRepository) Update(ctx context.Context, addr *entities.Address) error {
	update := r.client.Address.UpdateOneID(addr.ID).
		SetLabel(addr.Label).
		SetRecipientName(addr.RecipientName).
		SetPhone(addr.Phone).
		SetAddressLine1(addr.AddressLine1).
		SetProvinceID(addr.ProvinceID).
		SetProvinceName(addr.ProvinceName).
		SetCityID(addr.CityID).
		SetCityName(addr.CityName).
		SetPostalCode(addr.PostalCode).
		SetNillableAddressLine2(addr.AddressLine2).
		SetNillableDistrict(addr.District)

	// If setting as default, unset others
	if addr.IsDefault {
		_, err := r.client.Address.
			Update().
			Where(address.UserID(addr.UserID), address.IsDefault(true), address.IDNEQ(addr.ID)).
			SetIsDefault(false).
			Exec(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return fmt.Errorf("failed to unset existing default: %w", err)
		}
		update.SetIsDefault(true)
	}

	err := update.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.ErrAddressNotFound
		}
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

// SoftDelete soft deletes an address (sets is_deleted = true)
func (r *AddressRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Address.UpdateOneID(id).
		SetIsDeleted(true).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.ErrAddressNotFound
		}
		return fmt.Errorf("failed to soft delete address: %w", err)
	}

	return nil
}

// SetDefault sets address as default and unsets others
func (r *AddressRepository) SetDefault(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) error {
	// Unset existing defaults
	_, err := r.client.Address.
		Update().
		Where(address.UserID(userID), address.IsDefault(true)).
		SetIsDefault(false).
		Exec(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("failed to unset existing default: %w", err)
	}

	// Set new default
	err = r.client.Address.UpdateOneID(addressID).
		SetIsDefault(true).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.ErrAddressNotFound
		}
		return fmt.Errorf("failed to set default: %w", err)
	}

	return nil
}

// CountActiveOrders checks if address is used in active orders
func (r *AddressRepository) CountActiveOrders(ctx context.Context, addressID uuid.UUID) (int, error) {
	// TODO: Implement after Order schema is created
	return 0, nil
}

// toEntity converts Ent address to domain entity
func (r *AddressRepository) toEntity(entAddr *ent.Address) *entities.Address {
	return &entities.Address{
		ID:               entAddr.ID,
		UserID:           entAddr.UserID,
		Label:            entAddr.Label,
		RecipientName:    entAddr.RecipientName,
		Phone:            entAddr.Phone,
		AddressLine1:     entAddr.AddressLine1,
		AddressLine2:     entAddr.AddressLine2,
		ProvinceID:       entAddr.ProvinceID,
		ProvinceName:     entAddr.ProvinceName,
		CityID:           entAddr.CityID,
		CityName:         entAddr.CityName,
		District:         entAddr.District,
		PostalCode:       entAddr.PostalCode,
		IsDefault:        entAddr.IsDefault,
		IsDeleted:        entAddr.IsDeleted,
		CreatedAt:        entAddr.CreatedAt,
		UpdatedAt:        entAddr.UpdatedAt,
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapters/persistence -v -run TestAddressRepository`

Expected: Tests PASS (or SKIP if no test database)

**Step 5: Commit**

```bash
git add internal/adapters/persistence/address_repository.go
git add internal/adapters/persistence/address_repository_test.go
git commit -m "feat(persistence): implement Address repository"
```

---

**[DOCUMENT CONTINUES IN NEXT PART...]**

**Note:** This is a massive implementation plan. I'll create the complete plan with all 100+ tasks. Due to length, I'm structuring this as bite-sized tasks. Would you like me to:

1. Continue writing the complete plan (all 100+ tasks) to the file first
2. Or start executing what we have so far using the executing-plans skill?

The plan currently covers up through Task 24 (Address Repository). The remaining tasks include:
- Order, OrderItem, Payment repositories
- Midtrans integration
- All services (Address, Shipping, Order, Payment, Email)
- All handlers and DTOs
- Cron jobs
- Tests

Which approach would you prefer?
