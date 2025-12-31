# Phase 2 Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement Product Catalog with Variants and Cart System for e-commerce backend

**Architecture:** Hexagonal architecture with Ent ORM, JSONB for flexible variant attributes, PostgreSQL full-text search, and database-backed cart for authenticated users.

**Tech Stack:** Go 1.23+, GoFiber, Huma v2, Ent ORM, PostgreSQL, JSONB, Full-Text Search

---

## Task 1: Update Product Schema (Remove SKU, Rename Price)

**Files:**
- Modify: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/schema/product.go`

**Step 1: Remove SKU field and rename price to base_price**

Update the `Fields()` function:

```go
func (Product) Fields() []ent.Field {
    return []ent.Field{
        // REMOVE: field.String("sku")... - Delete this entire field

        field.String("slug").
            NotEmpty().
            Unique().
            Comment("URL-friendly identifier"),

        field.String("name").
            NotEmpty().
            Comment("Product name"),

        // RENAME: price → base_price
        field.Float("base_price").
            Positive().
            Comment("Base price for price calculation with variants"),

        field.Text("description").
            Optional().
            Comment("Product description"),

        // Add new fields after description
        field.Int("stock_quantity").
            NonNegative().
            Optional().
            Nillable().
            Comment("Stock for non-variant products, NULL for variant products"),

        field.Int("low_stock_threshold").
            NonNegative().
            Default(10).
            Comment("Alert when stock below this number"),

        field.Int("weight").
            NonNegative().
            Default(0).
            Comment("Weight in grams for courier calculation"),

        // ... rest of existing fields remain unchanged
    }
}
```

**Step 2: Add variants edge to Product**

Update the `Edges()` function:

```go
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).
            Ref("products").
            Unique().
            Field("category_id"),

        edge.From("brand", Brand.Type).
            Ref("products").
            Unique().
            Field("brand_id"),

        // NEW: Add variants edge
        edge.To("variants", ProductVariant.Type),
    }
}
```

**Step 3: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/schema/product.go
git commit -m "refactor: update Product schema - remove SKU, rename price to base_price, add stock fields"
```

---

## Task 2: Create ProductVariant Schema

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/schema/product_variant.go`

**Step 1: Create ProductVariant schema file**

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

// ProductVariant holds the schema definition for the ProductVariant entity.
type ProductVariant struct {
	ent.Schema
}

// Fields of the ProductVariant.
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

// Edges of the ProductVariant.
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

// Indexes of the ProductVariant.
func (ProductVariant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("sku").Unique(),
	}
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/schema/product_variant.go
git commit -m "feat: add ProductVariant schema with JSONB attributes"
```

---

## Task 3: Create Cart Schema

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/schema/cart.go`

**Step 1: Create Cart schema file**

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

// Cart holds the schema definition for the Cart entity.
type Cart struct {
	ent.Schema
}

// Fields of the Cart.
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

// Edges of the Cart.
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

// Indexes of the Cart.
func (Cart) Indexes() []ent.Index {
	return []ent.Index{
		// One cart per user
		index.Fields("user_id").Unique(),
	}
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/schema/cart.go
git commit -m "feat: add Cart schema with unique user constraint"
```

---

## Task 4: Create CartItem Schema

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/schema/cart_item.go`

**Step 1: Create CartItem schema file**

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

// CartItem holds the schema definition for the CartItem entity.
type CartItem struct {
	ent.Schema
}

// Fields of the CartItem.
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

// Edges of the CartItem.
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

// Indexes of the CartItem.
func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id"),
		index.Fields("product_variant_id"),
		// Prevent duplicate variant in same cart
		index.Fields("cart_id", "product_variant_id").Unique(),
	}
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/schema/cart_item.go
git commit -m "feat: add CartItem schema with unique cart-variant constraint"
```

---

## Task 5: Update User Schema (Add Cart Edge)

**Files:**
- Modify: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/schema/user.go`

**Step 1: Add cart edge to User**

Update the `Edges()` function:

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

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/schema/user.go
git commit -m "feat: add cart edge to User schema"
```

---

## Task 6: Generate Ent Code

**Files:**
- Generated files in: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/db/ent/`

**Step 1: Run Ent code generation**

```bash
cd /Volumes/External/Work/personal/go-yippi
go generate ./internal/adapters/persistence/db/schema
```

Expected output: Successfully generated Ent client code

**Step 2: Verify generated files exist**

```bash
ls internal/adapters/persistence/db/ent/ | grep -E "(productvariant|cart|cartitem)"
```

Expected output: Should show productvariant.go, cart.go, cartitem.go files

**Step 3: Run go mod tidy**

```bash
cd /Volumes/External/Work/personal/go-yippi
go mod tidy
```

**Step 4: Commit generated files**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/db/ent/
git commit -m "build: generate Ent code for new schemas"
```

