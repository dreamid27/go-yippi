package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// ProductService defines the interface for product business logic operations
type ProductService interface {
	CreateProduct(ctx context.Context, product *entities.Product) error
	GetProduct(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*entities.Product, error)
	ListProducts(ctx context.Context) ([]*entities.Product, error)
	ListPublishedProducts(ctx context.Context) ([]*entities.Product, error)
	ListProductsByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error)
	UpdateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	PublishProduct(ctx context.Context, id uuid.UUID) error
	ArchiveProduct(ctx context.Context, id uuid.UUID) error
	QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error)

	// Phase 2: Search & Variants
	SearchProducts(ctx context.Context, params SearchProductsParams) (*ProductSearchResult, error)
	GetProductWithVariants(ctx context.Context, id uuid.UUID) (*ProductDetail, error)
}

// SearchProductsParams contains parameters for product search (Phase 2)
type SearchProductsParams struct {
	Search     string      // Full-text search query
	CategoryID *uuid.UUID  // Filter by category
	BrandID    *uuid.UUID  // Filter by brand
	MinPrice   *float64    // Min base_price
	MaxPrice   *float64    // Max base_price
	Size       *string     // Filter by variant attribute
	Color      *string     // Filter by variant attribute
	Status     *string     // published, draft, archived
	SortBy     string      // name, price, created_at, relevance
	SortOrder  string      // asc, desc
	Page       int         // Page number (1-indexed)
	Limit      int         // Items per page (max 100)
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CategorySummary contains category summary information
type CategorySummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// BrandSummary contains brand summary information
type BrandSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// ProductListItem contains product information for list view
type ProductListItem struct {
	ID            uuid.UUID        `json:"id"`
	Slug          string           `json:"slug"`
	Name          string           `json:"name"`
	BasePrice     float64          `json:"base_price"`
	Description   string           `json:"description"`
	ImageURLs     []string         `json:"image_urls"`
	Status        string           `json:"status"`
	Category      *CategorySummary `json:"category,omitempty"`
	Brand         *BrandSummary    `json:"brand,omitempty"`
	VariantsCount int              `json:"variants_count"`
	MinPrice      float64          `json:"min_price"`
	MaxPrice      float64          `json:"max_price"`
	HasStock      bool             `json:"has_stock"`
}

// ProductSearchResult contains search results with pagination
type ProductSearchResult struct {
	Products   []*ProductListItem `json:"products"`
	Pagination PaginationInfo     `json:"pagination"`
}

// ProductVariantDetail contains variant information with calculated fields
type ProductVariantDetail struct {
	ID              uuid.UUID         `json:"id"`
	SKU             string            `json:"sku"`
	Attributes      map[string]string `json:"attributes"`
	StockQuantity   int               `json:"stock_quantity"`
	PriceAdjustment float64           `json:"price_adjustment"`
	FinalPrice      float64           `json:"final_price"`
	IsActive        bool              `json:"is_active"`
	IsInStock       bool              `json:"is_in_stock"`
}

// ProductDetail contains complete product information with variants
type ProductDetail struct {
	Product  *entities.Product       `json:"product"`
	Category *entities.Category      `json:"category,omitempty"`
	Brand    *entities.Brand         `json:"brand,omitempty"`
	Variants []*ProductVariantDetail `json:"variants"`
}
