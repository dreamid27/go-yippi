# Phase 2 Backend Design - Product Catalog & Cart System

**Date:** 2025-12-31
**Project:** go-yippi (Backend)
**Phase:** 2 - Product Catalog & Cart (Week 3-4)
**Status:** Design Complete, Ready for Implementation

---

## Executive Summary

This document details the technical design for Phase 2 backend implementation, covering:
- **Product Variants System** with flexible JSONB attributes
- **Cart System** with stock validation (authenticated users only)
- **PostgreSQL Full-Text Search** for product discovery
- **Enhanced Product APIs** with filtering & search capabilities

**Key Design Decisions:**
1. ✅ **JSONB-based variant attributes** - Maximum flexibility without schema changes
2. ✅ **Base price + adjustment model** - Easy bulk price updates
3. ✅ **Always create default variant** - Unified logic for all products
4. ✅ **Guest cart in localStorage** - Clean backend, focus on logged-in UX
5. ✅ **PostgreSQL Full-Text Search** - Better ranking, native solution

---

## Database Schema Changes

### 1. Product Schema Updates

**File:** `internal/adapters/persistence/db/schema/product.go`

**Changes to existing Product:**

```go
// REMOVE these fields:
// - field.String("sku") → Moved to ProductVariant level

// RENAME this field:
// - field.Float("price") → Rename to "base_price"

// ADD these new fields:
func (Product) Fields() []ent.Field {
    return []ent.Field{
        // ... existing fields (slug, name, description, etc.) ...

        field.Float("base_price").
            Positive().
            Comment("Base price for price calculation with variants"),

        field.Int("stock_quantity").
            NonNegative().
            Optional().
            Nillable().
            Comment("Stock for non-variant products, NULL for variant products"),

        field.Int("low_stock_threshold").
            NonNegative().
            Default(10).
            Comment("Alert when stock below this number"),

        field.Other("search_vector", &tsvector{}).
            SchemaType(map[string]string{
                dialect.Postgres: "tsvector",
            }).
            Optional().
            Comment("Full-text search vector"),

        // ... existing fields (weight, dimensions, image_urls, status, etc.) ...
    }
}

// ADD this edge:
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        // ... existing edges (category, brand) ...

        edge.To("variants", ProductVariant.Type),
    }
}
```

**Migration Summary:**
- Remove `sku` column (data migration: copy to default variant)
- Rename `price` → `base_price`
- Add `stock_quantity` (nullable)
- Add `low_stock_threshold` (default 10)
- Add `search_vector` (tsvector, nullable)

---

### 2. ProductVariant Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/product_variant.go`

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

type ProductVariant struct {
    ent.Schema
}

func (ProductVariant) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            StorageKey("id"),

        field.String("sku").
            NotEmpty().
            Unique().
            Comment("Variant SKU (unique across all variants)"),

        field.JSON("attributes", map[string]string{}).
            Comment("Variant attributes: {\"size\": \"M\", \"color\": \"Red\"}"),

        field.Int("stock_quantity").
            NonNegative().
            Default(0).
            Comment("Stock quantity for this variant"),

        field.Float("price_adjustment").
            Default(0).
            Comment("Price adjustment from base_price (can be negative)"),

        field.Bool("is_active").
            Default(true).
            Comment("Is this variant available for sale"),

        field.UUID("product_id", uuid.UUID{}).
            Comment("Product ID"),

        field.Time("created_at").
            Default(time.Now).
            Immutable(),

        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (ProductVariant) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("product", Product.Type).
            Ref("variants").
            Unique().
            Required().
            Field("product_id"),

        edge.To("cart_items", CartItem.Type),
    }
}

func (ProductVariant) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("product_id"),
        index.Fields("sku").Unique(),
    }
}
```

**Key Features:**
- **JSONB attributes** - Flexible variant types (size, color, flavor, packaging, etc.)
- **Price calculation** - `final_price = product.base_price + variant.price_adjustment`
- **Stock per variant** - Each combination has independent stock
- **Always create 1 variant** - Even for products without variants (empty attributes `{}`)

**Example Data:**

```json
// Product: T-Shirt (with variants)
{
  "id": "uuid-1",
  "name": "Cotton T-Shirt",
  "base_price": 100000,
  "stock_quantity": null
}