---

## Task 7: Create ProductVariant Entity

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/domain/entities/product_variant.go`

**Step 1: Create ProductVariant entity**

```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProductVariant represents a variant of a product with specific attributes
type ProductVariant struct {
	ID              uuid.UUID         `json:"id"`
	ProductID       uuid.UUID         `json:"product_id"`
	SKU             string            `json:"sku"`
	Attributes      map[string]string `json:"attributes"` // {"size": "M", "color": "Red"}
	StockQuantity   int               `json:"stock_quantity"`
	PriceAdjustment float64           `json:"price_adjustment"`
	IsActive        bool              `json:"is_active"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CalculateFinalPrice calculates the final price based on product base price
func (v *ProductVariant) CalculateFinalPrice(basePrice float64) float64 {
	return basePrice + v.PriceAdjustment
}

// IsInStock checks if the variant has stock available
func (v *ProductVariant) IsInStock() bool {
	return v.StockQuantity > 0
}

// IsAvailable checks if variant is active and in stock
func (v *ProductVariant) IsAvailable() bool {
	return v.IsActive && v.IsInStock()
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/domain/entities/product_variant.go
git commit -m "feat: add ProductVariant entity with price calculation"
```

---

## Task 8: Create Cart Entities

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/domain/entities/cart.go`

**Step 1: Create Cart and CartItem entities**

```go
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
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/domain/entities/cart.go
git commit -m "feat: add Cart and CartItem entities with calculations"
```

---

## Task 9: Create ProductVariant Repository Interface

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/domain/ports/product_variant_repository.go`

**Step 1: Create repository interface**

```go
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/go-yippi/internal/domain/entities"
)

// ProductVariantRepository defines the interface for product variant persistence
type ProductVariantRepository interface {
	// Create creates a new product variant
	Create(ctx context.Context, variant *entities.ProductVariant) error

	// FindByID finds a variant by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error)

	// FindBySKU finds a variant by SKU
	FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error)

	// FindByProductID finds all variants for a product
	FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error)

	// Update updates an existing variant
	Update(ctx context.Context, variant *entities.ProductVariant) error

	// UpdateStock updates stock quantity for a variant
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error

	// Delete deletes a variant
	Delete(ctx context.Context, id uuid.UUID) error
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/domain/ports/product_variant_repository.go
git commit -m "feat: add ProductVariant repository interface"
```

---

## Task 10: Create Cart Repository Interface

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/domain/ports/cart_repository.go`

**Step 1: Create repository interface**

```go
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/go-yippi/internal/domain/entities"
)

