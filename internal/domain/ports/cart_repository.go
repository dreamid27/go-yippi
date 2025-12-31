package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
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