// Variants:
[
  {
    "sku": "TSHIRT-S-BLACK",
    "attributes": {"size": "S", "color": "Black"},
    "stock_quantity": 50,
    "price_adjustment": 0,
    "final_price": 100000
  },
  {
    "sku": "TSHIRT-XL-WHITE",
    "attributes": {"size": "XL", "color": "White"},
    "stock_quantity": 20,
    "price_adjustment": 10000,
    "final_price": 110000
  }
]

// Product: Leather Bag (no variants)
{
  "id": "uuid-2",
  "name": "Leather Bag",
  "base_price": 500000,
  "stock_quantity": null
}

// Variant (default):
{
  "sku": "BAG-001",
  "attributes": {},
  "stock_quantity": 10,
  "price_adjustment": 0,
  "final_price": 500000
}
```

---

### 3. Cart Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/cart.go`

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

type Cart struct {
    ent.Schema
}

func (Cart) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            StorageKey("id"),

        field.UUID("user_id", uuid.UUID{}).
            Comment("User ID - owner of this cart"),

        field.Time("created_at").
            Default(time.Now).
            Immutable(),

        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (Cart) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).
            Ref("cart").
            Unique().
            Required().
            Field("user_id"),

        edge.To("items", CartItem.Type),
    }
}

func (Cart) Indexes() []ent.Index {
    return []ent.Index{
        // One cart per user
        index.Fields("user_id").Unique(),
    }
}
```

---

### 4. CartItem Schema (NEW)

**File:** `internal/adapters/persistence/db/schema/cart_item.go`

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

type CartItem struct {
    ent.Schema
}

func (CartItem) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            StorageKey("id"),

        field.UUID("cart_id", uuid.UUID{}).
            Comment("Cart ID"),

        field.UUID("product_variant_id", uuid.UUID{}).
            Comment("Product Variant ID"),

        field.Int("quantity").
            Positive().
            Default(1).
            Comment("Quantity of this variant in cart"),

        field.Float("price_snapshot").
            Positive().
            Comment("Price snapshot at time of adding to cart"),

        field.Time("created_at").
            Default(time.Now).
            Immutable(),

        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (CartItem) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("cart", Cart.Type).
            Ref("items").
            Unique().
            Required().
            Field("cart_id"),

        edge.From("product_variant", ProductVariant.Type).
            Ref("cart_items").
            Unique().
            Required().
            Field("product_variant_id"),
    }
}

func (CartItem) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("cart_id"),
        index.Fields("product_variant_id"),
        // Prevent duplicate variant in same cart
        index.Fields("cart_id", "product_variant_id").Unique(),
    }
}
```

**Key Design Points:**
- **One cart per user** - Unique constraint on `user_id`
- **No duplicate variants in cart** - Unique constraint on `(cart_id, product_variant_id)`
- **Price snapshot** - Store price at add-to-cart time (handle price changes gracefully)
- **Always reference variant_id** - Unified logic for all products

---

### 5. User Schema Update

**File:** `internal/adapters/persistence/db/schema/user.go`

**Add this edge:**

```go
func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("refresh_tokens", RefreshToken.Type),

        // NEW: One cart per user
        edge.To("cart", Cart.Type).
            Unique(),
    }
}
```

---

## PostgreSQL Full-Text Search Setup

### Migration: Add Search Vector

**File:** `migrations/002_add_product_search.sql`

```sql
-- Create function to update search vector
CREATE OR REPLACE FUNCTION products_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for auto-update
CREATE TRIGGER products_search_vector_trigger
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION products_search_vector_update();

-- Create GIN index for fast full-text search
CREATE INDEX products_search_vector_idx ON products USING GIN(search_vector);

-- Update existing products
UPDATE products SET search_vector =
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(description, '')), 'B');
```

**How it works:**
- `setweight(..., 'A')` - Product name has higher ranking priority
- `setweight(..., 'B')` - Description has lower priority
- Trigger auto-updates `search_vector` on INSERT/UPDATE
- GIN index enables fast full-text queries
- Supports relevance ranking with `ts_rank()`

---

## API Endpoints

### 1. Product APIs (Enhanced)

#### GET /api/products

**Description:** List products with search, filter, and pagination

**Query Parameters:**

