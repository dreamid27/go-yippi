# Phase 3 Backend Design - Checkout & Payment System

**Date:** 2026-01-01
**Project:** go-yippi (Backend)
**Phase:** 3 - Checkout & Payment (Week 5-6)
**Status:** Design Document

---

## Executive Summary

This document details technical design for Phase 3 backend implementation, covering:
- **Address Management** system with multiple addresses per user
- **Shipping Integration** with RajaOngkir API (provinces, cities, courier costs)
- **Order Management** with complete order lifecycle (pending → completed)
- **Payment Integration** with Midtrans SDK (multiple payment methods)
- **Order Status Workflow** with automated transitions
- **Invoice Generation** for completed orders

**Key Design Decisions:**
1. ✅ **RajaOngkir caching** - Cache provinces/cities in-memory to reduce API calls and rate limit issues
2. ✅ **Order calculations in service layer** - Subtotal, tax (11%), shipping, discount, total calculated consistently
3. ✅ **Price snapshots in orders** - Lock prices at order creation (handle future price changes)
4. ✅ **Midtrans Snap integration** - Use Snap popup for secure payment handling
5. ✅ **Payment webhook verification** - Verify webhook signatures to prevent fraud
6. ✅ **Auto-cancel expired payments** - Cron job cancels unpaid orders after 24 hours
7. ✅ **Soft delete for addresses** - Keep historical data with `is_deleted` flag

---

## Database Schema Design

### 1. Address Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/address.go`

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

type Address struct {
    ent.Schema
}

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

func (Address) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id"),
        index.Fields("user_id", "is_default").Unique(), // Only one default per user
    }
}
```

**Key Design Points:**
- **Soft delete** - `is_deleted` flag preserves historical data (orders reference addresses)
- **One default per user** - Unique constraint on `(user_id, is_default = true)`
- **Cache RajaOngkir data** - Store `province_name`, `city_name` to avoid repeated API calls
- **Indonesian phone format** - Should validate for `+62` or `0` prefix
- **Multiple addresses per user** - No limit on number of addresses

**Validation Rules:**
- `phone`: Must be valid Indonesian phone number (8-15 digits, starts with `0` or `+62`)
- `postal_code`: 5 digits (Indonesian format)
- `recipient_name`: 2-100 characters
- `address_line1`: 5-500 characters
- `label`: 2-50 characters

---

### 2. Order Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/order.go`

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

type Order struct {
    ent.Schema
}

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

**Key Design Points:**
- **Order number format** - `ORD-YYYYMMDD-XXXXX` (human-readable, chronological)
- **Status workflow** - Linear progression with branches:
  - `pending_payment` → `payment_verified` → `processing` → `shipped` → `delivered` → `completed`
  - Can cancel from `pending_payment` only (before payment)
- **Price snapshots** - `subtotal`, `discount`, `tax`, `shipping_cost`, `total` stored (immutable after creation)
- **Courier tracking** - `courier`, `courier_service`, `tracking_number` for RajaOngkir integration
- **Tax calculation** - Indonesia PPN 11% applied to subtotal (not shipping or discount)
- **Metadata** - JSONB field for extensibility (source, device, utm parameters)

**Status Transitions:**
```go
// Valid status transitions
pending_payment → payment_verified (after webhook confirms payment)
payment_verified → processing (order confirmed, stock allocated)
processing → shipped (courier picked up, tracking number added)
shipped → delivered (courier confirmed delivery)
delivered → completed (order finalized, 7-day review period over)
pending_payment → cancelled (user cancelled or expired)
```

---

### 3. OrderItem Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/order_item.go`

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

type OrderItem struct {
    ent.Schema
}

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

func (OrderItem) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("order_id"),
        index.Fields("product_id"),
        index.Fields("product_variant_id"),
        index.Fields("order_id", "product_variant_id"),
    }
}
```

**Key Design Points:**
- **Price snapshots** - Store `price`, `subtotal`, `product_name`, `product_variant_attributes` at order time
- **Preserve historical data** - Even if product/variant is deleted later, order items keep data
- **Image snapshot** - Store `product_image_url` so order history still shows product images
- **SKU snapshot** - Store `sku` to identify variant even if renamed later
- **No cascading delete** - Product/Variant deletion should NOT delete OrderItem (soft delete)

---

### 4. Payment Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/payment.go`

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

type Payment struct {
    ent.Schema
}

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

func (Payment) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("order", Order.Type).
            Ref("payment").
            Unique().
            Required().
            Field("order_id"),
    }
}