// CartRepository defines the interface for cart persistence
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
	GetCartWithItems(ctx context.Context, userID uuid.UUID) (*entities.CartDetail, error)

	// Validation
	ItemExists(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (bool, error)
	GetCartItemByVariant(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (*entities.CartItem, error)
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/domain/ports/cart_repository.go
git commit -m "feat: add Cart repository interface"
```

---

## Task 11: Implement ProductVariant Repository (Ent Adapter)

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/product_variant_repository.go`

**Step 1: Create repository implementation**

```go
package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/go-yippi/internal/adapters/persistence/db/ent"
	"github.com/yourusername/go-yippi/internal/adapters/persistence/db/ent/productvariant"
	"github.com/yourusername/go-yippi/internal/domain/entities"
)

// ProductVariantRepository implements ports.ProductVariantRepository
type ProductVariantRepository struct {
	client *ent.Client
}

// NewProductVariantRepository creates a new product variant repository
func NewProductVariantRepository(client *ent.Client) *ProductVariantRepository {
	return &ProductVariantRepository{client: client}
}

// Create creates a new product variant
func (r *ProductVariantRepository) Create(ctx context.Context, variant *entities.ProductVariant) error {
	created, err := r.client.ProductVariant.
		Create().
		SetProductID(variant.ProductID).
		SetSku(variant.SKU).
		SetAttributes(variant.Attributes).
		SetStockQuantity(variant.StockQuantity).
		SetPriceAdjustment(variant.PriceAdjustment).
		SetIsActive(variant.IsActive).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create product variant: %w", err)
	}

	variant.ID = created.ID
	variant.CreatedAt = created.CreatedAt
	variant.UpdatedAt = created.UpdatedAt

	return nil
}

// FindByID finds a variant by ID
func (r *ProductVariantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error) {
	v, err := r.client.ProductVariant.
		Query().
		Where(productvariant.ID(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("product variant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find product variant: %w", err)
	}

	return toProductVariantEntity(v), nil
}

// FindBySKU finds a variant by SKU
func (r *ProductVariantRepository) FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error) {
	v, err := r.client.ProductVariant.
		Query().
		Where(productvariant.Sku(sku)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("product variant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find product variant: %w", err)
	}

	return toProductVariantEntity(v), nil
}

// FindByProductID finds all variants for a product
func (r *ProductVariantRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error) {
	variants, err := r.client.ProductVariant.
		Query().
		Where(productvariant.ProductID(productID)).
		Order(ent.Asc(productvariant.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find product variants: %w", err)
	}

	result := make([]*entities.ProductVariant, len(variants))
	for i, v := range variants {
		result[i] = toProductVariantEntity(v)
	}

	return result, nil
}

// Update updates an existing variant
func (r *ProductVariantRepository) Update(ctx context.Context, variant *entities.ProductVariant) error {
	updated, err := r.client.ProductVariant.
		UpdateOneID(variant.ID).
		SetSku(variant.SKU).
		SetAttributes(variant.Attributes).
		SetStockQuantity(variant.StockQuantity).
		SetPriceAdjustment(variant.PriceAdjustment).
		SetIsActive(variant.IsActive).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product variant: %w", err)
	}

	variant.UpdatedAt = updated.UpdatedAt

	return nil
}

// UpdateStock updates stock quantity for a variant
func (r *ProductVariantRepository) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	_, err := r.client.ProductVariant.
		UpdateOneID(id).
		SetStockQuantity(quantity).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// Delete deletes a variant
func (r *ProductVariantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.ProductVariant.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	return nil
}

// toProductVariantEntity converts Ent model to domain entity
func toProductVariantEntity(v *ent.ProductVariant) *entities.ProductVariant {
	return &entities.ProductVariant{
		ID:              v.ID,
		ProductID:       v.ProductID,
		SKU:             v.Sku,
		Attributes:      v.Attributes,
		StockQuantity:   v.StockQuantity,
		PriceAdjustment: v.PriceAdjustment,
		IsActive:        v.IsActive,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/product_variant_repository.go
git commit -m "feat: implement ProductVariant repository with Ent"
```

---

## Task 12: Implement Cart Repository (Ent Adapter) - Part 1

**Files:**
- Create: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/cart_repository.go`

**Step 1: Create cart repository implementation (basic methods)**

```go
package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/go-yippi/internal/adapters/persistence/db/ent"
	"github.com/yourusername/go-yippi/internal/adapters/persistence/db/ent/cart"
	"github.com/yourusername/go-yippi/internal/adapters/persistence/db/ent/cartitem"
	"github.com/yourusername/go-yippi/internal/domain/entities"
)

// CartRepository implements ports.CartRepository
type CartRepository struct {
	client *ent.Client
}

// NewCartRepository creates a new cart repository
func NewCartRepository(client *ent.Client) *CartRepository {
	return &CartRepository{client: client}
}

// Create creates a new cart for a user
func (r *CartRepository) Create(ctx context.Context, userID uuid.UUID) (*entities.Cart, error) {
	created, err := r.client.Cart.
		Create().
		SetUserID(userID).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}

	return toCartEntity(created), nil
}

// FindByUserID finds cart by user ID
func (r *CartRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Cart, error) {
	c, err := r.client.Cart.
		Query().
		Where(cart.UserID(userID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("cart not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find cart: %w", err)
	}

	return toCartEntity(c), nil
}

// GetOrCreate gets existing cart or creates new one
func (r *CartRepository) GetOrCreate(ctx context.Context, userID uuid.UUID) (*entities.Cart, error) {
	c, err := r.FindByUserID(ctx, userID)
	if err == nil {
		return c, nil
	}

	// Cart not found, create new one
	return r.Create(ctx, userID)
}

// AddItem adds an item to cart
func (r *CartRepository) AddItem(ctx context.Context, item *entities.CartItem) error {
	created, err := r.client.CartItem.
		Create().
		SetCartID(item.CartID).
		SetProductVariantID(item.ProductVariantID).
		SetQuantity(item.Quantity).
		SetPriceSnapshot(item.PriceSnapshot).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to add cart item: %w", err)
	}

	item.ID = created.ID
	item.CreatedAt = created.CreatedAt
	item.UpdatedAt = created.UpdatedAt

	return nil
}