```go
type ProductFilterParams struct {
    // Search
    Search string `query:"search"` // Full-text search on name & description

    // Filters
    CategoryID *uuid.UUID `query:"category_id"`
    BrandID    *uuid.UUID `query:"brand_id"`
    MinPrice   *float64   `query:"min_price"`
    MaxPrice   *float64   `query:"max_price"`
    Status     *string    `query:"status"` // published, draft, archived

    // Variant filters (JSONB)
    Size  *string `query:"size"`  // Filter by variant size
    Color *string `query:"color"` // Filter by variant color

    // Sorting
    SortBy    string `query:"sort_by"`    // name, price, created_at
    SortOrder string `query:"sort_order"` // asc, desc

    // Pagination
    Page  int `query:"page"`  // Default: 1
    Limit int `query:"limit"` // Default: 20, Max: 100
}
```

**Response:**

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "cotton-tshirt",
      "name": "Cotton T-Shirt",
      "base_price": 100000,
      "description": "...",
      "image_urls": ["url1", "url2"],
      "status": "published",
      "category": { "id": "uuid", "name": "Tops" },
      "brand": { "id": "uuid", "name": "Brand Name" },
      "variants_count": 6,
      "min_price": 100000,
      "max_price": 110000,
      "has_stock": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

#### GET /api/products/:id

**Description:** Get product detail with all variants

**Response:**

```json
{
  "id": "uuid",
  "slug": "cotton-tshirt",
  "name": "Cotton T-Shirt",
  "base_price": 100000,
  "description": "High quality cotton t-shirt",
  "image_urls": ["url1", "url2"],
  "status": "published",
  "weight": 200,
  "dimensions": { "length": 30, "width": 20, "height": 2 },
  "category": { "id": "uuid", "name": "Tops" },
  "brand": { "id": "uuid", "name": "Brand Name" },
  "variants": [
    {
      "id": "variant-uuid-1",
      "sku": "TSHIRT-S-BLACK",
      "attributes": { "size": "S", "color": "Black" },
      "stock_quantity": 50,
      "price_adjustment": 0,
      "final_price": 100000,
      "is_active": true,
      "is_in_stock": true
    },
    {
      "id": "variant-uuid-2",
      "sku": "TSHIRT-XL-WHITE",
      "attributes": { "size": "XL", "color": "White" },
      "stock_quantity": 0,
      "price_adjustment": 10000,
      "final_price": 110000,
      "is_active": true,
      "is_in_stock": false
    }
  ],
  "created_at": "2025-12-31T10:00:00Z",
  "updated_at": "2025-12-31T10:00:00Z"
}
```

---

#### POST /api/products (Admin only)

**Request:**

```json
{
  "slug": "cotton-tshirt",
  "name": "Cotton T-Shirt",
  "base_price": 100000,
  "description": "High quality cotton t-shirt",
  "image_urls": ["url1", "url2"],
  "weight": 200,
  "length": 30,
  "width": 20,
  "height": 2,
  "category_id": "category-uuid",
  "brand_id": "brand-uuid",
  "status": "draft"
}
```

**Response:** Created product (201)

---

#### PUT /api/products/:id (Admin only)

**Request:** Same as POST (partial updates allowed)

**Response:** Updated product (200)

---

#### DELETE /api/products/:id (Admin only)

**Response:** 204 No Content

**Business Logic:** Cascade delete all variants

---

### 2. Product Variant APIs (NEW)

#### GET /api/products/:product_id/variants

**Description:** List all variants for a product

**Response:**

```json
{
  "data": [
    {
      "id": "variant-uuid",
      "sku": "TSHIRT-M-RED",
      "attributes": { "size": "M", "color": "Red" },
      "stock_quantity": 30,
      "price_adjustment": 0,
      "final_price": 100000,
      "is_active": true,
      "is_in_stock": true
    }
  ]
}
```

---

#### POST /api/products/:product_id/variants (Admin only)

**Request:**

```json
{
  "sku": "TSHIRT-M-RED",
  "attributes": { "size": "M", "color": "Red" },
  "stock_quantity": 30,
  "price_adjustment": 0,
  "is_active": true
}
```

**Response:** Created variant (201)

**Validation:**
- SKU must be unique across all variants
- stock_quantity >= 0
- Product must exist

---

#### PUT /api/products/:product_id/variants/:variant_id (Admin only)

**Request:** Same as POST

**Response:** Updated variant (200)

---

#### PATCH /api/products/:product_id/variants/:variant_id/stock (Admin only)

**Description:** Quick stock update (for inventory management)

**Request:**

```json
{
  "stock_quantity": 50
}
```

**Response:** Updated variant (200)

---

#### DELETE /api/products/:product_id/variants/:variant_id (Admin only)

**Response:** 204 No Content

