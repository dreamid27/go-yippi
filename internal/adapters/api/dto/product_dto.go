package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProductRequest defines the request body for creating a product
type CreateProductRequest struct {
	Body struct {
		SKU         string     `json:"sku" minLength:"1" doc:"Stock Keeping Unit (must be unique)"`
		Slug        *string    `json:"slug,omitempty" minLength:"1" doc:"URL-friendly identifier (optional, auto-generated from name if not provided)"`
		Name        string     `json:"name" minLength:"1" doc:"Product name"`
		Price       float64    `json:"price" minimum:"0.01" doc:"Product price"`
		Description string     `json:"description" doc:"Product description"`
		Weight      *int       `json:"weight,omitempty" minimum:"0" doc:"Weight in grams for courier calculation (optional)"`
		Length      *int       `json:"length,omitempty" minimum:"0" doc:"Length in cm (optional)"`
		Width       *int       `json:"width,omitempty" minimum:"0" doc:"Width in cm (optional)"`
		Height      *int       `json:"height,omitempty" minimum:"0" doc:"Height in cm (optional)"`
		ImageURLs   []string   `json:"image_urls,omitempty" doc:"Access links to product images (optional)"`
		Status      *string    `json:"status,omitempty" enum:"draft,published,archived" doc:"Product status (optional, defaults to draft)"`
		CategoryID  *int       `json:"category_id,omitempty" doc:"Category ID (optional)"`
		BrandID     *uuid.UUID `json:"brand_id,omitempty" doc:"Brand ID (optional)"`
	}
}

// ProductResponse defines the response for product operations
type ProductResponse struct {
	Body struct {
		ID          int        `json:"id"`
		SKU         string     `json:"sku"`
		Slug        string     `json:"slug"`
		Name        string     `json:"name"`
		Price       float64    `json:"price"`
		Description string     `json:"description"`
		Weight      int        `json:"weight"`
		Length      int        `json:"length"`
		Width       int        `json:"width"`
		Height      int        `json:"height"`
		ImageURLs   []string   `json:"image_urls"`
		Status      string     `json:"status"`
		CategoryID  *int       `json:"category_id,omitempty"`
		BrandID     *uuid.UUID `json:"brand_id,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}
}

// GetProductRequest defines the request for getting a single product
type GetProductRequest struct {
	ID int `path:"id" doc:"Product ID"`
}

// GetProductBySKURequest defines the request for getting a product by SKU
type GetProductBySKURequest struct {
	SKU string `path:"sku" doc:"Product SKU"`
}

// GetProductBySlugRequest defines the request for getting a product by slug
type GetProductBySlugRequest struct {
	Slug string `path:"slug" doc:"Product slug"`
}

// ProductListItem represents a product in a list response
type ProductListItem struct {
	ID          int        `json:"id" doc:"Product ID"`
	SKU         string     `json:"sku" doc:"Stock Keeping Unit"`
	Slug        string     `json:"slug" doc:"URL-friendly identifier"`
	Name        string     `json:"name" doc:"Product name"`
	Price       float64    `json:"price" doc:"Product price"`
	Description string     `json:"description" doc:"Product description"`
	Weight      int        `json:"weight" doc:"Weight in grams"`
	Length      int        `json:"length" doc:"Length in cm"`
	Width       int        `json:"width" doc:"Width in cm"`
	Height      int        `json:"height" doc:"Height in cm"`
	ImageURLs   []string   `json:"image_urls" doc:"Access links to product images"`
	Status      string     `json:"status" doc:"Product status"`
	CategoryID  *int       `json:"category_id,omitempty" doc:"Category ID"`
	BrandID     *uuid.UUID `json:"brand_id,omitempty" doc:"Brand ID"`
	CreatedAt   time.Time  `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt   time.Time  `json:"updated_at" doc:"Last update timestamp"`
}

// QueryProductsRequest defines the request for querying products with filters, sorting, and pagination
type QueryProductsRequest struct {
	// Filters - array of filter conditions (parsed by custom Resolver)
	// Usage: ?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published
	// Category filtering: ?filter[0][field]=category_id&filter[0][operator]=in&filter[0][value]=[1,2,5]
	// Note: Category filtering automatically includes all subcategories
	// Note: This is populated by the Resolve() method and hidden from OpenAPI schema
	Filters []FilterDTO

	// Sort - array of sort parameters (parsed by custom Resolver)
	// Usage: ?sort[0][field]=price&sort[0][order]=desc
	// Note: This is populated by the Resolve() method and hidden from OpenAPI schema
	Sort []SortDTO

	// Pagination parameters
	Cursor       string `query:"cursor" doc:"Pagination cursor from previous response"`
	Limit        int    `query:"limit" default:"20" doc:"Items per page (default: 20, max: 100)"`
	Direction    string `query:"direction" default:"forward" doc:"Pagination direction: forward or backward (default: forward)"`
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
	ID   int `path:"id" doc:"Product ID"`
	Body struct {
		SKU         string     `json:"sku" minLength:"1" doc:"Stock Keeping Unit (must be unique)"`
		Slug        *string    `json:"slug,omitempty" minLength:"1" doc:"URL-friendly identifier (optional, auto-generated from name if not provided)"`
		Name        string     `json:"name" minLength:"1" doc:"Product name"`
		Price       float64    `json:"price" minimum:"0.01" doc:"Product price"`
		Description string     `json:"description" doc:"Product description"`
		Weight      *int       `json:"weight,omitempty" minimum:"0" doc:"Weight in grams for courier calculation (optional)"`
		Length      *int       `json:"length,omitempty" minimum:"0" doc:"Length in cm (optional)"`
		Width       *int       `json:"width,omitempty" minimum:"0" doc:"Width in cm (optional)"`
		Height      *int       `json:"height,omitempty" minimum:"0" doc:"Height in cm (optional)"`
		ImageURLs   []string   `json:"image_urls,omitempty" doc:"Access links to product images (optional)"`
		Status      *string    `json:"status,omitempty" enum:"draft,published,archived" doc:"Product status (optional, defaults to draft)"`
		CategoryID  *int       `json:"category_id,omitempty" doc:"Category ID (optional)"`
		BrandID     *uuid.UUID `json:"brand_id,omitempty" doc:"Brand ID (optional)"`
	}
}

// DeleteProductRequest defines the request for deleting a product
type DeleteProductRequest struct {
	ID int `path:"id" doc:"Product ID"`
}

// PublishProductRequest defines the request for publishing a product
type PublishProductRequest struct {
	ID int `path:"id" doc:"Product ID"`
}

// ArchiveProductRequest defines the request for archiving a product
type ArchiveProductRequest struct {
	ID int `path:"id" doc:"Product ID"`
}