// UpdateItemQuantity updates cart item quantity
func (r *CartRepository) UpdateItemQuantity(ctx context.Context, itemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		// Delete item if quantity is 0 or negative
		return r.RemoveItem(ctx, itemID)
	}

	_, err := r.client.CartItem.
		UpdateOneID(itemID).
		SetQuantity(quantity).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	return nil
}

// RemoveItem removes an item from cart
func (r *CartRepository) RemoveItem(ctx context.Context, itemID uuid.UUID) error {
	err := r.client.CartItem.
		DeleteOneID(itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	return nil
}

// ClearCart removes all items from cart
func (r *CartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.client.CartItem.
		Delete().
		Where(cartitem.CartID(cartID)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// Helper functions
func toCartEntity(c *ent.Cart) *entities.Cart {
	return &entities.Cart{
		ID:        c.ID,
		UserID:    c.UserID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toCartItemEntity(ci *ent.CartItem) *entities.CartItem {
	return &entities.CartItem{
		ID:               ci.ID,
		CartID:           ci.CartID,
		ProductVariantID: ci.ProductVariantID,
		Quantity:         ci.Quantity,
		PriceSnapshot:    ci.PriceSnapshot,
		CreatedAt:        ci.CreatedAt,
		UpdatedAt:        ci.UpdatedAt,
	}
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/cart_repository.go
git commit -m "feat: implement Cart repository basic methods"
```

---

## Task 13: Implement Cart Repository (Ent Adapter) - Part 2

**Files:**
- Modify: `/Volumes/External/Work/personal/go-yippi/internal/adapters/persistence/cart_repository.go`

**Step 1: Add query and validation methods**

Append to existing cart_repository.go:

```go
// GetItems gets all items in a cart
func (r *CartRepository) GetItems(ctx context.Context, cartID uuid.UUID) ([]*entities.CartItem, error) {
	items, err := r.client.CartItem.
		Query().
		Where(cartitem.CartID(cartID)).
		Order(ent.Asc(cartitem.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	result := make([]*entities.CartItem, len(items))
	for i, item := range items {
		result[i] = toCartItemEntity(item)
	}

	return result, nil
}

// GetCartWithItems gets cart with all items and computed totals
func (r *CartRepository) GetCartWithItems(ctx context.Context, userID uuid.UUID) (*entities.CartDetail, error) {
	// Get or create cart
	cart, err := r.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get items
	items, err := r.GetItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}

	// Convert to []CartItem (not []*CartItem)
	itemsValue := make([]entities.CartItem, len(items))
	for i, item := range items {
		itemsValue[i] = *item
	}

	// Build cart detail
	detail := &entities.CartDetail{
		Cart:  cart,
		Items: itemsValue,
	}
	detail.CalculateTotals()

	return detail, nil
}

// ItemExists checks if item exists in cart
func (r *CartRepository) ItemExists(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (bool, error) {
	exists, err := r.client.CartItem.
		Query().
		Where(
			cartitem.CartID(cartID),
			cartitem.ProductVariantID(variantID),
		).
		Exist(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check cart item existence: %w", err)
	}

	return exists, nil
}

// GetCartItemByVariant gets cart item by variant ID
func (r *CartRepository) GetCartItemByVariant(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID) (*entities.CartItem, error) {
	item, err := r.client.CartItem.
		Query().
		Where(
			cartitem.CartID(cartID),
			cartitem.ProductVariantID(variantID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("cart item not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	return toCartItemEntity(item), nil
}
```

**Step 2: Commit changes**

```bash
cd /Volumes/External/Work/personal/go-yippi
git add internal/adapters/persistence/cart_repository.go
git commit -m "feat: add cart query and validation methods"
```

---

## Summary

This implementation plan covers:
- ✅ Database schema updates (Product, ProductVariant, Cart, CartItem, User)
- ✅ Ent code generation
- ✅ Domain entities
- ✅ Repository interfaces
- ✅ Repository implementations

**Next steps after completing this plan:**
1. Implement services (StockValidator, PriceCalculator, CartService)
2. Implement HTTP handlers with Huma
3. Add PostgreSQL full-text search migration
4. Write tests
5. Integration testing

**Estimated time:** 4-6 hours for this plan

---

**Plan Status:** ✅ Ready for Execution
**Created:** 2025-12-31
**Target Completion:** Day 1-2 of Phase 2