func (Payment) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("order_id").Unique(),
        index.Fields("midtrans_transaction_id"),
        index.Fields("status"),
        index.Fields("expires_at"),
    }
}
```

**Key Design Points:**
- **One-to-one with Order** - Unique constraint on `order_id`
- **Payment methods** - Supported: VA (4 banks), E-wallets (4), QRIS, Credit Card, COD
- **Status workflow** - Tracks payment status from Midtrans:
  - `pending` → `processing` → `success` (paid)
  - `pending` → `failed` / `expired` / `cancelled` (unpaid)
- **VA details** - Store `va_number`, `va_bank` for manual VA payments
- **Midtrans integration** - Store `midtrans_transaction_id`, `payment_url`, `raw_midtrans_response`
- **Expiration** - `expires_at` set to 24 hours after payment initiation (auto-cancel via cron)
- **Debugging** - `raw_midtrans_response` stores complete Midtrans webhook payload

---

### 5. User Schema Update

**File:** `internal/adapters/persistence/db/schema/user.go`

**Add these edges:**

```go
func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("refresh_tokens", RefreshToken.Type),
        edge.To("cart", Cart.Type).Unique(),

        // NEW: Address relationship
        edge.To("addresses", Address.Type),

        // NEW: Orders relationship
        edge.To("orders", Order.Type),
    }
}
```

---

### 6. Product Schema Update

**File:** `internal/adapters/persistence/db/schema/product.go`

**Add this edge:**

```go
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("category", Category.Type),
        edge.To("brand", Brand.Type),
        edge.To("variants", ProductVariant.Type),

        // NEW: Order items relationship (historical data)
        edge.To("order_items", OrderItem.Type),
    }
}
```

---

### 7. ProductVariant Schema Update

**File:** `internal/adapters/persistence/db/schema/product_variant.go`

**Add this edge:**

```go
func (ProductVariant) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("product", Product.Type).
            Ref("variants").
            Unique().
            Required().
            Field("product_id"),

        edge.To("cart_items", CartItem.Type),

        // NEW: Order items relationship (historical data)
        edge.To("order_items", OrderItem.Type),
    }
}
```

---

## RajaOngkir Integration

### Overview

RajaOngkir API provides shipping cost calculation for Indonesian couriers (JNE, J&T, SiCepat, Anteraja). Integration requires:
- **API Key** (obtained from RajaOngkir dashboard)
- **Account Tier** (Starter, Basic, Pro - affects rate limits)
- **Webhook handling** (optional, for tracking updates)

### Architecture Design

**File Structure:**
```
internal/
├── infrastructure/
│   └── rajaongkir/
│       ├── client.go          # HTTP client wrapper
│       ├── province_service.go # Province API operations
│       ├── city_service.go     # City API operations
│       ├── cost_service.go     # Shipping cost calculation
│       ├── cache.go            # In-memory cache layer
│       └── types.go           # RajaOngkir DTOs
├── application/
│   └── services/
│       └── shipping_service.go # Shipping business logic
```

### Caching Strategy

**Problem:** RajaOngkir has strict rate limits (Starter: 100 requests/day). Provinces/cities rarely change, so fetching on every request is wasteful.

**Solution:** In-memory caching only (no Redis):
1. **In-memory cache** (sync.Map) for provinces/cities/shipping costs - Fast, no external dependencies
2. **App-level cache** - Shared across all requests within single instance

**Cache Keys:**
```
rajaongkir:provinces              # List of all provinces
rajaongkir:cities:{province_id}    # Cities for specific province
rajaongkir:cost:{origin_city}:{dest_city}:{weight}  # Shipping cost calculation
```

**Cache TTL:**
- Provinces: 24 hours (never changes)
- Cities: 24 hours (rarely changes)
- Shipping costs: 1 hour (prices change occasionally)

---

## API Endpoints

### 1. Address APIs (NEW)

#### POST /api/addresses

**Authentication:** Required (JWT)

**Description:** Create new shipping address for current user

**Request Body:**
```json
{
  "label": "Home",
  "recipient_name": "John Doe",
  "phone": "08123456789",
  "address_line1": "Jl. Sudirman No. 123",
  "address_line2": "Apt. ABC, Lt. 5",
  "province_id": "1",
  "province_name": "DKI Jakarta",
  "city_id": "17",
  "city_name": "Jakarta Selatan",
  "district": "Kebayoran Baru",
  "postal_code": "12190",
  "is_default": false
}
```

**Response:** Created address (201)

**Business Logic:**
- Validate phone number format (Indonesian: 8-15 digits, starts with `0` or `+62`)
- Validate postal code format (5 digits)
- If `is_default = true`, set all other user's addresses `is_default = false`
- Auto-populate `province_name`, `city_name` from RajaOngkir (optional: validate IDs)

**Error Responses:**
- 400: Validation failed (phone, postal_code, required fields)
- 401: Unauthorized
- 404: Province/City ID not found in RajaOngkir

---

#### GET /api/addresses

**Authentication:** Required (JWT)

**Description:** Get all addresses for current user (excluding soft-deleted)

**Response:** List of addresses (200)

**Business Logic:**
- Filter `user_id = current_user.id`
- Filter `is_deleted = false`
- Sort by `is_default DESC` (default address first)
- Then sort by `created_at DESC`

---

#### GET /api/addresses/:id

**Authentication:** Required (JWT)

**Description:** Get specific address by ID

**Response:** Address details (200)

**Business Logic:**
- Validate address belongs to current user
- Return 404 if not found or soft-deleted

---

#### PUT /api/addresses/:id

**Authentication:** Required (JWT)

**Description:** Update existing address

**Request Body:** Same as POST (all fields optional)

**Response:** Updated address (200)

**Business Logic:**
- Validate address belongs to current user
- If updating `is_default = true`, set all other addresses `is_default = false`
- Prevent updating `user_id`, `created_at`, `is_deleted`
- Validate phone, postal_code format

---

#### DELETE /api/addresses/:id

**Authentication:** Required (JWT)

**Description:** Soft delete address (set `is_deleted = true`)

**Response:** 204 No Content

**Business Logic:**
- Validate address belongs to current user
- Soft delete: Set `is_deleted = true`
- If deleting default address, set another address as default (if exists)
- Prevent deletion if address is referenced by active orders

**Error Responses:**
- 400: Cannot delete - address used in active orders
- 404: Address not found
- 403: Address doesn't belong to user

---

#### PUT /api/addresses/:id/default

**Authentication:** Required (JWT)

**Description:** Set address as default shipping address

**Response:** Updated address (200)

**Business Logic:**
- Validate address belongs to current user
- Set all user's addresses `is_default = false`
- Set target address `is_default = true`

---

### 2. Shipping APIs (NEW)

#### GET /api/shipping/provinces

**Description:** Get all Indonesian provinces (public, no auth required)

**Response:** List of provinces (200)

**Business Logic:**
- Fetch from RajaOngkir API (with caching - 24h TTL)
- Return cached data if available
- Include `cached: true` in response if served from cache

---

#### GET /api/shipping/provinces/:id

**Description:** Get province by ID (public, no auth required)

**Response:** Province details (200)

---

#### GET /api/shipping/cities

**Query Parameters:**
- `province_id` (required) - Filter cities by province

**Description:** Get cities for specific province (public, no auth required)

**Response:** List of cities (200)

**Business Logic:**
- Fetch from RajaOngkir API (with caching - 24h TTL)
- Filter by `province_id` if provided
- Return cached data if available

---

#### GET /api/shipping/cities/:id

**Description:** Get city by ID (public, no auth required)

**Response:** City details (200)

---

#### POST /api/shipping/cost

**Authentication:** Optional (required only for user-specific origin addresses)

**Description:** Calculate shipping costs from multiple couriers

**Request Body:**
```json
{
  "origin_city_id": "17",
  "destination_city_id": "153",
  "weight": 500,
  "courier": "jne:jnt:sicepat:anteraja"
}
```

**Alternative Request (user's default address as origin):**
```json
{
  "use_default_address": true,
  "destination_city_id": "153",
  "weight": 500,
  "courier": "jne:jnt:sicepat:anteraja"
}
```

**Response:** List of shipping options (200)

**Business Logic:**
- Validate city IDs exist in RajaOngkir
- Fetch from RajaOngkir cost API (with caching - 1h TTL)
- Support multiple couriers in single request (colon-separated)
- If `use_default_address = true`, use current user's default address `city_id` as origin
- Include `cached: true` if served from cache

**Error Responses:**
- 400: Invalid city IDs, weight <= 0, invalid courier code
- 404: City not found

---

### 3. Order APIs (NEW)

#### POST /api/orders

**Authentication:** Required (JWT)

**Description:** Create new order from cart

**Request Body:**
```json
{
  "shipping_address_id": "address-uuid",
  "courier": "jne",
  "courier_service": "REG",
  "shipping_cost": 32000,
  "voucher_code": "SAVE10",
  "notes": "Please deliver between 9 AM - 5 PM"
}
```

**Response:** Created order with payment (201)

**Business Logic:**
1. **Validate user's cart:**
   - Fetch current user's cart with all items
   - Validate cart is not empty
   - Validate stock availability for all items (use StockValidatorService)

2. **Validate shipping address:**
   - Validate address belongs to current user
   - Validate address is not soft-deleted

3. **Validate shipping cost:**
   - Calculate shipping cost from RajaOngkir (origin: default warehouse, destination: address)
   - Validate provided `shipping_cost` matches calculated cost (within tolerance)

4. **Calculate order totals:**
   - Subtotal: Sum of `cart_item.price_snapshot * quantity`
   - Discount: Apply voucher discount (if valid)
   - Tax: 11% of (subtotal - discount)
   - Total: Subtotal - discount + tax + shipping_cost

5. **Create order:**
   - Generate unique `order_number`: `ORD-YYYYMMDD-XXXXX` (auto-increment per day)
   - Set status = `pending_payment`
   - Create Order record

6. **Create order items:**
   - For each cart item:
     - Copy product data snapshots (name, attributes, image, sku, price)
     - Create OrderItem records

7. **Create payment:**
   - Create Payment record (status = `pending`)
   - Call Midtrans API to create payment
   - Store `payment_url` (Snap popup URL)
   - Set `expires_at` = now + 24 hours

8. **Stock handling:**
   - NO stock deduction yet (will deduct after payment verified)
   - Reserve stock (optional: add `reserved_quantity` field to variant)

9. **Clear cart:**
   - Clear all items from user's cart after successful order creation

10. **Send confirmation email:**
    - Email: "Order Created - Waiting for Payment"
    - Include order details, payment instructions, expires_at

**Error Responses:**
- 400: Cart empty, stock validation failed, shipping cost mismatch, invalid address
- 401: Unauthorized
- 404: Address not found
- 409: Voucher code invalid/expired

---

#### GET /api/orders

**Authentication:** Required (JWT)

**Query Parameters:**
- `status` (optional) - Filter by order status
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)

**Description:** Get current user's orders with pagination

**Response:** List of orders (200)

**Business Logic:**
- Filter `user_id = current_user.id`
- Sort by `created_at DESC` (newest first)
- Apply pagination
- Apply status filter if provided

---

#### GET /api/orders/:id

**Authentication:** Required (JWT)

**Description:** Get order details by ID

**Response:** Full order details (200)

**Business Logic:**
- Validate order belongs to current user
- Return 404 if not found
- Include all details (items, payment, shipping address)

---

#### POST /api/orders/:id/cancel

**Authentication:** Required (JWT)

**Request Body:**
```json
{
  "reason": "Changed my mind"
}
```

**Response:** Updated order (200)

**Business Logic:**
- Validate order belongs to current user
- Validate order status = `pending_payment` (can only cancel before payment)
- Set status = `cancelled`
- Set `cancelled_at` = now
- Set `cancellation_reason` = provided reason
- Cancel pending Midtrans payment (if exists)
- Send cancellation email
- Return error if status is not `pending_payment`

**Error Responses:**
- 400: Cannot cancel - order already paid/processed
- 404: Order not found
- 403: Order doesn't belong to user

---

### 4. Payment APIs (NEW)

#### POST /api/payments/:order_id

**Authentication:** Required (JWT)

**Description:** Initialize payment for existing order (create Midtrans transaction)

**Request Body:**
```json
{
  "method": "credit_card"
}
```

**Response:** Payment initialization response (201)

**Business Logic:**
- Validate order belongs to current user
- Validate order status = `pending_payment`
- If payment already exists → Return existing payment URL
- Create/update Payment record
- Call Midtrans API to create transaction
- Return `payment_url` (Snap popup URL)
- Set `expires_at` = now + 24 hours

---

#### GET /api/payments/:id/status

**Authentication:** Required (JWT)

**Description:** Check payment status (polling endpoint for frontend)

**Response:** Payment status (200)

**Business Logic:**
- Fetch payment by ID
- Validate payment belongs to current user (via order_id)
- Return payment status and details

---

#### POST /api/payments/webhook

**Authentication:** Midtrans signature verification (no JWT)

**Description:** Midtrans webhook handler - payment status notifications

**Request Headers:**
- `X-Callback-Token: YOUR_MIDTRANS_SERVER_KEY`

**Response:** 200 OK (must return 200, even if processing fails)

**Business Logic:**
1. **Verify signature:**
   - Calculate expected signature: `SHA512(order_id + status_code + gross_amount + server_key)`
   - Compare with `signature_key` from request
   - Return 400 if signature invalid

2. **Find order by `order_number`:**
   - Parse `order_id` from webhook
   - Fetch order from database

3. **Process payment status:**

   **If `transaction_status = settlement` / `capture`:**
   - Update Payment status = `success`
   - Update Order status = `payment_verified`
   - Set `paid_at` = now
   - DEDUCT STOCK from variants (order items)
   - Send "Payment Success" email
   - Update Order status = `processing`
   - Send "Order Confirmed" email to warehouse

   **If `transaction_status = pending`:**
   - Update Payment status = `processing`
   - Do nothing else (waiting for user to complete payment)

   **If `transaction_status = deny` / `cancel` / `expire`:**
   - Update Payment status = `failed` / `expired` / `cancelled`
   - Set `failure_reason` = status_message
   - Update Order status = `cancelled`
   - Set `cancelled_at` = now
   - Send "Payment Failed" email with retry instructions

4. **Store raw webhook:**
   - Save complete webhook payload in `raw_midtrans_response` for debugging

5. **Always return 200:**
   - Even if processing fails (Midtrans will retry if not 200)
   - Log errors internally

**Security:**
- Verify signature key (CRITICAL - prevents fake webhooks)
- Rate limit webhook endpoint (prevent spam)
- Log all webhook requests for audit

---

## Business Logic & Service Layer

### Order Service

**File:** `internal/application/services/order_service.go`

```go
type OrderService struct {
    orderRepo       ports.OrderRepository
    orderItemRepo   ports.OrderItemRepository
    cartRepo        ports.CartRepository
    productVariantRepo ports.ProductVariantRepository
    addressRepo     ports.AddressRepository
    paymentService  *PaymentService
    shippingService *ShippingService
    stockValidator  *StockValidatorService
    priceCalculator *PriceCalculatorService
    emailService    *EmailService
    logger          *zap.Logger
}

