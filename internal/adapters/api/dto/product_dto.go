package dto

import (
	"time"
)

// CreateProductRequest defines the request body for creating a product
type CreateProductRequest struct {
	Body struct {
		Slug           *string  `json:"slug,omitempty" minLength:"1" doc:"URL-friendly identifier (optional, auto-generated from name if not provided)"`
		Name           string   `json:"name" minLength:"1" doc:"Product name"`
		BasePrice      float64  `json:"base_price" minimum:"0.01" doc:"Base product price"`
		Description    string   `json:"description" doc:"Product description"`
		Weight         *int     `json:"weight,omitempty" minimum:"0" doc:"Weight in grams for courier calculation (optional)"`
		Length         *int     `json:"length,omitempty" minimum:"0" doc:"Length in cm (optional)"`
		Width          *int     `json:"width,omitempty" minimum:"0" doc:"Width in cm (optional)"`
		Height         *int     `json:"height,omitempty" minimum:"0" doc:"Height in cm (optional)"`
		ImageURLs      []string `json:"image_urls,omitempty" doc:"Access links to product images (optional)"`
		Status         *string  `json:"status,omitempty" enum:"draft,published,archived" doc:"Product status (optional, defaults to draft)"`
		CategoryID     *string  `json:"category_id,omitempty" doc:"Category ID (optional UUID)"`
		BrandID        *string  `json:"brand_id,omitempty" doc:"Brand ID (optional UUID)"`
		StockQuantity  *int     `json:"stock_quantity,omitempty" minimum:"0" doc:"Initial stock quantity for auto-created default variant (optional)"`
	}
}

