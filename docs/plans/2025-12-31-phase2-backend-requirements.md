# Phase 2 Backend Requirements - Product Catalog & Cart System

**Date:** 2026-01-01
**Project:** go-yippi (Backend)
**Phase:** 2 - Product Catalog & Cart (Week 3-4)
**Document Type:** Requirements & API Contract Specification
**Status:** Ready for Implementation

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Implementation Status](#implementation-status)
3. [Developer Requirements Checklist](#developer-requirements-checklist)
4. [API Contract Specifications](#api-contract-specifications)
5. [Acceptance Criteria](#acceptance-criteria)
6. [Testing Requirements](#testing-requirements)
7. [Non-Functional Requirements](#non-functional-requirements)

---

## Executive Summary

This document provides detailed requirements and API contracts for completing Phase 2 backend implementation. Phase 2 introduces:

- **Product Variants System** with flexible JSONB attributes
- **Cart Management** for authenticated users
- **Enhanced Product Search** with PostgreSQL Full-Text Search
- **Stock Validation** to prevent overselling

**Target Completion:** Week 3-4 (10 days)

---

## Implementation Status

### ✅ COMPLETED (Commit: 782c733)

**Schema & Database:**
- ✅ ProductVariant schema created with JSONB attributes
- ✅ Cart and CartItem schemas created
- ✅ Product schema updated (SKU removed, base_price added)
- ✅ UUID migration for all entities (Product, User, Category, Brand)

**Domain Layer:**
- ✅ ProductVariant entity with business logic
- ✅ Cart and CartItem entities
- ✅ ProductVariantRepository interface
- ✅ CartRepository interface

**Persistence Layer:**
- ✅ ProductVariantRepository implementation (Ent adapter)
- ✅ CartRepository implementation (Ent adapter)
- ✅ Product repository updated for UUID and base_price

### ⏳ REMAINING WORK

**Application Layer:**
- ⏳ ProductVariantService (business logic)
- ⏳ CartService (cart operations & validation)
- ⏳ StockValidatorService (stock checking)
- ⏳ PriceCalculatorService (price calculations)
- ⏳ Enhanced ProductService (search & filters)

**API Layer:**
- ⏳ ProductVariant handlers (CRUD endpoints)
- ⏳ Cart handlers (cart operations)
- ⏳ Enhanced Product handlers (search & filters)
- ⏳ DTOs for ProductVariant and Cart

**Infrastructure:**
- ⏳ PostgreSQL Full-Text Search setup (migration + trigger)
- ⏳ Product search index optimization
- ⏳ API documentation (Huma OpenAPI)

**Testing:**
- ⏳ Unit tests for services
- ⏳ Integration tests for API endpoints
- ⏳ Performance benchmarks

---

## Developer Requirements Checklist

### REQ-1: Product Variant Service

**File:** `internal/application/services/product_variant_service.go`

**Requirements:**

- [ ] **REQ-1.1:** Create ProductVariantService struct with dependencies
  - **Dependencies:** ProductVariantRepository, ProductRepository
  - **Interface:** Define ports.ProductVariantService interface

- [ ] **REQ-1.2:** Implement CreateVariant method
  - **Input:** productID (uuid.UUID), sku (string), attributes (map[string]string), stockQuantity (int), priceAdjustment (float64)
  - **Validation:**
    - Product must exist
    - SKU must be unique across all variants
    - stockQuantity >= 0
    - attributes must be valid JSON (not empty keys)
  - **Business Logic:**
    - Calculate final_price = product.base_price + variant.price_adjustment
    - Set is_active = true by default
  - **Output:** Created ProductVariant entity or error

- [ ] **REQ-1.3:** Implement UpdateVariant method
  - **Input:** variantID (uuid.UUID), updates (partial fields)
  - **Validation:**
    - Variant must exist
    - If updating SKU, ensure uniqueness
    - If updating stockQuantity, ensure >= 0
  - **Output:** Updated ProductVariant entity or error

- [ ] **REQ-1.4:** Implement GetVariantsByProduct method
  - **Input:** productID (uuid.UUID)
  - **Business Logic:**
    - Return all variants for a product (active and inactive)
    - Include calculated final_price for each variant
    - Order by: is_active DESC, created_at ASC
  - **Output:** List of ProductVariant entities

- [ ] **REQ-1.5:** Implement UpdateStock method (admin only)
  - **Input:** variantID (uuid.UUID), newStockQuantity (int)
  - **Validation:**
    - Variant must exist
    - newStockQuantity >= 0
  - **Business Logic:**
    - Update stock_quantity atomically
    - Log stock change for audit trail (future enhancement)
  - **Output:** Updated variant or error

- [ ] **REQ-1.6:** Implement DeleteVariant method (admin only)
  - **Input:** variantID (uuid.UUID)
  - **Validation:**
    - Variant must exist
    - Check if variant is in any active cart (if yes, soft delete only)
  - **Business Logic:**
    - Soft delete: Set is_active = false
    - Hard delete: Only if not referenced in carts/orders
  - **Output:** Success or error

**Acceptance Criteria:**
- All methods handle context cancellation
- All errors are domain-specific (not database errors)
- All calculations are accurate (final_price)
- Unit tests achieve 80%+ coverage

---

### REQ-2: Stock Validator Service

**File:** `internal/application/services/stock_validator_service.go`

**Requirements:**

- [ ] **REQ-2.1:** Create StockValidatorService struct
  - **Dependencies:** ProductVariantRepository

- [ ] **REQ-2.2:** Implement ValidateStock method
  - **Input:** variantID (uuid.UUID), requestedQty (int)
  - **Validation:**
    - Variant must exist
    - Variant must be active (is_active = true)
    - variant.stock_quantity >= requestedQty
  - **Output:** nil (success) or specific error

- [ ] **REQ-2.3:** Implement ValidateCartStock method
  - **Input:** cartItems ([]*entities.CartItem)
  - **Business Logic:**
    - For each cart item, validate stock availability
    - Collect all validation errors (don't fail fast)
    - Return structured error list with item details
  - **Output:** []StockValidationError or nil

- [ ] **REQ-2.4:** Define custom errors
  ```go
  type ErrVariantNotFound struct { VariantID uuid.UUID }
  type ErrVariantInactive struct { VariantID uuid.UUID }
  type ErrInsufficientStock struct {
      VariantID uuid.UUID
      Available int
      Requested int
  }
  type StockValidationError struct {
      CartItemID uuid.UUID
      VariantID  uuid.UUID
      Available  int
      InCart     int
      Message    string
  }
  ```

**Acceptance Criteria:**
- Errors contain actionable information for frontend display
- Validation is atomic (no partial updates)
- Performance: O(n) complexity where n = number of cart items
- Unit tests cover all error scenarios

---

### REQ-3: Price Calculator Service

**File:** `internal/application/services/price_calculator_service.go`

**Requirements:**

- [ ] **REQ-3.1:** Create PriceCalculatorService struct
  - **Dependencies:** None (pure calculation service)

- [ ] **REQ-3.2:** Implement CalculateVariantPrice method
  - **Input:** basePrice (float64), priceAdjustment (float64)
  - **Formula:** `finalPrice = basePrice + priceAdjustment`
  - **Output:** float64 (finalPrice)

- [ ] **REQ-3.3:** Implement CalculateCartSubtotal method
  - **Input:** cartItems ([]*entities.CartItem)
  - **Formula:** `subtotal = sum(item.price_snapshot * item.quantity)`
  - **Output:** float64 (subtotal)

- [ ] **REQ-3.4:** Implement CalculateItemSubtotal method
  - **Input:** priceSnapshot (float64), quantity (int)
  - **Formula:** `itemSubtotal = priceSnapshot * quantity`
  - **Output:** float64 (itemSubtotal)

- [ ] **REQ-3.5:** Implement GetPriceRange method
  - **Input:** variants ([]*entities.ProductVariant), basePrice (float64)
  - **Business Logic:**
    - Calculate final_price for each active variant
    - Return min and max prices
  - **Output:** (minPrice float64, maxPrice float64)

**Acceptance Criteria:**
- All calculations use float64 (no rounding errors)
- Handle empty input gracefully (return 0.0)
- Pure functions (no side effects)
- 100% test coverage

---

### REQ-4: Cart Service

**File:** `internal/application/services/cart_service.go`

**Requirements:**

- [ ] **REQ-4.1:** Create CartService struct
  - **Dependencies:** CartRepository, ProductVariantRepository, StockValidatorService, PriceCalculatorService

- [ ] **REQ-4.2:** Implement GetOrCreateCart method
  - **Input:** userID (uuid.UUID)
  - **Business Logic:**
    - Check if user has existing cart
    - If not, create new cart
    - Return cart with all items (with product & variant details)
  - **Output:** *entities.CartDetail or error

- [ ] **REQ-4.3:** Implement AddItemToCart method
  - **Input:** userID (uuid.UUID), variantID (uuid.UUID), quantity (int)
  - **Validation:**
    - Variant exists and is active (use StockValidatorService)
    - Stock available for requested quantity
  - **Business Logic:**
    1. Get or create cart for user
    2. Check if variant already in cart
       - If yes: Update quantity (existing + new)
       - If no: Create new cart item
    3. Get variant's current final_price
    4. Save as price_snapshot
    5. Re-validate total stock after update
  - **Output:** *entities.CartItem or error

- [ ] **REQ-4.4:** Implement UpdateCartItemQuantity method
  - **Input:** userID (uuid.UUID), cartItemID (uuid.UUID), newQuantity (int)
  - **Validation:**
    - Cart item exists
    - Cart item belongs to user (authorization check)
    - If newQuantity = 0, delete item
    - If newQuantity > 0, validate stock availability
  - **Business Logic:**
    - Update quantity
    - If newQuantity = 0, call RemoveCartItem instead
  - **Output:** *entities.CartItem or error

- [ ] **REQ-4.5:** Implement RemoveCartItem method
  - **Input:** userID (uuid.UUID), cartItemID (uuid.UUID)
  - **Validation:**
    - Cart item exists
    - Cart item belongs to user
  - **Business Logic:**
    - Delete cart item permanently
  - **Output:** nil (success) or error

- [ ] **REQ-4.6:** Implement ClearCart method
  - **Input:** userID (uuid.UUID)
  - **Business Logic:**
    - Get user's cart
    - Delete all cart items
    - Keep cart entity (for future use)
  - **Output:** nil (success) or error

- [ ] **REQ-4.7:** Implement MergeGuestCart method
  - **Input:** userID (uuid.UUID), guestItems ([]GuestCartItem)
    ```go
    type GuestCartItem struct {
        ProductVariantID uuid.UUID
        Quantity         int
    }
    ```
  - **Validation:**
    - All variants exist and are active
    - Validate stock for all items (batch validation)
  - **Business Logic:**
    1. Get or create user's cart
    2. For each guest item:
       - If variant already in DB cart → Sum quantities
       - Else → Add new item
    3. Validate total stock for all merged items
    4. Return complete cart with all items
  - **Output:** *entities.CartDetail or []StockValidationError

- [ ] **REQ-4.8:** Implement GetCartWithDetails method
  - **Input:** userID (uuid.UUID)
  - **Business Logic:**
    - Fetch cart with items
    - Eager load: product_variant → product (with category & brand)
    - Calculate subtotal using PriceCalculatorService
    - Count total items
  - **Output:** *entities.CartDetail
    ```go
    type CartDetail struct {
        ID         uuid.UUID
        UserID     uuid.UUID
        Items      []*CartItemDetail
        Subtotal   float64
        ItemCount  int
        CreatedAt  time.Time
        UpdatedAt  time.Time
    }
    type CartItemDetail struct {
        ID              uuid.UUID
        Quantity        int
        PriceSnapshot   float64
        ItemSubtotal    float64
        ProductVariant  *entities.ProductVariant
        Product         *entities.Product
    }
    ```

**Acceptance Criteria:**
- All cart operations are atomic (use transactions)
- Authorization checks prevent cross-user cart access
- Stock validation prevents overselling
- Price snapshot is immutable after item creation
- MergeGuestCart handles duplicates correctly
- Unit tests cover all error paths
- Integration tests verify cart workflows

---

### REQ-5: Enhanced Product Service

**File:** `internal/application/services/product_service.go` (update existing)

**Requirements:**

- [ ] **REQ-5.1:** Implement SearchProducts method
  - **Input:** SearchProductsParams
    ```go
    type SearchProductsParams struct {
        Search    string      // Full-text search query
        CategoryID *uuid.UUID // Filter by category
        BrandID    *uuid.UUID // Filter by brand
        MinPrice   *float64   // Min base_price
        MaxPrice   *float64   // Max base_price
        Size       *string    // Filter by variant attribute
        Color      *string    // Filter by variant attribute
        Status     *string    // published, draft, archived
        SortBy     string     // name, price, created_at
        SortOrder  string     // asc, desc
        Page       int        // Page number (1-indexed)
        Limit      int        // Items per page (max 100)
    }
    ```
  - **Business Logic:**
    - If Search is provided: Use PostgreSQL ts_rank() for relevance sorting
    - Apply all filters (category, brand, price, status)
    - Apply variant attribute filters (JSONB queries on ProductVariant)
    - Calculate min_price and max_price from variants
    - Check has_stock (any variant with stock > 0)
    - Apply sorting and pagination
  - **Output:** ProductSearchResult
    ```go
    type ProductSearchResult struct {
        Products   []*ProductListItem
        Pagination PaginationInfo
    }
    type ProductListItem struct {
        ID            uuid.UUID
        Slug          string
        Name          string
        BasePrice     float64
        Description   string
        ImageURLs     []string
        Status        string
        Category      *CategorySummary
        Brand         *BrandSummary
        VariantsCount int
        MinPrice      float64
        MaxPrice      float64
        HasStock      bool
    }
    ```

- [ ] **REQ-5.2:** Implement GetProductWithVariants method
  - **Input:** productID (uuid.UUID)
  - **Business Logic:**
    - Fetch product with category & brand
    - Eager load all variants (active and inactive)
    - Calculate final_price for each variant
    - Set is_in_stock = (stock_quantity > 0)
  - **Output:** *ProductDetail
    ```go
    type ProductDetail struct {
        *entities.Product
        Category *entities.Category
        Brand    *entities.Brand
        Variants []*ProductVariantDetail
    }
    type ProductVariantDetail struct {
        ID              uuid.UUID
        SKU             string
        Attributes      map[string]string
        StockQuantity   int
        PriceAdjustment float64
        FinalPrice      float64
        IsActive        bool
        IsInStock       bool
    }
    ```

**Acceptance Criteria:**
- Search returns results sorted by relevance (ts_rank)
- Filters work in combination (category + price + size)
- Pagination is accurate (total count, total pages)
- Performance: < 200ms (p95) for 50k products
- Empty search returns all published products
- GetProductWithVariants includes inactive variants for admin

---

### REQ-6: PostgreSQL Full-Text Search

**File:** `migrations/YYYYMMDD_add_product_search.sql`

**Requirements:**

- [ ] **REQ-6.1:** Create migration file
  - **File naming:** Use timestamp format (e.g., `20260101_add_product_search.sql`)

- [ ] **REQ-6.2:** Add search_vector column
  ```sql
  ALTER TABLE products ADD COLUMN search_vector tsvector;
  ```

- [ ] **REQ-6.3:** Create trigger function
  ```sql
  CREATE OR REPLACE FUNCTION products_search_vector_update() RETURNS trigger AS $$
  BEGIN
      NEW.search_vector :=
          setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
          setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;
  ```

- [ ] **REQ-6.4:** Create trigger
  ```sql
  CREATE TRIGGER products_search_vector_trigger
  BEFORE INSERT OR UPDATE ON products
  FOR EACH ROW EXECUTE FUNCTION products_search_vector_update();
  ```

- [ ] **REQ-6.5:** Create GIN index
  ```sql
  CREATE INDEX products_search_vector_idx ON products USING GIN(search_vector);
  ```

- [ ] **REQ-6.6:** Update existing products
  ```sql
  UPDATE products SET search_vector =
      setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
      setweight(to_tsvector('english', COALESCE(description, '')), 'B');
  ```

**Acceptance Criteria:**
- Migration is idempotent (can be run multiple times)
- Trigger auto-updates search_vector on INSERT/UPDATE
- Index improves search performance (verify with EXPLAIN ANALYZE)
- Rollback script provided

---

### REQ-7: Product Variant Handlers (API Layer)

**File:** `internal/adapters/handlers/product_variant_handler.go`

**Requirements:**

- [ ] **REQ-7.1:** Create ProductVariantHandler struct
  - **Dependencies:** ProductVariantService, ProductService

- [ ] **REQ-7.2:** Define DTOs
  ```go
  // Request DTOs
  type CreateProductVariantRequest struct {
      SKU             string            `json:"sku" validate:"required,min=3,max=100"`
      Attributes      map[string]string `json:"attributes" validate:"required"`
      StockQuantity   int               `json:"stock_quantity" validate:"min=0"`
      PriceAdjustment float64           `json:"price_adjustment"`
      IsActive        *bool             `json:"is_active"`
  }

  type UpdateProductVariantRequest struct {
      SKU             *string            `json:"sku,omitempty" validate:"omitempty,min=3,max=100"`
      Attributes      *map[string]string `json:"attributes,omitempty"`
      StockQuantity   *int               `json:"stock_quantity,omitempty" validate:"omitempty,min=0"`
      PriceAdjustment *float64           `json:"price_adjustment,omitempty"`
      IsActive        *bool              `json:"is_active,omitempty"`
  }

  type UpdateStockRequest struct {
      StockQuantity int `json:"stock_quantity" validate:"min=0"`
  }

  // Response DTOs
  type ProductVariantResponse struct {
      ID              string            `json:"id"`
      ProductID       string            `json:"product_id"`
      SKU             string            `json:"sku"`
      Attributes      map[string]string `json:"attributes"`
      StockQuantity   int               `json:"stock_quantity"`
      PriceAdjustment float64           `json:"price_adjustment"`
      FinalPrice      float64           `json:"final_price"`
      IsActive        bool              `json:"is_active"`
      IsInStock       bool              `json:"is_in_stock"`
      CreatedAt       string            `json:"created_at"`
      UpdatedAt       string            `json:"updated_at"`
  }

  type ProductVariantListResponse struct {
      Data []*ProductVariantResponse `json:"data"`
  }
  ```

- [ ] **REQ-7.3:** Implement GET /api/products/:product_id/variants
  - **Method:** GET
  - **Auth:** Public
  - **Path Params:** product_id (UUID string)
  - **Response:** 200 ProductVariantListResponse
  - **Errors:**
    - 400: Invalid product_id format
    - 404: Product not found

- [ ] **REQ-7.4:** Implement POST /api/products/:product_id/variants
  - **Method:** POST
  - **Auth:** Required (Admin only)
  - **Path Params:** product_id (UUID string)
  - **Request Body:** CreateProductVariantRequest
  - **Response:** 201 ProductVariantResponse
  - **Errors:**
    - 400: Validation error
    - 401: Unauthorized
    - 403: Not admin
    - 404: Product not found
    - 409: SKU already exists

- [ ] **REQ-7.5:** Implement PUT /api/products/:product_id/variants/:variant_id
  - **Method:** PUT
  - **Auth:** Required (Admin only)
  - **Path Params:** product_id, variant_id (UUID strings)
  - **Request Body:** UpdateProductVariantRequest
  - **Response:** 200 ProductVariantResponse
  - **Errors:**
    - 400: Validation error
    - 401: Unauthorized
    - 403: Not admin
    - 404: Variant not found
    - 409: SKU already exists

- [ ] **REQ-7.6:** Implement PATCH /api/products/:product_id/variants/:variant_id/stock
  - **Method:** PATCH
  - **Auth:** Required (Admin only)
  - **Path Params:** product_id, variant_id (UUID strings)
  - **Request Body:** UpdateStockRequest
  - **Response:** 200 ProductVariantResponse
  - **Errors:**
    - 400: Validation error
    - 401: Unauthorized
    - 403: Not admin
    - 404: Variant not found

- [ ] **REQ-7.7:** Implement DELETE /api/products/:product_id/variants/:variant_id
  - **Method:** DELETE
  - **Auth:** Required (Admin only)
  - **Path Params:** product_id, variant_id (UUID strings)
  - **Response:** 204 No Content
  - **Errors:**
    - 401: Unauthorized
    - 403: Not admin
    - 404: Variant not found
    - 409: Variant in active cart/order

**Acceptance Criteria:**
- All handlers use Huma v2 for validation
- UUIDs are validated and parsed correctly
- Admin middleware prevents non-admin access
- OpenAPI documentation auto-generated
- Integration tests cover happy path and error cases

---

### REQ-8: Cart Handlers (API Layer)

**File:** `internal/adapters/handlers/cart_handler.go`

**Requirements:**

- [ ] **REQ-8.1:** Create CartHandler struct
  - **Dependencies:** CartService

- [ ] **REQ-8.2:** Define DTOs
  ```go
  // Request DTOs
  type AddCartItemRequest struct {
      ProductVariantID string `json:"product_variant_id" validate:"required,uuid"`
      Quantity         int    `json:"quantity" validate:"required,min=1"`
  }

  type UpdateCartItemRequest struct {
      Quantity int `json:"quantity" validate:"min=0"`
  }

  type MergeGuestCartRequest struct {
      Items []GuestCartItemDTO `json:"items" validate:"required,dive"`
  }

  type GuestCartItemDTO struct {
      ProductVariantID string `json:"product_variant_id" validate:"required,uuid"`
      Quantity         int    `json:"quantity" validate:"required,min=1"`
  }

  // Response DTOs
  type CartItemResponse struct {
      ID             string                  `json:"id"`
      Quantity       int                     `json:"quantity"`
      PriceSnapshot  float64                 `json:"price_snapshot"`
      Subtotal       float64                 `json:"subtotal"`
      ProductVariant *ProductVariantResponse `json:"product_variant"`
      Product        *ProductSummaryResponse `json:"product"`
  }

  type ProductSummaryResponse struct {
      ID        string   `json:"id"`
      Name      string   `json:"name"`
      Slug      string   `json:"slug"`
      ImageURLs []string `json:"image_urls"`
  }

  type CartResponse struct {
      ID        string              `json:"id"`
      UserID    string              `json:"user_id"`
      Items     []*CartItemResponse `json:"items"`
      Subtotal  float64             `json:"subtotal"`
      ItemCount int                 `json:"item_count"`
      CreatedAt string              `json:"created_at"`
      UpdatedAt string              `json:"updated_at"`
  }

  type StockValidationErrorResponse struct {
      CartItemID string `json:"cart_item_id,omitempty"`
      VariantID  string `json:"variant_id"`
      Available  int    `json:"available"`
      InCart     int    `json:"in_cart"`
      Message    string `json:"message"`
  }
  ```

- [ ] **REQ-8.3:** Implement GET /api/cart
  - **Method:** GET
  - **Auth:** Required (JWT)
  - **Response:** 200 CartResponse
  - **Business Logic:**
    - Get cart for authenticated user
    - If cart doesn't exist, create empty cart
  - **Errors:**
    - 401: Unauthorized

- [ ] **REQ-8.4:** Implement POST /api/cart/items
  - **Method:** POST
  - **Auth:** Required (JWT)
  - **Request Body:** AddCartItemRequest
  - **Response:** 201 CartItemResponse
  - **Business Logic:**
    - Extract user_id from JWT token
    - Call CartService.AddItemToCart
  - **Errors:**
    - 400: Validation error or insufficient stock
    - 401: Unauthorized
    - 404: Variant not found or inactive

- [ ] **REQ-8.5:** Implement PUT /api/cart/items/:item_id
  - **Method:** PUT
  - **Auth:** Required (JWT)
  - **Path Params:** item_id (UUID string)
  - **Request Body:** UpdateCartItemRequest
  - **Response:** 200 CartItemResponse (or 204 if deleted)
  - **Business Logic:**
    - Extract user_id from JWT token
    - Validate cart item belongs to user
    - If quantity = 0, delete item (return 204)
    - Else update quantity
  - **Errors:**
    - 400: Validation error or insufficient stock
    - 401: Unauthorized
    - 403: Cart item doesn't belong to user
    - 404: Cart item not found

- [ ] **REQ-8.6:** Implement DELETE /api/cart/items/:item_id
  - **Method:** DELETE
  - **Auth:** Required (JWT)
  - **Path Params:** item_id (UUID string)
  - **Response:** 204 No Content
  - **Business Logic:**
    - Extract user_id from JWT token
    - Validate cart item belongs to user
    - Delete cart item
  - **Errors:**
    - 401: Unauthorized
    - 403: Cart item doesn't belong to user
    - 404: Cart item not found

- [ ] **REQ-8.7:** Implement DELETE /api/cart
  - **Method:** DELETE
  - **Auth:** Required (JWT)
  - **Response:** 204 No Content
  - **Business Logic:**
    - Extract user_id from JWT token
    - Clear all items from user's cart
  - **Errors:**
    - 401: Unauthorized

- [ ] **REQ-8.8:** Implement POST /api/cart/merge
  - **Method:** POST
  - **Auth:** Required (JWT)
  - **Request Body:** MergeGuestCartRequest
  - **Response:** 200 CartResponse
  - **Business Logic:**
    - Extract user_id from JWT token
    - Call CartService.MergeGuestCart
    - If stock validation fails, return 400 with error details
  - **Errors:**
    - 400: Stock validation failed (return []StockValidationErrorResponse)
    - 401: Unauthorized
    - 404: Variant not found

**Acceptance Criteria:**
- All endpoints require JWT authentication
- User can only access their own cart items
- Stock validation errors return actionable messages
- OpenAPI documentation complete
- Integration tests verify cart workflows

---

### REQ-9: Enhanced Product Handlers (API Layer)

**File:** `internal/adapters/handlers/product_handler.go` (update existing)

**Requirements:**

- [ ] **REQ-9.1:** Update ProductHandler struct
  - **Add Dependency:** ProductVariantService

- [ ] **REQ-9.2:** Define DTOs
  ```go
  // Request DTOs
  type SearchProductsRequest struct {
      Search     string  `query:"search"`
      CategoryID *string `query:"category_id" validate:"omitempty,uuid"`
      BrandID    *string `query:"brand_id" validate:"omitempty,uuid"`
      MinPrice   *float64 `query:"min_price" validate:"omitempty,min=0"`
      MaxPrice   *float64 `query:"max_price" validate:"omitempty,min=0"`
      Size       *string `query:"size"`
      Color      *string `query:"color"`
      Status     *string `query:"status" validate:"omitempty,oneof=published draft archived"`
      SortBy     string  `query:"sort_by" validate:"omitempty,oneof=name price created_at"`
      SortOrder  string  `query:"sort_order" validate:"omitempty,oneof=asc desc"`
      Page       int     `query:"page" validate:"min=1"`
      Limit      int     `query:"limit" validate:"min=1,max=100"`
  }

  // Response DTOs
  type ProductListItemResponse struct {
      ID            string                `json:"id"`
      Slug          string                `json:"slug"`
      Name          string                `json:"name"`
      BasePrice     float64               `json:"base_price"`
      Description   string                `json:"description"`
      ImageURLs     []string              `json:"image_urls"`
      Status        string                `json:"status"`
      Category      *CategorySummary      `json:"category"`
      Brand         *BrandSummary         `json:"brand"`
      VariantsCount int                   `json:"variants_count"`
      MinPrice      float64               `json:"min_price"`
      MaxPrice      float64               `json:"max_price"`
      HasStock      bool                  `json:"has_stock"`
  }

  type ProductSearchResponse struct {
      Data       []*ProductListItemResponse `json:"data"`
      Pagination PaginationResponse         `json:"pagination"`
  }

  type PaginationResponse struct {
      Page       int `json:"page"`
      Limit      int `json:"limit"`
      Total      int `json:"total"`
      TotalPages int `json:"total_pages"`
  }

  type ProductDetailResponse struct {
      ID          string                       `json:"id"`
      Slug        string                       `json:"slug"`
      Name        string                       `json:"name"`
      BasePrice   float64                      `json:"base_price"`
      Description string                       `json:"description"`
      ImageURLs   []string                     `json:"image_urls"`
      Status      string                       `json:"status"`
      Weight      int                          `json:"weight"`
      Dimensions  DimensionsResponse           `json:"dimensions"`
      Category    *CategoryResponse            `json:"category"`
      Brand       *BrandResponse               `json:"brand"`
      Variants    []*ProductVariantResponse    `json:"variants"`
      CreatedAt   string                       `json:"created_at"`
      UpdatedAt   string                       `json:"updated_at"`
  }
  ```

- [ ] **REQ-9.3:** Implement GET /api/products (update existing)
  - **Method:** GET
  - **Auth:** Public
  - **Query Params:** SearchProductsRequest
  - **Response:** 200 ProductSearchResponse
  - **Business Logic:**
    - Default: page=1, limit=20, sort_by=created_at, sort_order=desc
    - If search provided: Sort by relevance (ts_rank) first
    - Apply all filters
    - Return only published products (unless admin)
  - **Errors:**
    - 400: Validation error (invalid UUID, etc.)

- [ ] **REQ-9.4:** Implement GET /api/products/:id (update existing)
  - **Method:** GET
  - **Auth:** Public
  - **Path Params:** id (UUID string)
  - **Response:** 200 ProductDetailResponse
  - **Business Logic:**
    - Fetch product with all variants
    - Include inactive variants only if admin
    - Calculate final_price for each variant
  - **Errors:**
    - 400: Invalid UUID format
    - 404: Product not found

- [ ] **REQ-9.5:** Update POST /api/products (existing endpoint)
  - **Remove field:** sku (no longer in Product schema)
  - **Rename field:** price → base_price
  - **Add field:** stock_quantity (optional, nullable)
  - **Business Logic:**
    - When creating product, also create default variant
    - Default variant: sku="PRODUCT-{slug}", attributes={}, stock_quantity=stock_quantity, price_adjustment=0

- [ ] **REQ-9.6:** Update PUT /api/products/:id (existing endpoint)
  - **Remove field:** sku
  - **Rename field:** price → base_price
  - **Business Logic:**
    - Updating base_price doesn't update variant prices (only affects new variants)

**Acceptance Criteria:**
- Search returns results sorted by relevance
- Filters work correctly (category + brand + price + variant attrs)
- Pagination metadata is accurate
- Performance: < 200ms (p95) with 50k products
- Default variant auto-created on product creation
- OpenAPI docs updated

---

## API Contract Specifications

### Authentication

All authenticated endpoints require JWT token in Authorization header:

```
Authorization: Bearer <jwt_token>
```

**JWT Token Claims:**
```json
{
  "user_id": "uuid-string",
  "email": "user@example.com",
  "role": "customer|admin",
  "exp": 1234567890,
  "iat": 1234567890
}
```

---

### API Endpoints Summary

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/products` | GET | Public | Search & list products |
| `/api/products/:id` | GET | Public | Get product detail with variants |
| `/api/products` | POST | Admin | Create product (auto-creates default variant) |
| `/api/products/:id` | PUT | Admin | Update product |
| `/api/products/:id` | DELETE | Admin | Delete product (cascade variants) |
| `/api/products/:product_id/variants` | GET | Public | List product variants |
| `/api/products/:product_id/variants` | POST | Admin | Create variant |
| `/api/products/:product_id/variants/:variant_id` | PUT | Admin | Update variant |
| `/api/products/:product_id/variants/:variant_id/stock` | PATCH | Admin | Update variant stock |
| `/api/products/:product_id/variants/:variant_id` | DELETE | Admin | Delete variant |
| `/api/cart` | GET | User | Get user's cart |
| `/api/cart/items` | POST | User | Add item to cart |
| `/api/cart/items/:item_id` | PUT | User | Update cart item quantity |
| `/api/cart/items/:item_id` | DELETE | User | Remove item from cart |
| `/api/cart` | DELETE | User | Clear cart |
| `/api/cart/merge` | POST | User | Merge guest cart on login |

---

### Detailed API Contracts

#### 1. GET /api/products

**Request:**
```http
GET /api/products?search=cotton&category_id=uuid&min_price=100000&max_price=500000&size=M&page=1&limit=20 HTTP/1.1
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "product-uuid-1",
      "slug": "cotton-tshirt",
      "name": "Cotton T-Shirt",
      "base_price": 100000,
      "description": "High quality cotton t-shirt",
      "image_urls": ["https://cdn.example.com/img1.jpg"],
      "status": "published",
      "category": {
        "id": "category-uuid",
        "name": "Tops",
        "slug": "tops"
      },
      "brand": {
        "id": "brand-uuid",
        "name": "Brand Name",
        "slug": "brand-name"
      },
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

**Error Responses:**
- `400 Bad Request`: Invalid query parameters
  ```json
  {
    "error": "validation_error",
    "message": "Invalid category_id: must be a valid UUID",
    "details": {
      "field": "category_id",
      "value": "invalid-uuid"
    }
  }
  ```

---

#### 2. GET /api/products/:id

**Request:**
```http
GET /api/products/product-uuid-1 HTTP/1.1
```

**Response (200 OK):**
```json
{
  "id": "product-uuid-1",
  "slug": "cotton-tshirt",
  "name": "Cotton T-Shirt",
  "base_price": 100000,
  "description": "High quality cotton t-shirt",
  "image_urls": ["https://cdn.example.com/img1.jpg", "https://cdn.example.com/img2.jpg"],
  "status": "published",
  "weight": 200,
  "dimensions": {
    "length": 30,
    "width": 20,
    "height": 2
  },
  "category": {
    "id": "category-uuid",
    "name": "Tops",
    "slug": "tops",
    "description": "Top clothing"
  },
  "brand": {
    "id": "brand-uuid",
    "name": "Brand Name",
    "slug": "brand-name",
    "description": "Brand description"
  },
  "variants": [
    {
      "id": "variant-uuid-1",
      "product_id": "product-uuid-1",
      "sku": "TSHIRT-S-BLACK",
      "attributes": {
        "size": "S",
        "color": "Black"
      },
      "stock_quantity": 50,
      "price_adjustment": 0,
      "final_price": 100000,
      "is_active": true,
      "is_in_stock": true,
      "created_at": "2025-12-31T10:00:00Z",
      "updated_at": "2025-12-31T10:00:00Z"
    },
    {
      "id": "variant-uuid-2",
      "product_id": "product-uuid-1",
      "sku": "TSHIRT-XL-WHITE",
      "attributes": {
        "size": "XL",
        "color": "White"
      },
      "stock_quantity": 0,
      "price_adjustment": 10000,
      "final_price": 110000,
      "is_active": true,
      "is_in_stock": false,
      "created_at": "2025-12-31T10:00:00Z",
      "updated_at": "2025-12-31T10:00:00Z"
    }
  ],
  "created_at": "2025-12-31T10:00:00Z",
  "updated_at": "2025-12-31T10:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid UUID format
- `404 Not Found`: Product not found

---

#### 3. POST /api/products/:product_id/variants

**Request:**
```http
POST /api/products/product-uuid-1/variants HTTP/1.1
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json

{
  "sku": "TSHIRT-M-RED",
  "attributes": {
    "size": "M",
    "color": "Red"
  },
  "stock_quantity": 30,
  "price_adjustment": 0,
  "is_active": true
}
```

**Response (201 Created):**
```json
{
  "id": "variant-uuid-3",
  "product_id": "product-uuid-1",
  "sku": "TSHIRT-M-RED",
  "attributes": {
    "size": "M",
    "color": "Red"
  },
  "stock_quantity": 30,
  "price_adjustment": 0,
  "final_price": 100000,
  "is_active": true,
  "is_in_stock": true,
  "created_at": "2026-01-01T12:00:00Z",
  "updated_at": "2026-01-01T12:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Validation error
  ```json
  {
    "error": "validation_error",
    "message": "stock_quantity must be >= 0"
  }
  ```
- `401 Unauthorized`: Missing or invalid JWT token
- `403 Forbidden`: User is not admin
- `404 Not Found`: Product not found
- `409 Conflict`: SKU already exists
  ```json
  {
    "error": "duplicate_sku",
    "message": "SKU 'TSHIRT-M-RED' already exists"
  }
  ```

---

#### 4. GET /api/cart

**Request:**
```http
GET /api/cart HTTP/1.1
Authorization: Bearer <user-jwt-token>
```

**Response (200 OK):**
```json
{
  "id": "cart-uuid",
  "user_id": "user-uuid",
  "items": [
    {
      "id": "cart-item-uuid-1",
      "quantity": 2,
      "price_snapshot": 100000,
      "subtotal": 200000,
      "product_variant": {
        "id": "variant-uuid-1",
        "sku": "TSHIRT-M-BLACK",
        "attributes": {
          "size": "M",
          "color": "Black"
        },
        "stock_quantity": 50,
        "price_adjustment": 0,
        "final_price": 100000,
        "is_active": true,
        "is_in_stock": true
      },
      "product": {
        "id": "product-uuid-1",
        "name": "Cotton T-Shirt",
        "slug": "cotton-tshirt",
        "image_urls": ["https://cdn.example.com/img1.jpg"]
      }
    }
  ],
  "subtotal": 200000,
  "item_count": 2,
  "created_at": "2026-01-01T10:00:00Z",
  "updated_at": "2026-01-01T11:30:00Z"
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid JWT token

---

#### 5. POST /api/cart/items

**Request:**
```http
POST /api/cart/items HTTP/1.1
Authorization: Bearer <user-jwt-token>
Content-Type: application/json

{
  "product_variant_id": "variant-uuid-2",
  "quantity": 1
}
```

**Response (201 Created):**
```json
{
  "id": "cart-item-uuid-2",
  "quantity": 1,
  "price_snapshot": 110000,
  "subtotal": 110000,
  "product_variant": {
    "id": "variant-uuid-2",
    "sku": "TSHIRT-XL-WHITE",
    "attributes": {
      "size": "XL",
      "color": "White"
    },
    "stock_quantity": 20,
    "price_adjustment": 10000,
    "final_price": 110000,
    "is_active": true,
    "is_in_stock": true
  },
  "product": {
    "id": "product-uuid-1",
    "name": "Cotton T-Shirt",
    "slug": "cotton-tshirt",
    "image_urls": ["https://cdn.example.com/img1.jpg"]
  }
}
```

**Error Responses:**
- `400 Bad Request`: Insufficient stock
  ```json
  {
    "error": "insufficient_stock",
    "message": "Insufficient stock for variant 'TSHIRT-XL-WHITE'",
    "details": {
      "variant_id": "variant-uuid-2",
      "available": 5,
      "requested": 10
    }
  }
  ```
- `401 Unauthorized`: Missing or invalid JWT token
- `404 Not Found`: Variant not found or inactive
  ```json
  {
    "error": "variant_not_found",
    "message": "Product variant not found or inactive"
  }
  ```

---

#### 6. POST /api/cart/merge

**Request:**
```http
POST /api/cart/merge HTTP/1.1
Authorization: Bearer <user-jwt-token>
Content-Type: application/json

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

**Response (200 OK):**
```json
{
  "id": "cart-uuid",
  "user_id": "user-uuid",
  "items": [
    {
      "id": "cart-item-uuid-1",
      "quantity": 4,
      "price_snapshot": 100000,
      "subtotal": 400000,
      "product_variant": { ... },
      "product": { ... }
    },
    {
      "id": "cart-item-uuid-2",
      "quantity": 1,
      "price_snapshot": 110000,
      "subtotal": 110000,
      "product_variant": { ... },
      "product": { ... }
    }
  ],
  "subtotal": 510000,
  "item_count": 5,
  "created_at": "2026-01-01T10:00:00Z",
  "updated_at": "2026-01-01T12:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Stock validation failed
  ```json
  {
    "error": "stock_validation_failed",
    "message": "Some items have insufficient stock",
    "details": [
      {
        "variant_id": "variant-uuid-2",
        "available": 5,
        "in_cart": 10,
        "message": "Insufficient stock for variant 'TSHIRT-XL-WHITE'"
      }
    ]
  }
  ```
- `401 Unauthorized`: Missing or invalid JWT token
- `404 Not Found`: Variant not found

---

## Acceptance Criteria

### Functional Acceptance Criteria

**Product Variants:**
- ✅ Can create product with variants
- ✅ Can update variant stock independently
- ✅ Variant price calculated correctly (base_price + adjustment)
- ✅ SKU uniqueness enforced across all variants
- ✅ Can filter products by variant attributes (size, color)
- ✅ Cannot delete variant if in active cart/order

**Cart System:**
- ✅ User can add item to cart with stock validation
- ✅ User can update item quantity with stock check
- ✅ User can remove item from cart
- ✅ User can clear entire cart
- ✅ Guest cart merges correctly on login
- ✅ Duplicate variants in cart are summed (not duplicated)
- ✅ Price snapshot is immutable after item creation
- ✅ User cannot access other users' carts

**Product Search:**
- ✅ Full-text search returns relevant results
- ✅ Search results sorted by relevance (ts_rank)
- ✅ Can filter by category, brand, price range, variant attributes
- ✅ Can sort by name, price, created_at (asc/desc)
- ✅ Pagination works correctly (total count, total pages)
- ✅ Empty search returns all published products

**Stock Validation:**
- ✅ Cannot add item to cart if stock insufficient
- ✅ Cannot update cart item quantity beyond stock
- ✅ Stock validation errors return actionable messages
- ✅ Cart merge validates stock for all items

### Technical Acceptance Criteria

**Performance:**
- ✅ Product search: < 200ms (p95) with 50k products
- ✅ Cart operations: < 100ms (p95)
- ✅ Product detail: < 150ms (p95)

**Testing:**
- ✅ Unit tests: 80%+ coverage
- ✅ Integration tests: All API endpoints covered
- ✅ All error paths tested
- ✅ Edge cases covered (empty cart, out of stock, etc.)

**API Quality:**
- ✅ OpenAPI documentation auto-generated (Huma v2)
- ✅ All request/response schemas documented
- ✅ Error responses standardized
- ✅ HTTP status codes used correctly

**Security:**
- ✅ JWT authentication enforced on cart endpoints
- ✅ Admin authorization enforced on variant management
- ✅ User can only access their own cart
- ✅ SQL injection prevented (parameterized queries)
- ✅ JSONB injection prevented (input validation)

**Code Quality:**
- ✅ Follows hexagonal architecture pattern
- ✅ Clear separation: Domain → Application → Adapters
- ✅ Repository interfaces in domain/ports
- ✅ No business logic in handlers
- ✅ Errors are domain-specific (not database errors)

---

## Testing Requirements

### Unit Tests

**Services to Test:**
1. **ProductVariantService:**
   - CreateVariant (success, duplicate SKU, invalid product)
   - UpdateVariant (success, not found, duplicate SKU)
   - GetVariantsByProduct (success, empty list)
   - UpdateStock (success, negative stock, not found)
   - DeleteVariant (success, in cart, not found)

2. **CartService:**
   - AddItemToCart (new item, existing item, insufficient stock)
   - UpdateCartItemQuantity (success, delete on zero, insufficient stock)
   - RemoveCartItem (success, not owned, not found)
   - ClearCart (success, empty cart)
   - MergeGuestCart (success, duplicates, stock validation)

3. **StockValidatorService:**
   - ValidateStock (success, insufficient, inactive, not found)
   - ValidateCartStock (all valid, some invalid, all invalid)

4. **PriceCalculatorService:**
   - CalculateVariantPrice (positive adjustment, negative adjustment)
   - CalculateCartSubtotal (empty cart, multiple items)
   - GetPriceRange (single variant, multiple variants)

**Coverage Target:** 80%+ for all services

---

### Integration Tests

**API Endpoints to Test:**

1. **Product Search API:**
   - Search with query (relevance sorting)
   - Filter by category (single, multiple)
   - Filter by brand
   - Filter by price range
   - Filter by variant attributes (size, color)
   - Combined filters (category + price + size)
   - Sorting (name, price, created_at)
   - Pagination (first page, middle page, last page)
   - Empty results
   - Invalid query parameters

2. **Product Variant API:**
   - Create variant (success, duplicate SKU, invalid product)
   - Update variant (success, not found)
   - Update stock (success, negative stock)
   - Delete variant (success, in cart)
   - List variants (success, empty list)

3. **Cart API:**
   - Get cart (empty, with items)
   - Add item (new, existing, insufficient stock)
   - Update item quantity (increase, decrease, delete on zero)
   - Remove item (success, not owned)
   - Clear cart (success)
   - Merge guest cart (success, duplicates, stock validation)

**Test Data:**
- Setup: Seed database with 100+ products, 500+ variants
- Teardown: Clean up after each test
- Isolation: Each test uses separate user accounts

---

### Performance Tests

**Benchmarks:**

1. **Product Search:**
   - 50k products, 200k variants
   - Full-text search: < 200ms (p95)
   - Filter by category + price: < 150ms (p95)
   - JSONB attribute filter: < 250ms (p95)

2. **Cart Operations:**
   - Add item: < 100ms (p95)
   - Update item: < 80ms (p95)
   - Get cart with 50 items: < 200ms (p95)

3. **Database Queries:**
   - Verify GIN index usage for search (EXPLAIN ANALYZE)
   - Verify JSONB index usage for attribute filters
   - No N+1 queries (use eager loading)

**Tools:**
- Use `pgbench` for database benchmarks
- Use `wrk` or `ab` for HTTP benchmarks
- Monitor with PostgreSQL EXPLAIN ANALYZE

---

## Non-Functional Requirements

### NFR-1: Scalability
- System must handle 50k+ products without performance degradation
- Cart operations must support 10k+ concurrent users
- Database schema must support horizontal scaling (partitioning ready)

### NFR-2: Reliability
- All cart operations must be atomic (use database transactions)
- Stock validation must prevent overselling (race condition safe)
- Price snapshot ensures price integrity during checkout

### NFR-3: Maintainability
- Code follows hexagonal architecture (easy to test & swap adapters)
- Business logic isolated in domain/application layers
- Clear error messages for debugging
- OpenAPI documentation auto-generated (always up-to-date)

### NFR-4: Security
- JWT authentication required for cart endpoints
- Admin authorization required for variant management
- SQL injection prevented (parameterized queries)
- JSONB injection prevented (input validation)
- Rate limiting on search endpoint (prevent abuse)

### NFR-5: Observability
- All services log errors with context (user_id, product_id, etc.)
- Stock validation failures logged for audit
- Performance metrics tracked (p95, p99 latencies)
- Database query performance monitored

### NFR-6: Data Integrity
- SKU uniqueness enforced at database level (unique constraint)
- One cart per user enforced (unique constraint)
- No duplicate variants in cart (composite unique constraint)
- Foreign key constraints prevent orphaned records
- Stock quantity cannot be negative (check constraint)

---

## Appendix

### A. Database Schema Summary

**Tables:**
- `products` (updated: removed sku, added base_price, stock_quantity, search_vector)
- `product_variants` (new: sku, attributes JSONB, stock_quantity, price_adjustment)
- `carts` (new: user_id)
- `cart_items` (new: cart_id, product_variant_id, quantity, price_snapshot)

**Indexes:**
- `products.search_vector` (GIN index for full-text search)
- `product_variants.sku` (unique)
- `product_variants.product_id` (foreign key index)
- `carts.user_id` (unique)
- `cart_items(cart_id, product_variant_id)` (composite unique)

### B. Error Code Reference

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `validation_error` | 400 | Request validation failed |
| `insufficient_stock` | 400 | Not enough stock for requested quantity |
| `stock_validation_failed` | 400 | Cart stock validation failed |
| `duplicate_sku` | 409 | SKU already exists |
| `variant_not_found` | 404 | Product variant not found or inactive |
| `product_not_found` | 404 | Product not found |
| `cart_item_not_found` | 404 | Cart item not found |
| `unauthorized` | 401 | Missing or invalid JWT token |
| `forbidden` | 403 | User not authorized (not admin or not owner) |
| `variant_in_use` | 409 | Cannot delete variant (in active cart/order) |

### C. Implementation Timeline

**Day 1-2:** Database & Schema
- ✅ COMPLETED: Schemas created, migrations run

**Day 3-4:** Services
- ProductVariantService
- StockValidatorService
- PriceCalculatorService
- CartService
- Enhanced ProductService

**Day 5-6:** Handlers (API Layer)
- ProductVariant handlers
- Cart handlers
- Enhanced Product handlers
- DTOs

**Day 7-8:** Testing
- Unit tests (services)
- Integration tests (API endpoints)
- Performance benchmarks

**Day 9-10:** Polish & Documentation
- OpenAPI documentation review
- Error handling consistency
- Code review & refactoring
- Final testing

---

**Document Status:** ✅ Ready for Implementation
**Approved By:** [Pending]
**Implementation Start:** 2026-01-01
**Target Completion:** 2026-01-10 (10 days)
**Next Review:** Daily standup @ 9:00 AM
