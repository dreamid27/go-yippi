package dto

import "time"

// CreateProductRequest defines the request body for creating a product
type CreateProductRequest struct {
	Body struct {
		SKU         string  `json:"sku" minLength:"1" doc:"Stock Keeping Unit (must be unique)"`
		Slug        string  `json:"slug" minLength:"1" doc:"URL-friendly identifier (must be unique)"`
		Name        string  `json:"name" minLength:"1" doc:"Product name"`
		Price       float64 `json:"price" minimum:"0.01" doc:"Product price"`
		Description string  `json:"description" doc:"Product description"`
		Weight      int     `json:"weight" minimum:"0" doc:"Weight in grams for courier calculation"`
		Length      int     `json:"length" minimum:"0" doc:"Length in cm"`
		Width       int     `json:"width" minimum:"0" doc:"Width in cm"`
		Height      int     `json:"height" minimum:"0" doc:"Height in cm"`
		Status      string  `json:"status" enum:"draft,published,archived" default:"draft" doc:"Product status"`
	}
}

// ProductResponse defines the response for product operations
type ProductResponse struct {
	Body struct {
		ID          int       `json:"id"`
		SKU         string    `json:"sku"`
		Slug        string    `json:"slug"`
		Name        string    `json:"name"`
		Price       float64   `json:"price"`
		Description string    `json:"description"`
		Weight      int       `json:"weight"`
		Length      int       `json:"length"`
		Width       int       `json:"width"`
		Height      int       `json:"height"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
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

// ListProductsResponse defines the response for listing products
type ListProductsResponse struct {
	Body struct {
		Products []struct {
			ID          int       `json:"id"`
			SKU         string    `json:"sku"`
			Slug        string    `json:"slug"`
			Name        string    `json:"name"`
			Price       float64   `json:"price"`
			Description string    `json:"description"`
			Weight      int       `json:"weight"`
			Length      int       `json:"length"`
			Width       int       `json:"width"`
			Height      int       `json:"height"`
			Status      string    `json:"status"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"products"`
	}
}

// ListProductsByStatusRequest defines the request for listing products by status
type ListProductsByStatusRequest struct {
	Status string `path:"status" enum:"draft,published,archived" doc:"Product status"`
}

// UpdateProductRequest defines the request for updating a product
type UpdateProductRequest struct {
	ID   int `path:"id" doc:"Product ID"`
	Body struct {
		SKU         string  `json:"sku" minLength:"1" doc:"Stock Keeping Unit (must be unique)"`
		Slug        string  `json:"slug" minLength:"1" doc:"URL-friendly identifier (must be unique)"`
		Name        string  `json:"name" minLength:"1" doc:"Product name"`
		Price       float64 `json:"price" minimum:"0.01" doc:"Product price"`
		Description string  `json:"description" doc:"Product description"`
		Weight      int     `json:"weight" minimum:"0" doc:"Weight in grams for courier calculation"`
		Length      int     `json:"length" minimum:"0" doc:"Length in cm"`
		Width       int     `json:"width" minimum:"0" doc:"Width in cm"`
		Height      int     `json:"height" minimum:"0" doc:"Height in cm"`
		Status      string  `json:"status" enum:"draft,published,archived" doc:"Product status"`
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