// CreateOrder creates new order from user's cart
func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*entities.Order, error) {
    // 1. Fetch and validate user's cart
    cart, err := s.cartRepo.GetCartWithItems(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch cart: %w", err)
    }

    if len(cart.Items) == 0 {
        return nil, ErrCartEmpty
    }

    // 2. Validate stock availability for all cart items
    stockErrors, err := s.stockValidator.ValidateCartStock(ctx, cart.ID)
    if err != nil {
        return nil, fmt.Errorf("stock validation failed: %w", err)
    }
    if len(stockErrors) > 0 {
        return nil, ErrInsufficientStock{Errors: stockErrors}
    }

    // 3. Validate shipping address
    address, err := s.addressRepo.FindByID(ctx, req.ShippingAddressID)
    if err != nil {
        return nil, ErrAddressNotFound
    }
    if address.UserID != userID {
        return nil, ErrAddressNotOwned
    }
    if address.IsDeleted {
        return nil, ErrAddressDeleted
    }

    // 4. Calculate shipping cost and validate
    warehouseCityID := s.shippingService.GetWarehouseCityID() // e.g., Jakarta
    calculatedCosts, err := s.shippingService.CalculateCost(ctx, ShippingCostRequest{
        OriginCityID:      warehouseCityID,
        DestinationCityID: address.CityID,
        Weight:           cart.TotalWeight,
        Courier:          req.Courier,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to calculate shipping cost: %w", err)
    }

    var selectedCost float64
    found := false
    for _, cost := range calculatedCosts {
        if cost.CourierCode == req.Courier && cost.Service == req.CourierService {
            selectedCost = cost.Cost
            found = true
            break
        }
    }
    if !found {
        return nil, ErrShippingCostMismatch
    }

    // Validate provided shipping cost matches calculated (within 1% tolerance)
    tolerance := selectedCost * 0.01
    if math.Abs(req.ShippingCost-selectedCost) > tolerance {
        return nil, ErrShippingCostMismatch{
            Provided: req.ShippingCost,
            Calculated: selectedCost,
        }
    }

    // 5. Calculate order totals
    subtotal := s.priceCalculator.CalculateCartSubtotal(cart.Items)

    // Apply voucher discount if provided
    var discount float64
    if req.VoucherCode != "" {
        discount, err = s.applyVoucher(ctx, userID, req.VoucherCode, subtotal)
        if err != nil {
            return nil, err
        }
    }

    tax := s.calculateTax(subtotal - discount) // 11% PPN
    total := subtotal - discount + tax + req.ShippingCost

    // 6. Generate order number
    orderNumber, err := s.generateOrderNumber(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate order number: %w", err)
    }

    // 7. Create order record
    order := &entities.Order{
        ID:              uuid.New(),
        UserID:          userID,
        ShippingAddressID: req.ShippingAddressID,
        OrderNumber:      orderNumber,
        Status:          "pending_payment",
        Courier:         req.Courier,
        CourierService:   req.CourierService,
        ShippingCost:    req.ShippingCost,
        Subtotal:        subtotal,
        Discount:        discount,
        Tax:             tax,
        Total:           total,
        Notes:           req.Notes,
        Metadata:        req.Metadata,
    }

    if err := s.orderRepo.Create(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    // 8. Create order items (with price snapshots)
    for _, cartItem := range cart.Items {
        variant, err := s.productVariantRepo.FindByID(ctx, cartItem.ProductVariantID)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch variant: %w", err)
        }

        product, err := s.productRepo.FindByID(ctx, variant.ProductID)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch product: %w", err)
        }

        orderItem := &entities.OrderItem{
            ID:                       uuid.New(),
            OrderID:                  order.ID,
            ProductID:                product.ID,
            ProductVariantID:          cartItem.ProductVariantID,
            ProductName:               product.Name,
            ProductVariantAttributes:   variant.Attributes,
            ProductImageURL:           product.ImageURLs[0], // First image
            SKU:                      variant.SKU,
            Quantity:                 cartItem.Quantity,
            Price:                    cartItem.PriceSnapshot,
            Subtotal:                 cartItem.PriceSnapshot * float64(cartItem.Quantity),
        }

        if err := s.orderItemRepo.Create(ctx, orderItem); err != nil {
            return nil, fmt.Errorf("failed to create order item: %w", err)
        }
    }

    // 9. Initialize payment via Midtrans
    payment, err := s.paymentService.CreatePayment(ctx, order.ID, req.PaymentMethod)
    if err != nil {
        // Rollback: delete order and order items
        s.orderRepo.Delete(ctx, order.ID)
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    // 10. Clear user's cart
    if err := s.cartRepo.ClearCart(ctx, cart.ID); err != nil {
        s.logger.Error("failed to clear cart", zap.Error(err))
        // Non-critical error, order already created
    }

    // 11. Send order creation email
    go s.sendOrderCreatedEmail(ctx, order, payment)

    return order, nil
}

// CancelOrder cancels a pending order
func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, reason string) error {
    order, err := s.orderRepo.FindByID(ctx, orderID)
    if err != nil {
        return ErrOrderNotFound
    }

    if order.UserID != userID {
        return ErrOrderNotOwned
    }

    if order.Status != "pending_payment" {
        return ErrOrderCannotCancel{Status: order.Status}
    }

    // Update order status
    now := time.Now()
    order.Status = "cancelled"
    order.CancelledAt = &now
    order.CancellationReason = &reason

    if err := s.orderRepo.Update(ctx, order); err != nil {
        return fmt.Errorf("failed to cancel order: %w", err)
    }

    // Cancel payment if exists
    if payment, err := s.paymentRepo.FindByOrderID(ctx, orderID); err == nil {
        s.paymentService.CancelPayment(ctx, payment.ID)
    }

    // Send cancellation email
    go s.sendOrderCancelledEmail(ctx, order, reason)

    return nil
}

