package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int) (*entities.User, error)
	List(ctx context.Context) ([]*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int) error
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id int) (*entities.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id int) error

	// Query performs a flexible query with filters, sorting, and pagination
	Query(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error)

	// Legacy methods (can be deprecated in favor of Query)
	List(ctx context.Context) ([]*entities.Product, error)
	ListByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error)
}
