package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// GuestCartItem represents an item in guest's cart (before login)
type GuestCartItem struct {
	ProductVariantID uuid.UUID
	Quantity         int
}

// CartItemDetail represents cart item with full product and variant information
type CartItemDetail struct {
	ID              uuid.UUID
	Quantity        int
	PriceSnapshot   float64
	ItemSubtotal    float64
	ProductVariant  *entities.ProductVariant
	Product         *entities.Product
}

// CartDetail represents cart with all items and computed fields
type CartDetail struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Items     []*CartItemDetail
	Subtotal  float64
	ItemCount int
}

// CartService defines the interface for cart business logic operations
type CartService interface {
	// GetOrCreateCart retrieves or creates a cart for a user
	GetOrCreateCart(ctx context.Context, userID uuid.UUID) (*CartDetail, error)

	// AddItemToCart adds an item to the user's cart
	AddItemToCart(ctx context.Context, userID uuid.UUID, variantID uuid.UUID, quantity int) (*entities.CartItem, error)

	// UpdateCartItemQuantity updates the quantity of a cart item
	UpdateCartItemQuantity(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID, newQuantity int) (*entities.CartItem, error)

	// RemoveCartItem removes an item from the cart
	RemoveCartItem(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID) error

	// ClearCart clears all items from the user's cart
	ClearCart(ctx context.Context, userID uuid.UUID) error

	// MergeGuestCart merges guest cart items with user's cart after login
	MergeGuestCart(ctx context.Context, userID uuid.UUID, guestItems []GuestCartItem) (*CartDetail, error)

	// GetCartWithDetails retrieves the cart with full details (items, products, variants)
	GetCartWithDetails(ctx context.Context, userID uuid.UUID) (*CartDetail, error)
}