// generateOrderNumber generates unique order number: ORD-YYYYMMDD-XXXXX
func (s *OrderService) generateOrderNumber(ctx context.Context) (string, error) {
    today := time.Now().Format("20060102")

    // Get today's order count
    count, err := s.orderRepo.CountOrdersByDate(ctx, today)
    if err != nil {
        return "", err
    }

    // Generate 5-digit sequence number (00001-99999)
    sequence := count + 1
    if sequence > 99999 {
        sequence = sequence % 99999 // Reset if exceeds
    }

    return fmt.Sprintf("ORD-%s-%05d", today, sequence), nil
}

// calculateTax calculates 11% PPN (Indonesia VAT)
func (s *OrderService) calculateTax(subtotal float64) float64 {
    return subtotal * 0.11
}

// applyVoucher applies discount code if valid
func (s *OrderService) applyVoucher(ctx context.Context, userID uuid.UUID, code string, subtotal float64) (float64, error) {
    voucher, err := s.voucherRepo.FindByCode(ctx, code)
    if err != nil {
        return 0, ErrVoucherInvalid
    }

    if voucher.ExpiresAt.Before(time.Now()) {
        return 0, ErrVoucherExpired
    }

    if voucher.UsedCount >= voucher.MaxUses {
        return 0, ErrVoucherExhausted
    }

    if subtotal < voucher.MinPurchase {
        return 0, ErrVoucherMinPurchase{Min: voucher.MinPurchase}
    }

    // Calculate discount
    var discount float64
    if voucher.DiscountType == "fixed" {
        discount = voucher.Amount
    } else {
        // Percentage discount
        discount = subtotal * (voucher.Amount / 100)
    }

    // Cap discount
    if voucher.MaxDiscount > 0 && discount > voucher.MaxDiscount {
        discount = voucher.MaxDiscount
    }

    return discount, nil
}

