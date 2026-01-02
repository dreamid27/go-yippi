# Checkout & Payment System - Implementation Design

**Date:** 2026-01-02
**Status:** Approved
**Phase:** Phase 3 Completion

## Overview

Complete implementation of the checkout and payment system including repositories, service layer, API handlers, and integration tests.

## Architecture Decisions

### 1. Implementation Approach
**Parallel with Checkpoints**
- Repositories built in parallel (Order, OrderItem, Payment)
- Checkpoint after repositories before checkout service
- Checkpoint after checkout service before handlers
- Balances speed with quality control

### 2. Testing Strategy
**Repository Integration Tests**
- Testcontainers for PostgreSQL isolation
- Full transaction testing
- Real database queries
- Fast enough (< 1 second per test)

### 3. Transaction Management
**Full Transaction Support**
- Entire checkout wrapped in database transaction
- Atomicity: all-or-nothing for order creation
- Rollback on any failure
- Simple and debuggable

### 4. API Framework
**Huma v2 with DTOs**
- Separate request/response DTOs
- Auto-generated OpenAPI docs
- Framework-level validation
- Clean API boundary

### 5. Validation Strategy
**Split Validation**
- Handlers: HTTP concerns (auth, basic input)
- Services: Business rules (stock, ownership, addresses)
- Clean separation of concerns

### 6. Payment Flow
**Payment Before Midtrans**
- Create payment record (pending status)
- Call Midtrans API
- Update payment with token and URL
- Complete audit trail

### 7. Stock Management
**Reserve on Order Creation**
- Decrement stock when order created
- Restock if cancelled
- Simple for Phase 3
- Can be enhanced later with proper reservation

### 8. Error Handling
**Domain Errors with HTTP Mapping**
- Services return typed domain errors
- Handlers map to HTTP status codes
- Consistent error responses
- HTTP-agnostic services

## Components

### 1. Repository Implementations

**Order Repository** (`internal/adapters/persistence/order_repository.go`)
- Implements OrderRepository interface (7 methods)
- Create, GetByID, GetByUserID, GetByUserIDWithStatus
- Update, UpdateStatus, Cancel
- Conversion functions between Ent and domain
- Composite index usage for performance

**OrderItem Repository** (`internal/adapters/persistence/order_item_repository.go`)
- Implements OrderItemRepository interface (4 methods)
- Create, GetByID, GetByOrderID, Update
- Price snapshot handling

**Payment Repository** (`internal/adapters/persistence/payment_repository.go`)
- Implements PaymentRepository interface (6 methods)
- Create, GetByID, GetByOrderID, Update
- UpdateStatus, MarkAsPaid
- Nullable field handling

**Common Patterns:**
- Struct: `client *ent.Client`
- Conversion: `toEntity()` method
- Constructor: `New*Repository(client *ent.Client)`
- Context-first parameters
- Domain error conversion

### 2. Checkout Service

**File:** `internal/application/services/checkout_service.go`

**Core Method:**
```go
CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
```

**Dependencies:**
- OrderRepository, OrderItemRepository, PaymentRepository
- ProductRepository, CartRepository, AddressRepository
- RajaOngkirService (already implemented)

**Flow:**
1. Validate cart exists and belongs to user
2. Validate shipping address belongs to user
3. Check all products active and in stock
4. Calculate total and verify shipping cost
5. Begin database transaction
6. Create Order (status=pending_payment)
7. Create OrderItems (with price snapshots)
8. Decrement stock for ProductVariants
9. Create Payment (status=pending)
10. Commit transaction
11. Call Midtrans API
12. Update Payment with payment_url and token
13. Return CheckoutResponse

**Error Handling:**
- Transaction rollback on error
- Domain errors (ErrInsufficientStock, ErrCartNotFound, etc.)
- Observability logging

### 3. API Handlers

**Checkout Handler** (`internal/adapters/api/handlers/checkout_handler.go`)

