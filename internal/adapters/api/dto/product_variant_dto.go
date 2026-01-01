package dto

import "time"

// CreateProductVariantRequest defines the request for creating a product variant
type CreateProductVariantRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
	Body      struct {
		SKU             string            `json:"sku" minLength:"3" maxLength:"100" doc:"Stock Keeping Unit (must be unique)"`
		Attributes      map[string]string `json:"attributes" doc:"Variant attributes (e.g., {\"size\": \"M\", \"color\": \"Red\"})"`
		StockQuantity   int               `json:"stock_quantity" minimum:"0" doc:"Available stock quantity"`
		PriceAdjustment float64           `json:"price_adjustment" doc:"Price adjustment from base price (can be negative)"`
		IsActive        *bool             `json:"is_active,omitempty" doc:"Whether variant is active (optional, defaults to true)"`
	}
}

// UpdateProductVariantRequest defines the request for updating a product variant
type UpdateProductVariantRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
	VariantID string `path:"variant_id" doc:"Variant ID (UUID)"`
	Body      struct {
		SKU             *string            `json:"sku,omitempty" minLength:"3" maxLength:"100" doc:"Stock Keeping Unit (optional)"`
		Attributes      *map[string]string `json:"attributes,omitempty" doc:"Variant attributes (optional)"`
		StockQuantity   *int               `json:"stock_quantity,omitempty" minimum:"0" doc:"Available stock quantity (optional)"`
		PriceAdjustment *float64           `json:"price_adjustment,omitempty" doc:"Price adjustment from base price (optional)"`
		IsActive        *bool              `json:"is_active,omitempty" doc:"Whether variant is active (optional)"`
	}
}

// UpdateStockRequest defines the request for updating variant stock
type UpdateStockRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
	VariantID string `path:"variant_id" doc:"Variant ID (UUID)"`
	Body      struct {
		StockQuantity int `json:"stock_quantity" minimum:"0" doc:"New stock quantity"`
	}
}

// GetVariantsRequest defines the request for getting variants by product
type GetVariantsRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
}

// GetVariantRequest defines the request for getting a single variant
type GetVariantRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
	VariantID string `path:"variant_id" doc:"Variant ID (UUID)"`
}

// DeleteVariantRequest defines the request for deleting a variant
type DeleteVariantRequest struct {
	ProductID string `path:"product_id" doc:"Product ID (UUID)"`
	VariantID string `path:"variant_id" doc:"Variant ID (UUID)"`
}

// VariantDetailResponse defines the response for single variant operations
type VariantDetailResponse struct {
	Body ProductVariantDetail
}

// ProductVariantDetail represents a detailed product variant (extends ProductVariantResponse with timestamps)
type ProductVariantDetail struct {
	ID              string            `json:"id" doc:"Variant ID (UUID)"`
	ProductID       string            `json:"product_id" doc:"Product ID (UUID)"`
	SKU             string            `json:"sku" doc:"Stock Keeping Unit"`
	Attributes      map[string]string `json:"attributes" doc:"Variant attributes (e.g., size, color)"`
	StockQuantity   int               `json:"stock_quantity" doc:"Available stock quantity"`
	PriceAdjustment float64           `json:"price_adjustment" doc:"Price adjustment from base price"`
	FinalPrice      float64           `json:"final_price" doc:"Final price (base_price + price_adjustment)"`
	IsActive        bool              `json:"is_active" doc:"Whether variant is active"`
	IsInStock       bool              `json:"is_in_stock" doc:"Whether variant has stock available"`
	CreatedAt       time.Time         `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt       time.Time         `json:"updated_at" doc:"Last update timestamp"`
}

// ProductVariantListResponse defines the response for listing variants
type ProductVariantListResponse struct {
	Body struct {
		Data []ProductVariantDetail `json:"data" doc:"List of product variants"`
	}
}