// ProcessPaymentVerified handles successful payment (called from webhook)
func (s *OrderService) ProcessPaymentVerified(ctx context.Context, payment *entities.Payment) error {
    order, err := s.orderRepo.FindByID(ctx, payment.OrderID)
    if err != nil {
        return err
    }

    // Update order status
    order.Status = "payment_verified"
    if err := s.orderRepo.Update(ctx, order); err != nil {
        return err
    }

    // Deduct stock from variants
    orderItems, err := s.orderItemRepo.FindByOrderID(ctx, order.ID)
    if err != nil {
        return err
    }

    for _, item := range orderItems {
        variant, err := s.productVariantRepo.FindByID(ctx, item.ProductVariantID)
        if err != nil {
            return err
        }

        newStock := variant.StockQuantity - item.Quantity
        if err := s.productVariantRepo.UpdateStock(ctx, variant.ID, newStock); err != nil {
            return err
        }
    }

    // Transition to processing
    order.Status = "processing"
    if err := s.orderRepo.Update(ctx, order); err != nil {
        return err
    }

    // Send emails
    go s.sendPaymentSuccessEmail(ctx, order)
    go s.sendOrderConfirmedToWarehouse(ctx, order)

    return nil
}
```

---

### Payment Service

**File:** `internal/application/services/payment_service.go`

```go
type PaymentService struct {
    paymentRepo    ports.PaymentRepository
    midtransClient *MidtransClient
    logger         *zap.Logger
}