**Business Logic:**
- Prevent deletion if variant is in active carts/orders
- Or soft delete (set `is_active = false`)

---

### 3. Cart APIs

#### GET /api/cart

**Authentication:** Required (JWT)

**Description:** Get current user's cart with all items

**Response:**

```json
{
  "id": "cart-uuid",
  "user_id": "user-uuid",
  "items": [
    {
      "id": "cart-item-uuid",
      "quantity": 2,
      "price_snapshot": 100000,
      "subtotal": 200000,
      "product_variant": {
        "id": "variant-uuid",
        "sku": "TSHIRT-M-BLACK",
        "attributes": { "size": "M", "color": "Black" },
        "stock_quantity": 50,
        "final_price": 100000,
        "is_active": true,
        "is_in_stock": true
      },
      "product": {
        "id": "product-uuid",
        "name": "Cotton T-Shirt",
        "slug": "cotton-tshirt",
        "image_urls": ["url1"]
      }
    }
  ],
  "subtotal": 200000,
  "item_count": 2,
  "created_at": "2025-12-31T10:00:00Z",
  "updated_at": "2025-12-31T11:30:00Z"
}
```

**Business Logic:**
- Auto-create cart if user doesn't have one
- Return empty cart if no items

---

#### POST /api/cart/items

**Authentication:** Required (JWT)

**Description:** Add item to cart

**Request:**

```json
{
  "product_variant_id": "variant-uuid",
  "quantity": 1
}
```

**Response:** Created/Updated cart item (201/200)

**Business Logic:**
1. Validate variant exists and `is_active = true`
2. Check stock availability: `variant.stock_quantity >= quantity`
3. If variant already in cart → Update quantity (sum)
4. Else → Create new cart item
5. Save `price_snapshot` from variant's current final price
6. Return updated cart item

**Error Responses:**
- 404: Variant not found or inactive
- 400: Insufficient stock
- 409: Variant already in cart (if not auto-merging)

---

#### PUT /api/cart/items/:item_id

**Authentication:** Required (JWT)

**Description:** Update cart item quantity

**Request:**

```json
{
  "quantity": 3
}
```

**Response:** Updated cart item (200)

**Business Logic:**
1. Validate item belongs to current user's cart
2. Check stock availability: `variant.stock_quantity >= new_quantity`
3. Update quantity
4. If quantity = 0 → Delete item

**Error Responses:**
- 404: Cart item not found
- 403: Cart item doesn't belong to user
- 400: Insufficient stock

---

#### DELETE /api/cart/items/:item_id

**Authentication:** Required (JWT)

**Description:** Remove item from cart

**Response:** 204 No Content

**Business Logic:**
- Validate item belongs to current user's cart
- Delete item

---

#### DELETE /api/cart

**Authentication:** Required (JWT)

**Description:** Clear all items from cart

**Response:** 204 No Content

**Business Logic:**
- Delete all items from current user's cart
- Keep cart entity (for future use)

---

#### POST /api/cart/merge

**Authentication:** Required (JWT)

**Description:** Merge guest cart (from localStorage) with user cart on login

**Request:**

```json
{
  "items": [
    {
      "product_variant_id": "variant-uuid-1",
      "quantity": 2
    },
    {
      "product_variant_id": "variant-uuid-2",
      "quantity": 1
    }
  ]
}
```

**Response:** Merged cart (200)

**Business Logic:**
1. For each item in request:
   - Validate variant exists & is active
   - Check stock availability
   - If variant already in user's DB cart → Sum quantities
   - Else → Add new item
2. Validate total stock for all merged items
3. Return complete cart with all items

**Error Responses:**
- 400: Stock validation failed (return which items failed)
- 404: Variant not found

---

## Business Logic & Validation

### Stock Validation Service

**File:** `internal/application/services/stock_validator.go`

