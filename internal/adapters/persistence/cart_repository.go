package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/cart"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/cartitem"
	"example.com/go-yippi/internal/domain/entities"
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