// CreatePayment creates Midtrans transaction for order
func (s *PaymentService) CreatePayment(ctx context.Context, orderID uuid.UUID, method string) (*entities.Payment, error) {
    order, err := s.orderRepo.FindByID(ctx, orderID)
    if err != nil {
        return nil, fmt.Errorf("order not found: %w", err)
    }

    if order.Status != "pending_payment" {
        return nil, ErrOrderNotPendingPayment
    }

    // Check if payment already exists
    existing, _ := s.paymentRepo.FindByOrderID(ctx, orderID)
    if existing != nil {
        // Return existing payment URL
        return existing, nil
    }

    // Build Midtrans request
    midtransReq := MidtransTransactionRequest{
        TransactionDetails: TransactionDetails{
            OrderID:     order.OrderNumber,
            GrossAmount: int64(order.Total),
        },
        ItemDetails: s.buildMidtransItems(order),
        CustomerDetails: CustomerDetails{
            FirstName: order.ShippingAddress.RecipientName,
            Email:     order.User.Email,
            Phone:     order.ShippingAddress.Phone,
        },
        Expiry: Expiry{
            Unit:    "hours",
            Duration: 24, // 24 hours expiration
        },
    }

    // Set payment method specifics
    switch method {
    case "credit_card":
        midtransReq.PaymentDetails = CreditCardDetails{
            Secure: true, // 3DS secure
        }
    case "va_bca", "va_mandiri", "va_bni", "va_bri":
        midtransReq.PaymentDetails = VADetails{
            Bank: strings.TrimPrefix(method, "va_"),
        }
    case "gopay", "ovo", "dana", "shopeepay":
        midtransReq.PaymentDetails = EWalletDetails{
            EWallet: method,
        }
    case "qris":
        midtransReq.PaymentDetails = QRISDetails{}
    case "cod":
        midtransReq.PaymentDetails = CODDetails{}
    }

    // Call Midtrans API
    midtransResp, err := s.midtransClient.CreateTransaction(midtransReq)
    if err != nil {
        return nil, fmt.Errorf("midtrans API call failed: %w", err)
    }

    // Create payment record
    expiresAt := time.Now().Add(24 * time.Hour)
    payment := &entities.Payment{
        ID:                     uuid.New(),
        OrderID:                orderID,
        Method:                 method,
        Status:                 "pending",
        Amount:                 order.Total,
        MidtransTransactionID:  midtransResp.TransactionID,
        MidtransPaymentType:    midtransResp.PaymentType,
        PaymentURL:             midtransResp.RedirectURL,
        VANumber:              midtransResp.VANumber,
        VABank:               midtransResp.Bank,
        ExpiresAt:             &expiresAt,
        RawMidtransResponse:   midtransResp.RawData,
    }

    if err := s.paymentRepo.Create(ctx, payment); err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    return payment, nil
}

// ProcessWebhook handles Midtrans payment status notification
func (s *PaymentService) ProcessWebhook(ctx context.Context, webhookData map[string]interface{}) error {
    // 1. Verify signature
    signatureKey := webhookData["signature_key"].(string)
    orderID := webhookData["order_id"].(string)
    statusCode := webhookData["status_code"].(string)
    grossAmount := webhookData["gross_amount"].(string)

    if !s.verifySignature(orderID, statusCode, grossAmount, signatureKey) {
        return ErrInvalidSignature
    }

    // 2. Find order by order_number
    order, err := s.orderRepo.FindByOrderNumber(ctx, orderID)
    if err != nil {
        return ErrOrderNotFound
    }

    // 3. Find or create payment record
    payment, err := s.paymentRepo.FindByOrderID(ctx, order.ID)
    if err != nil {
        // Payment not found, create new
        payment = &entities.Payment{
            ID:       uuid.New(),
            OrderID:  order.ID,
            Amount:    order.Total,
            Status:    "pending",
        }
    }

    // 4. Update payment based on transaction_status
    transactionStatus := webhookData["transaction_status"].(string)
    payment.MidtransTransactionID = webhookData["transaction_id"].(string)
    payment.MidtransPaymentType = webhookData["payment_type"].(string)
    payment.RawMidtransResponse = webhookData

    switch transactionStatus {
    case "settlement", "capture":
        payment.Status = "success"
        paidAt := time.Now()
        payment.PaidAt = &paidAt
        payment.FailureReason = nil

        // Trigger order processing
        s.orderService.ProcessPaymentVerified(ctx, payment)

    case "pending":
        payment.Status = "processing"

    case "deny", "cancel":
        payment.Status = "failed"
        payment.FailureReason = new(string)
        *payment.FailureReason = webhookData["status_message"].(string)

        s.orderService.ProcessPaymentFailed(ctx, payment)

    case "expire":
        payment.Status = "expired"
        payment.FailureReason = new(string)
        *payment.FailureReason = "Payment expired after 24 hours"

        s.orderService.ProcessPaymentExpired(ctx, payment)
    }

    if err := s.paymentRepo.Update(ctx, payment); err != nil {
        s.logger.Error("failed to update payment", zap.Error(err))
    }

    return nil
}

// verifySignature validates Midtrans webhook signature
func (s *PaymentService) verifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
    serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

    // Calculate expected signature
    data := fmt.Sprintf("%s%s%s", orderID, statusCode, grossAmount)
    hash := sha512.Sum512([]byte(data + serverKey))
    expectedSig := fmt.Sprintf("%x", hash)

    // Compare (timing-safe)
    return subtle.ConstantTimeCompare([]byte(expectedSig), []byte(signatureKey)) == 1
}

// CancelExpiredPayments runs as cron job to cancel orders with expired payments
func (s *PaymentService) CancelExpiredPayments(ctx context.Context) error {
    expiredPayments, err := s.paymentRepo.FindExpiredPayments(ctx, time.Now())
    if err != nil {
        return err
    }

    for _, payment := range expiredPayments {
        if payment.Status == "pending" {
            payment.Status = "expired"
            payment.FailureReason = new(string)
            *payment.FailureReason = "Payment expired after 24 hours"

            s.paymentRepo.Update(ctx, payment)
            s.orderService.ProcessPaymentExpired(ctx, payment)

            s.logger.Info("cancelled expired payment",
                zap.String("payment_id", payment.ID.String()),
                zap.String("order_id", payment.OrderID.String()))
        }
    }

    return nil
}