```go
type StockValidator struct {
    variantRepo ports.ProductVariantRepository
}

// ValidateStock checks if variant has sufficient stock
func (s *StockValidator) ValidateStock(ctx context.Context, variantID uuid.UUID, requestedQty int) error {
    variant, err := s.variantRepo.FindByID(ctx, variantID)
    if err != nil {
        return ErrVariantNotFound
    }

    if !variant.IsActive {
        return ErrVariantInactive{VariantID: variantID}
    }

    if variant.StockQuantity < requestedQty {
        return ErrInsufficientStock{
            VariantID: variantID,
            Available: variant.StockQuantity,
            Requested: requestedQty,
        }
    }

    return nil
}

// ValidateCartStock validates all items in cart before checkout
func (s *StockValidator) ValidateCartStock(ctx context.Context, cartID uuid.UUID) ([]StockError, error) {
    items, err := s.cartRepo.GetItems(ctx, cartID)
    if err != nil {
        return nil, err
    }

    var stockErrors []StockError

    for _, item := range items {
        variant, err := s.variantRepo.FindByID(ctx, item.ProductVariantID)
        if err != nil {
            stockErrors = append(stockErrors, StockError{
                CartItemID: item.ID,
                VariantID: item.ProductVariantID,
                Error: "Variant not found",
            })
            continue
        }

        if !variant.IsActive {
            stockErrors = append(stockErrors, StockError{
                CartItemID: item.ID,
                VariantID: variant.ID,
                Error: "Variant no longer available",
            })
            continue
        }

        if variant.StockQuantity < item.Quantity {
            stockErrors = append(stockErrors, StockError{
                CartItemID: item.ID,
                VariantID: variant.ID,
                Available: variant.StockQuantity,
                InCart: item.Quantity,
                Error: "Insufficient stock",
            })
        }
    }

    if len(stockErrors) > 0 {
        return stockErrors, ErrCartStockValidationFailed
    }

    return nil, nil
}
```

---

### Price Calculation Service

**File:** `internal/application/services/price_calculator.go`

```go
type PriceCalculator struct {
    productRepo ports.ProductRepository
}

// CalculateVariantPrice calculates final price for a variant
func (p *PriceCalculator) CalculateVariantPrice(product *entities.Product, variant *entities.ProductVariant) float64 {
    return product.BasePrice + variant.PriceAdjustment
}

// CalculateCartSubtotal calculates cart subtotal
func (p *PriceCalculator) CalculateCartSubtotal(items []*entities.CartItem) float64 {
    var subtotal float64
    for _, item := range items {
        subtotal += item.PriceSnapshot * float64(item.Quantity)
    }
    return subtotal
}
```

---

## Repository Layer

### ProductVariantRepository Interface

**File:** `internal/domain/ports/product_variant_repository.go`

