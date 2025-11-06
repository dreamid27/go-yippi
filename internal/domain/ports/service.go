package ports

import (
	"context"
	"io"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
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

// CategoryService defines the interface for category business logic operations
type CategoryService interface {
	CreateCategory(ctx context.Context, category *entities.Category) error
	GetCategory(ctx context.Context, id int) (*entities.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*entities.Category, error)
	ListCategories(ctx context.Context) ([]*entities.Category, error)
	ListCategoriesByParentID(ctx context.Context, parentID *int) ([]*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

// BrandService defines the interface for brand business logic operations
type BrandService interface {
	CreateBrand(ctx context.Context, brand *entities.Brand) error
	GetBrand(ctx context.Context, id uuid.UUID) (*entities.Brand, error)
	GetBrandByName(ctx context.Context, name string) (*entities.Brand, error)
	ListBrands(ctx context.Context) ([]*entities.Brand, error)
	UpdateBrand(ctx context.Context, brand *entities.Brand) error
	DeleteBrand(ctx context.Context, id uuid.UUID) error
}

// StorageService defines the interface for file storage operations
type StorageService interface {
	UploadFile(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error)
	DeleteFile(ctx context.Context, bucket, fileName string) error
	GetFileURL(ctx context.Context, bucket, fileName string) (string, error)
	DownloadFile(ctx context.Context, bucket, fileName string) (io.ReadCloser, int64, string, error)
}