// ProductResponse defines the response for product operations
type ProductResponse struct {
	Body struct {
		ID          string     `json:"id"`
		Slug        string     `json:"slug"`
		Name        string     `json:"name"`
		BasePrice   float64    `json:"base_price"`
		Description string     `json:"description"`
		Weight      int        `json:"weight"`
		Length      int        `json:"length"`
		Width       int        `json:"width"`
		Height      int        `json:"height"`
		ImageURLs   []string   `json:"image_urls"`
		Status      string     `json:"status"`
		CategoryID  *string    `json:"category_id,omitempty"`
		BrandID     *string    `json:"brand_id,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}
}

// GetProductRequest defines the request for getting a single product
type GetProductRequest struct {
	ID string `path:"id" doc:"Product ID (UUID)"`
}

// GetProductBySlugRequest defines the request for getting a product by slug
type GetProductBySlugRequest struct {
	Slug string `path:"slug" doc:"Product slug"`
}

// ProductListItem represents a product in a list response
type ProductListItem struct {
	ID          string     `json:"id" doc:"Product ID (UUID)"`
	Slug        string     `json:"slug" doc:"URL-friendly identifier"`
	Name        string     `json:"name" doc:"Product name"`
	BasePrice   float64    `json:"base_price" doc:"Base product price"`
	Description string     `json:"description" doc:"Product description"`
	Weight      int        `json:"weight" doc:"Weight in grams"`
	Length      int        `json:"length" doc:"Length in cm"`
	Width       int        `json:"width" doc:"Width in cm"`
	Height      int        `json:"height" doc:"Height in cm"`
	ImageURLs   []string   `json:"image_urls" doc:"Access links to product images"`
	Status      string     `json:"status" doc:"Product status"`
	CategoryID  *string    `json:"category_id,omitempty" doc:"Category ID (UUID)"`
	BrandID     *string    `json:"brand_id,omitempty" doc:"Brand ID (UUID)"`
	CreatedAt   time.Time  `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt   time.Time  `json:"updated_at" doc:"Last update timestamp"`
}

// QueryProductsRequest defines the request for querying products with filters, sorting, and pagination
type QueryProductsRequest struct {
	// Filters - array of filter conditions
	// Usage: ?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published
	// Example: ?filter[0][field]=price&filter[0][operator]=gte&filter[0][value]=100
	// Category filtering: ?filter[0][field]=category_id&filter[0][operator]=in&filter[0][value]=[uuid1,uuid2]
	// Note: Category filtering automatically includes all subcategories
	Filters []FilterDTO `query:"filter" doc:"Array of filter conditions. Use filter[i][field], filter[i][operator], filter[i][value] format. Operators: eq, ne, gt, gte, lt, lte, like, ilike, in, not_in, is_null, not_null, starts, ends"`

	// Sort - array of sort parameters
	// Usage: ?sort[0][field]=price&sort[0][order]=desc
	// Example: ?sort[0][field]=created_at&sort[0][order]=asc&sort[1][field]=name&sort[1][order]=asc
	Sort []SortDTO `query:"sort" doc:"Array of sort parameters. Use sort[i][field] and sort[i][order] format. Order values: asc, desc"`

	// Pagination parameters
	Cursor       string `query:"cursor" doc:"Pagination cursor from previous response"`
	Limit        int    `query:"limit" default:"20" doc:"Items per page (default: 20, max: 100)"`
	Direction    string `query:"direction" default:"forward" enum:"forward,backward" doc:"Pagination direction (default: forward)"`
	IncludeTotal bool   `query:"include_total" default:"false" doc:"Include total count in response (default: false, may be expensive)"`
}

// QueryProductsResponse defines the response for querying products with pagination
type QueryProductsResponse struct {
	Body struct {
		Data     []ProductListItem `json:"data" doc:"List of products"`
		PageInfo PageInfoDTO       `json:"page_info" doc:"Pagination information"`
	}
}

// ListProductsResponse defines the response for listing products (legacy, kept for backward compatibility)
type ListProductsResponse struct {
	Body struct {
		Products []ProductListItem `json:"products" doc:"List of products"`
	}
}

// ListProductsByStatusRequest defines the request for listing products by status (legacy)
type ListProductsByStatusRequest struct {
	Status string `path:"status" enum:"draft,published,archived" doc:"Product status"`
}

// UpdateProductRequest defines the request for updating a product
type UpdateProductRequest struct {
	ID   string `path:"id" doc:"Product ID (UUID)"`
	Body struct {
		Slug        *string    `json:"slug,omitempty" minLength:"1" doc:"URL-friendly identifier (optional, auto-generated from name if not provided)"`
		Name        string     `json:"name" minLength:"1" doc:"Product name"`
		BasePrice   float64    `json:"base_price" minimum:"0.01" doc:"Base product price"`
		Description string     `json:"description" doc:"Product description"`
		Weight      *int       `json:"weight,omitempty" minimum:"0" doc:"Weight in grams for courier calculation (optional)"`
		Length      *int       `json:"length,omitempty" minimum:"0" doc:"Length in cm (optional)"`
		Width       *int       `json:"width,omitempty" minimum:"0" doc:"Width in cm (optional)"`
		Height      *int       `json:"height,omitempty" minimum:"0" doc:"Height in cm (optional)"`
		ImageURLs   []string   `json:"image_urls,omitempty" doc:"Access links to product images (optional)"`
		Status      *string    `json:"status,omitempty" enum:"draft,published,archived" doc:"Product status (optional, defaults to draft)"`
		CategoryID  *string    `json:"category_id,omitempty" doc:"Category ID (optional UUID)"`
		BrandID     *string    `json:"brand_id,omitempty" doc:"Brand ID (optional UUID)"`
	}
}

// DeleteProductRequest defines the request for deleting a product
type DeleteProductRequest struct {
	ID string `path:"id" doc:"Product ID (UUID)"`
}

// PublishProductRequest defines the request for publishing a product
type PublishProductRequest struct {
	ID string `path:"id" doc:"Product ID (UUID)"`
}

// ArchiveProductRequest defines the request for archiving a product
type ArchiveProductRequest struct {
	ID string `path:"id" doc:"Product ID (UUID)"`
}

// Phase 2: Search and Variant DTOs

// SearchProductsRequest defines the request for searching products with filters (REQ-9.2)
type SearchProductsRequest struct {
	Search     string  `query:"search" doc:"Full-text search query (optional)"`
	CategoryID string  `query:"category_id" doc:"Filter by category ID (UUID)"`
	BrandID    string  `query:"brand_id" doc:"Filter by brand ID (UUID)"`
	MinPrice   float64 `query:"min_price" minimum:"0" doc:"Minimum base price filter"`
	MaxPrice   float64 `query:"max_price" minimum:"0" doc:"Maximum base price filter"`
	Size       string  `query:"size" doc:"Filter by variant size attribute"`
	Color      string  `query:"color" doc:"Filter by variant color attribute"`
	Status     string  `query:"status" enum:",published,draft,archived" doc:"Filter by product status"`
	SortBy     string  `query:"sort_by" default:"created_at" enum:"name,price,created_at,relevance" doc:"Sort field (default: created_at)"`
	SortOrder  string  `query:"sort_order" default:"desc" enum:"asc,desc" doc:"Sort order (default: desc)"`
	Page       int     `query:"page" default:"1" minimum:"1" doc:"Page number (default: 1)"`
	Limit      int     `query:"limit" default:"20" minimum:"1" maximum:"100" doc:"Items per page (default: 20, max: 100)"`
}

// CategorySummaryDTO represents category summary in product list
type CategorySummaryDTO struct {
	ID   string `json:"id" doc:"Category ID (UUID)"`
	Name string `json:"name" doc:"Category name"`
	Slug string `json:"slug" doc:"Category slug"`
}

// BrandSummaryDTO represents brand summary in product list
type BrandSummaryDTO struct {
	ID   string `json:"id" doc:"Brand ID (UUID)"`
	Name string `json:"name" doc:"Brand name"`
	Slug string `json:"slug" doc:"Brand slug"`
}

// ProductListItemResponse represents a product in search results (REQ-9.2)
type ProductListItemResponse struct {
	ID            string               `json:"id" doc:"Product ID (UUID)"`
	Slug          string               `json:"slug" doc:"URL-friendly identifier"`
	Name          string               `json:"name" doc:"Product name"`
	BasePrice     float64              `json:"base_price" doc:"Base product price"`
	Description   string               `json:"description" doc:"Product description"`
	ImageURLs     []string             `json:"image_urls" doc:"Product images"`
	Status        string               `json:"status" doc:"Product status"`
	Category      *CategorySummaryDTO  `json:"category,omitempty" doc:"Category information"`
	Brand         *BrandSummaryDTO     `json:"brand,omitempty" doc:"Brand information"`
	VariantsCount int                  `json:"variants_count" doc:"Number of variants"`
	MinPrice      float64              `json:"min_price" doc:"Minimum variant price"`
	MaxPrice      float64              `json:"max_price" doc:"Maximum variant price"`
	HasStock      bool                 `json:"has_stock" doc:"Has available stock"`
}

// PaginationResponse contains pagination metadata (REQ-9.2)
type PaginationResponse struct {
	Page       int `json:"page" doc:"Current page number"`
	Limit      int `json:"limit" doc:"Items per page"`
	Total      int `json:"total" doc:"Total number of items"`
	TotalPages int `json:"total_pages" doc:"Total number of pages"`
}

// ProductSearchResponse defines the response for product search (REQ-9.2)
type ProductSearchResponse struct {
	Body struct {
		Data       []ProductListItemResponse `json:"data" doc:"List of products"`
		Pagination PaginationResponse        `json:"pagination" doc:"Pagination information"`
	}
}

// DimensionsResponse represents product dimensions
type DimensionsResponse struct {
	Length int `json:"length" doc:"Length in cm"`
	Width  int `json:"width" doc:"Width in cm"`
	Height int `json:"height" doc:"Height in cm"`
}

// ProductVariantResponse represents a product variant (REQ-9.2)
type ProductVariantResponse struct {
	ID              string            `json:"id" doc:"Variant ID (UUID)"`
	ProductID       string            `json:"product_id" doc:"Product ID (UUID)"`
	SKU             string            `json:"sku" doc:"Stock Keeping Unit"`
	Attributes      map[string]string `json:"attributes" doc:"Variant attributes (size, color, etc.)"`
	StockQuantity   int               `json:"stock_quantity" doc:"Available stock quantity"`
	PriceAdjustment float64           `json:"price_adjustment" doc:"Price adjustment from base price"`
	FinalPrice      float64           `json:"final_price" doc:"Calculated final price (base_price + adjustment)"`
	IsActive        bool              `json:"is_active" doc:"Variant is active"`
	IsInStock       bool              `json:"is_in_stock" doc:"Variant has stock available"`
	CreatedAt       time.Time         `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt       time.Time         `json:"updated_at" doc:"Last update timestamp"`
}

// ProductDetailResponse defines the response for product detail with variants (REQ-9.2)
type ProductDetailResponse struct {
	Body struct {
		ID          string                   `json:"id" doc:"Product ID (UUID)"`
		Slug        string                   `json:"slug" doc:"URL-friendly identifier"`
		Name        string                   `json:"name" doc:"Product name"`
		BasePrice   float64                  `json:"base_price" doc:"Base product price"`
		Description string                   `json:"description" doc:"Product description"`
		ImageURLs   []string                 `json:"image_urls" doc:"Product images"`
		Status      string                   `json:"status" doc:"Product status"`
		Weight      int                      `json:"weight" doc:"Weight in grams"`
		Dimensions  DimensionsResponse       `json:"dimensions" doc:"Product dimensions"`
		Category    *CategoryResponse        `json:"category,omitempty" doc:"Category information"`
		Brand       *BrandResponse           `json:"brand,omitempty" doc:"Brand information"`
		Variants    []ProductVariantResponse `json:"variants" doc:"Product variants"`
		CreatedAt   time.Time                `json:"created_at" doc:"Creation timestamp"`
		UpdatedAt   time.Time                `json:"updated_at" doc:"Last update timestamp"`
	}
}
