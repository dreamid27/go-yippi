package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
)

// ProductService defines the interface for product business logic operations
type ProductService interface {
	CreateProduct(ctx context.Context, product *entities.Product) error
	GetProduct(ctx context.Context, id int) (*entities.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*entities.Product, error)
	ListProducts(ctx context.Context) ([]*entities.Product, error)
	ListPublishedProducts(ctx context.Context) ([]*entities.Product, error)
	ListProductsByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error)
	UpdateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id int) error
	PublishProduct(ctx context.Context, id int) error
	ArchiveProduct(ctx context.Context, id int) error
	QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error)
}