```go
type ProductVariantRepository interface {
    Create(ctx context.Context, variant *entities.ProductVariant) error
    FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error)
    FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error)
    FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error)
    Update(ctx context.Context, variant *entities.ProductVariant) error
    UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

---

### CartRepository Interface

**File:** `internal/domain/ports/cart_repository.go`

```go
type CartRepository interface {
    // Cart management
    Create(ctx context.Context, userID uuid.UUID) (*entities.Cart, error)
    FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Cart, error)
    GetOrCreate(ctx context.Context, userID uuid.UUID) (*entities.Cart, error)

    // Cart item management
    AddItem(ctx context.Context, item *entities.CartItem) error
    UpdateItemQuantity(ctx context.Context, itemID uuid.UUID, quantity int) error
    RemoveItem(ctx context.Context, itemID uuid.UUID) error
    ClearCart(ctx context.Context, cartID uuid.UUID) error

    // Queries
    GetItems(ctx context.Context, cartID uuid.UUID) ([]*entities.CartItem, error)
    GetItemWithDetails(ctx context.Context, itemID uuid.UUID) (*entities.CartItemDetail, error)
    GetCartWithItems(ctx context.Context, userID uuid.UUID) (*entities.CartDetail, error)

    // Validation
    ItemExists(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (bool, error)
    GetCartItemByVariant(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (*entities.CartItem, error)
}
```

---

## Implementation Checklist

### Backend Tasks (Week 3-4)

**Database & Schema (Day 1-2):**
- [ ] Create ProductVariant schema in Ent
- [ ] Create Cart schema in Ent
- [ ] Create CartItem schema in Ent
- [ ] Update Product schema (remove sku, rename price → base_price, add fields)
- [ ] Update User schema (add cart edge)
- [ ] Run `go generate ./internal/adapters/persistence/db/schema`
- [ ] Create migration: Add search_vector with trigger
- [ ] Test migrations on dev database

**Repositories (Day 3-4):**
- [ ] Implement ProductVariantRepository (Ent adapter)
- [ ] Enhance ProductRepository with search & filter
- [ ] Implement CartRepository (Ent adapter)
- [ ] Write repository unit tests

**Services (Day 5-6):**
- [ ] Implement StockValidatorService
- [ ] Implement PriceCalculatorService
- [ ] Enhance ProductService with variant support
- [ ] Implement CartService
- [ ] Write service unit tests

**Handlers (Day 7-8):**
- [ ] Implement ProductVariant handlers (CRUD)
- [ ] Enhance Product handlers (search, filter, pagination)
- [ ] Implement Cart handlers (GET, POST, PUT, DELETE)
- [ ] Implement cart/merge handler
- [ ] Add Huma OpenAPI documentation
- [ ] Integration tests for all endpoints

**Testing & Polish (Day 9-10):**
- [ ] Full integration testing
- [ ] Performance testing (search with 10k+ products)
- [ ] Stock validation edge cases
- [ ] Error handling & validation
- [ ] API documentation review

---

## Testing Strategy

### Unit Tests

**ProductVariant:**
- Price calculation (base_price + adjustment)
- Stock validation
- JSONB attributes querying

**Cart:**
- Add item (new vs existing variant)
- Update quantity with stock check
- Cart merge logic
- Price snapshot accuracy

### Integration Tests

**Search & Filter:**
- Full-text search ranking
- Combined filters (category + price + size)
- Pagination accuracy
- Performance with large dataset

**Cart Flow:**
- Guest cart merge on login
- Stock validation before checkout
- Price changes handling (old price_snapshot vs new price)

### Performance Benchmarks

- Product search: < 200ms (p95) for 50k products
- Cart operations: < 100ms (p95)
- Filter queries: < 150ms (p95)

---

## Error Handling

### Custom Errors

```go
// Stock errors
ErrVariantNotFound = errors.New("variant not found")
ErrVariantInactive = errors.New("variant is not active")
ErrInsufficientStock = errors.New("insufficient stock")

// Cart errors
ErrCartNotFound = errors.New("cart not found")
ErrCartItemNotFound = errors.New("cart item not found")
ErrCartItemNotOwned = errors.New("cart item does not belong to user")
ErrCartStockValidationFailed = errors.New("cart stock validation failed")
```

### HTTP Status Codes

- 200: Success
- 201: Created
- 204: No Content (deleted)
- 400: Bad Request (validation failed, insufficient stock)
- 401: Unauthorized
- 403: Forbidden (cart item not owned)
- 404: Not Found
- 409: Conflict (duplicate variant in cart)
- 500: Internal Server Error

---

## Security Considerations

1. **Cart Ownership:**
   - Always validate cart/cart_item belongs to authenticated user
   - Prevent users from modifying other users' carts

2. **Stock Manipulation:**
   - Stock updates only via admin endpoints
   - Validate stock before checkout (not just at add-to-cart)

3. **Price Integrity:**
   - Use `price_snapshot` to prevent price manipulation
   - Admin-only access to price changes

4. **JSONB Injection:**
   - Validate variant attributes structure
   - Sanitize user input before JSONB queries

---

## Migration Strategy

### Data Migration from Old Schema

If existing products have data:

```sql
-- Step 1: Rename price to base_price
ALTER TABLE products RENAME COLUMN price TO base_price;

-- Step 2: Create default variant for each product
INSERT INTO product_variants (product_id, sku, attributes, stock_quantity, price_adjustment, is_active)
SELECT
    id,
    sku, -- Copy from product
    '{}', -- Empty attributes
    COALESCE(stock_quantity, 0), -- Use product stock or 0
    0, -- No price adjustment
    true
FROM products;

-- Step 3: Set product.stock_quantity to NULL (now managed in variants)
UPDATE products SET stock_quantity = NULL;

-- Step 4: Remove sku column from products (optional, can keep for backward compat)
-- ALTER TABLE products DROP COLUMN sku;
```

---

## Success Criteria

**Phase 2 Complete When:**
- ✅ Product variants CRUD working
- ✅ Cart system functional (add, update, remove, merge)
- ✅ Full-text search working with ranking
- ✅ Filters working (category, brand, price, variant attributes)
- ✅ Stock validation preventing overselling
- ✅ All tests passing (unit + integration)
- ✅ API documentation complete
- ✅ Performance benchmarks met

---

## Next Steps

**After Phase 2:**
1. Frontend integration (Phase 2 frontend)
2. Move to Phase 3: Checkout & Payment
3. Consider post-MVP enhancements:
   - Best seller tracking
   - Product recommendations
   - Wishlist functionality
   - Review system

---

**Document Status:** ✅ Complete
**Approved By:** [Pending]
**Implementation Start:** [TBD]
**Target Completion:** Week 3-4 (10 days)
