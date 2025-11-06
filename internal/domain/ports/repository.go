package ports

import (
	"context"
	"io"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
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

// BrandRepository defines the interface for brand data operations
type BrandRepository interface {
	Create(ctx context.Context, brand *entities.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Brand, error)
	GetByName(ctx context.Context, name string) (*entities.Brand, error)
	List(ctx context.Context) ([]*entities.Brand, error)
	Update(ctx context.Context, brand *entities.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// StorageRepository defines the interface for file storage operations
type StorageRepository interface {
	// Store uploads a file to storage and returns metadata
	Store(ctx context.Context, bucket, fileName string, reader io.Reader, size int64, contentType string) (*entities.FileMetadata, error)

	// Remove deletes a file from storage
	Remove(ctx context.Context, bucket, fileName string) error

	// GetURL generates a public URL for accessing the file
	GetURL(ctx context.Context, bucket, fileName string) (string, error)

	// GetFile retrieves a file from storage and returns its content
	GetFile(ctx context.Context, bucket, fileName string) (io.ReadCloser, int64, string, error)

	// EnsureBucket creates a bucket if it doesn't exist
	EnsureBucket(ctx context.Context, bucket string) error
}