**DTOs:**
```go
type CreateCheckoutRequest struct {
    Body struct {
        CartID            uuid.UUID `json:"cart_id"`
        ShippingAddressID uuid.UUID `json:"shipping_address_id"`
        Courier           string    `json:"courier"`
    }
}

type CheckoutResponse struct {
    Body struct {
        OrderID     uuid.UUID `json:"order_id"`
        PaymentURL  string    `json:"payment_url"`
        TotalAmount float64   `json:"total_amount"`
        ExpiresAt   time.Time `json:"expires_at"`
    }
}
```

**Endpoints:**
- POST /checkout - Create checkout
- GET /orders/{id} - Get order details
- GET /orders - List user orders (with status filter, pagination)
- POST /orders/{id}/cancel - Cancel order

**Error Mapping:**
- ErrNotFound → 404
- ErrInsufficientStock → 400
- ErrInvalidInput → 400
- Other → 500

**Additional Handlers:**
- AddressHandler - CRUD operations
- PaymentHandler - Midtrans webhook

### 4. Integration Tests

**Framework:** Testcontainers with PostgreSQL

**Test Coverage:**

**Order Repository:**
- Create, GetByID, GetByUserID, GetByUserIDWithStatus
- Update, UpdateStatus, Cancel
- Deleted record filtering
- Error scenarios

**OrderItem Repository:**
- Create, GetByID, GetByOrderID, Update
- Price snapshot preservation

**Payment Repository:**
- Create, GetByID, GetByOrderID
- UpdateStatus, MarkAsPaid
- Nullable fields

**Checkout Service:**
- Successful checkout flow
- Empty cart validation
- Insufficient stock handling
- Address ownership validation
- Transaction rollback
- Midtrans API (mocked)

**Test Pattern:**
```go
func TestOrderRepository_Create(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Stop()

    repo := persistence.NewOrderRepository(db.Client())
    order := &entities.Order{...}

    err := repo.Create(ctx, order)

    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, order.ID)
}
```

## Data Flow

```
HTTP Request (CreateCheckout)
    ↓
[Handler] Validate JWT, basic input
    ↓
[Handler] Call CheckoutService.CreateCheckout()
    ↓
[Service] Validate cart exists, user owns cart
    ↓
[Service] Validate address ownership
    ↓
[Service] Validate stock availability
    ↓
[Service] Begin Transaction
    ↓
[Service] Create Order
    ↓
[Service] Create OrderItems
    ↓
[Service] Decrement Stock
    ↓
[Service] Create Payment
    ↓
[Service] Commit Transaction
    ↓
[Service] Call Midtrans API
    ↓
[Service] Update Payment with token
    ↓
[Handler] Return CheckoutResponse with payment URL
```

## Implementation Order

**Phase 1: Repositories** (Parallel)
1. Order repository implementation
2. OrderItem repository implementation
3. Payment repository implementation
4. Integration tests for all repositories
5. Checkpoint: Review and validate

**Phase 2: Checkout Service**
1. Checkout service implementation
2. Business logic validation
3. Transaction handling
4. Integration tests
5. Checkpoint: Review and validate

**Phase 3: API Handlers**
1. Checkout handler with DTOs
2. Address handler
3. Payment handler (webhook)
4. Route registration
5. Integration tests
6. Final review

## Success Criteria

✅ All 3 repositories implemented with integration tests
✅ Checkout service handles complete flow with transactions
✅ API handlers expose REST endpoints with OpenAPI docs
✅ Integration tests cover critical paths
✅ Full checkout flow works end-to-end
✅ Midtrans integration functional
✅ Error handling consistent across layers
✅ Code compiles and tests pass

## Future Enhancements

- Proper inventory reservation system (soft reservations)
- Async payment processing
- Enhanced order management (filtering, sorting, pagination)
- Payment retry logic
- Order status webhooks
- Admin panel integration
