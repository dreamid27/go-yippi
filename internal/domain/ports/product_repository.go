package ports

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/domain/entities"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query performs a flexible query with filters, sorting, and pagination
	Query(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error)

	// Legacy methods (can be deprecated in favor of Query)
	List(ctx context.Context) ([]*entities.Product, error)
	ListByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error)
}