// buildMidtransItems converts order items to Midtrans format
func (s *PaymentService) buildMidtransItems(order *entities.Order) []ItemDetail {
    orderItems, _ := s.orderItemRepo.FindByOrderID(context.Background(), order.ID)

    items := make([]ItemDetail, 0, len(orderItems))
    for _, item := range orderItems {
        items = append(items, ItemDetail{
            ID:    item.SKU,
            Price: int64(item.Price),
            Qty:   item.Quantity,
            Name:  item.ProductName,
        })
    }

    // Add shipping as separate item
    items = append(items, ItemDetail{
        ID:    "SHIPPING",
        Price: int64(order.ShippingCost),
        Qty:   1,
        Name:  fmt.Sprintf("Shipping (%s %s)", order.Courier, order.CourierService),
    })

    return items
}
```

---

## Error Handling

### Custom Error Definitions

**File:** `internal/domain/errors/errors.go`

```go
package errors

import (
    "fmt"
    "github.com/google/uuid"
)

// Address errors
var (
    ErrAddressNotFound    = fmt.Errorf("address not found")
    ErrAddressNotOwned   = fmt.Errorf("address does not belong to user")
    ErrAddressDeleted    = fmt.Errorf("address has been deleted")
    ErrAddressInUse     = fmt.Errorf("address is used in active orders")
    ErrInvalidPhone     = fmt.Errorf("invalid phone number format")
    ErrInvalidPostalCode = fmt.Errorf("invalid postal code format (must be 5 digits)")
)

// Order errors
var (
    ErrOrderNotFound          = fmt.Errorf("order not found")
    ErrOrderNotOwned         = fmt.Errorf("order does not belong to user")
    ErrCartEmpty            = fmt.Errorf("cart is empty")
    ErrOrderCannotCancel     = fmt.Errorf("order cannot be cancelled")
    ErrOrderNotPendingPayment = fmt.Errorf("order is not in pending_payment status")
    ErrShippingCostMismatch  = fmt.Errorf("shipping cost does not match calculated cost")
)

// Payment errors
var (
    ErrPaymentNotFound    = fmt.Errorf("payment not found")
    ErrInvalidSignature  = fmt.Errorf("invalid webhook signature")
)

// Voucher errors
var (
    ErrVoucherInvalid      = fmt.Errorf("voucher code is invalid")
    ErrVoucherExpired     = fmt.Errorf("voucher has expired")
    ErrVoucherExhausted   = fmt.Errorf("voucher has reached maximum uses")
    ErrVoucherMinPurchase = fmt.Errorf("minimum purchase amount not met")
)

// RajaOngkir errors
var (
    ErrProvinceNotFound = fmt.Errorf("province not found")
    ErrCityNotFound     = fmt.Errorf("city not found")
    ErrRajaongkirAPI   = fmt.Errorf("rajaongkir API error")
)

// Structured errors
type ErrInsufficientStock struct {
    Available int
    Requested int
}

func (e ErrInsufficientStock) Error() string {
    return fmt.Sprintf("insufficient stock: %d requested, %d available", e.Requested, e.Available)
}

type ErrCartStockValidationFailed struct {
    Errors []StockError
}

func (e ErrCartStockValidationFailed) Error() string {
    return fmt.Sprintf("cart stock validation failed: %d items have issues", len(e.Errors))
}

type StockError struct {
    CartItemID uuid.UUID
    VariantID  uuid.UUID
    Available  int
    InCart     int
    Error      string
}

type ErrOrderCannotCancel struct {
    Status string
}

func (e ErrOrderCannotCancel) Error() string {
    return fmt.Sprintf("order cannot be cancelled in status: %s", e.Status)
}

type ErrShippingCostMismatch struct {
    Provided   float64
    Calculated float64
}

func (e ErrShippingCostMismatch) Error() string {
    return fmt.Sprintf("shipping cost mismatch: provided %.2f, calculated %.2f", e.Provided, e.Calculated)
}

type ErrVoucherMinPurchase struct {
    Min float64
}

func (e ErrVoucherMinPurchase) Error() string {
    return fmt.Sprintf("minimum purchase amount %.2f not met", e.Min)
}
```

---

### HTTP Status Code Mapping

**File:** `internal/adapters/api/middleware/error_handler.go`

```go
func MapErrorToHTTP(err error) (int, string) {
    switch {
    case errors.Is(err, ErrAddressNotFound), errors.Is(err, ErrOrderNotFound), errors.Is(err, ErrProvinceNotFound), errors.Is(err, ErrCityNotFound):
        return 404, err.Error()

    case errors.Is(err, ErrAddressNotOwned), errors.Is(err, ErrOrderNotOwned):
        return 403, err.Error()

    case errors.Is(err, ErrCartEmpty), errors.Is(err, ErrInvalidPhone), errors.Is(err, ErrInvalidPostalCode):
        return 400, err.Error()

    case errors.Is(err, ErrAddressInUse):
        return 400, "Cannot delete: address is used in active orders"

    case errors.Is(err, ErrOrderCannotCancel):
        return 400, "Order cannot be cancelled in current status"

    case errors.Is(err, ErrShippingCostMismatch):
        return 400, "Shipping cost does not match calculated cost"

    case errors.Is(err, ErrVoucherInvalid), errors.Is(err, ErrVoucherExpired), errors.Is(err, ErrVoucherExhausted), errors.Is(err, ErrVoucherMinPurchase):
        return 400, err.Error()

    case errors.Is(err, ErrInvalidSignature):
        return 400, "Invalid webhook signature"

    case errors.Is(err, ErrOrderNotPendingPayment):
        return 400, "Order is not eligible for payment"

    // Type assertions for structured errors
    case errors.As(err, &ErrInsufficientStock{}), errors.As(err, &ErrCartStockValidationFailed{}):
        return 400, err.Error()

    case errors.As(err, &ErrShippingCostMismatch{}):
        return 400, err.Error()

    case errors.As(err, &ErrVoucherMinPurchase{}):
        return 400, err.Error()

    default:
        // Log unexpected errors
        log.Printf("Unexpected error: %v", err)
        return 500, "Internal server error"
    }
}
```

---

## Testing Strategy

### Unit Tests

Covered in brainstorming process with detailed test examples for:
- Address Service (validation, default handling)
- Order Service (creation, cancellation, stock validation)
- Payment Service (webhook processing, signature verification)
- Shipping Service (cost calculation, caching)

---

### Integration Tests

Covered in brainstorming process with:
- Order Creation Flow (cart → order → payment → email)
- Payment Webhook Flow (midtrans → stock deduction → status updates)

---

### Performance Benchmarks

**Expected Performance Targets:**
- Order creation: < 500ms (p95)
- Shipping cost calculation: < 200ms (with cache), < 1000ms (without cache)
- Payment webhook processing: < 100ms (p95)

---

## Implementation Checklist

### Backend Tasks (Week 5-6)

**Database & Schema (Day 1-2):**
- [ ] Create Address schema in Ent
- [ ] Create Order schema in Ent
- [ ] Create OrderItem schema in Ent
- [ ] Create Payment schema in Ent
- [ ] Update User schema (add addresses, orders edges)
- [ ] Update Product schema (add order_items edge)
- [ ] Update ProductVariant schema (add order_items edge)
- [ ] Run `go generate ./internal/adapters/persistence/db/schema`
- [ ] Create migration files
- [ ] Test migrations on dev database

**Infrastructure - RajaOngkir (Day 3):**
- [ ] Implement RajaOngkir HTTP client with rate limiting
- [ ] Implement Province service (with in-memory caching)
- [ ] Implement City service (with in-memory caching)
- [ ] Implement Cost calculation service (with in-memory caching)
- [ ] Implement in-memory cache (sync.Map based)
- [ ] Write unit tests for RajaOngkir client
- [ ] Test RajaOngkir API integration with real API key

**Infrastructure - Midtrans (Day 3):**
- [ ] Implement Midtrans HTTP client
- [ ] Implement payment transaction creation
- [ ] Implement webhook handler with signature verification
- [ ] Write unit tests for Midtrans client
- [ ] Test Midtrans integration in sandbox environment

**Repositories (Day 4):**
- [ ] Implement AddressRepository (Ent adapter)
- [ ] Implement OrderRepository (Ent adapter)
- [ ] Implement OrderItemRepository (Ent adapter)
- [ ] Implement PaymentRepository (Ent adapter)
- [ ] Write repository unit tests

**Services (Day 5-6):**
- [ ] Implement AddressService (CRUD, validation)
- [ ] Implement ShippingService (provinces, cities, cost calculation)
- [ ] Implement OrderService (create, cancel, status transitions)
- [ ] Implement PaymentService (create, webhook, expiration handler)
- [ ] Implement StockValidatorService (reuse from Phase 2)
- [ ] Implement PriceCalculatorService (tax calculation)
- [ ] Write service unit tests

**Handlers (Day 7):**
- [ ] Implement Address handlers (GET, POST, PUT, DELETE, set default)
- [ ] Implement Shipping handlers (provinces, cities, cost)
- [ ] Implement Order handlers (create, get list, get detail, cancel)
- [ ] Implement Payment handlers (create, status check, webhook)
- [ ] Add Huma OpenAPI documentation
- [ ] Write integration tests for all endpoints

**Email Notifications (Day 8):**
- [ ] Implement EmailService with SMTP
- [ ] Create email templates (order created, payment success, cancelled, etc.)
- [ ] Integrate email sending into OrderService and PaymentService
- [ ] Write email service tests

**Cron Jobs & Automation (Day 8):**
- [ ] Implement expired payment cancellation cron job
- [ ] Implement email notification retry job
- [ ] Configure cron scheduler (e.g., robfig/cron)
- [ ] Write cron job tests

**Testing & Polish (Day 9-10):**
- [ ] Full integration testing (end-to-end flows)
- [ ] Load testing (concurrent order creation, webhook handling)
- [ ] Performance benchmarks (order creation, shipping cost calculation)
- [ ] Error handling & validation testing
- [ ] Security testing (webhook signature verification, authorization)
- [ ] API documentation review (OpenAPI spec)
- [ ] Bug fixes & edge case handling
- [ ] Final code review

---

## Success Criteria

**Phase 3 Complete When:**
- ✅ Address management working (create, update, delete, set default)
- ✅ Shipping cost calculation working with RajaOngkir
- ✅ Order creation from cart working
- ✅ Payment initialization with Midtrans Snap working
- ✅ Payment webhook handler working with signature verification
- ✅ Order status transitions working (pending → payment_verified → processing)
- ✅ Stock deduction after payment working
- ✅ Expired payment cancellation working (cron job)
- ✅ Email notifications working (order created, payment success)
- ✅ All tests passing (unit + integration + benchmarks)
- ✅ API documentation complete
- ✅ Performance benchmarks met (< 500ms order creation, < 200ms shipping cost)

---

## Next Steps

**After Phase 3:**
1. Frontend integration (Phase 3 frontend)
2. Move to Phase 4: Order Tracking & Reviews
3. Consider post-MVP enhancements:
   - Voucher system implementation
   - Order status tracking with courier API integration
   - Email notification improvements (templates, attachments)
   - PDF invoice generation

---

**Document Status:** ✅ Complete
**Approved By:** [Pending]
**Implementation Start:** [TBD]
**Target Completion:** Week 5-6 (10 days)
