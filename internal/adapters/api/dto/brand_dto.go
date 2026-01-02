package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateBrandRequest defines the request body for creating a brand
type CreateBrandRequest struct {
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"255" doc:"Brand name (must be unique)"`
	}
}

// UpdateBrandRequest defines the request body for updating a brand
type UpdateBrandRequest struct {
	ID uuid.UUID `path:"id" doc:"Brand ID"`
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"255" doc:"Brand name (must be unique)"`
	}
}

// BrandResponse defines the response for brand operations
type BrandResponse struct {
	Body struct {
		ID          uuid.UUID `json:"id" doc:"Brand unique identifier"`
		Name        string    `json:"name" doc:"Brand name"`
		Slug        string    `json:"slug" doc:"URL-friendly slug"`
		Description string    `json:"description,omitempty" doc:"Brand description"`
		CreatedAt   time.Time `json:"created_at" doc:"Creation timestamp"`
		UpdatedAt   time.Time `json:"updated_at" doc:"Last update timestamp"`
	}
}

// GetBrandRequest defines the request for getting a single brand by ID
type GetBrandRequest struct {
	ID uuid.UUID `path:"id" doc:"Brand ID"`
}

// GetBrandByNameRequest defines the request for getting a brand by name
type GetBrandByNameRequest struct {
	Name string `path:"name" doc:"Brand name"`
}

// DeleteBrandRequest defines the request for deleting a brand
type DeleteBrandRequest struct {
	ID uuid.UUID `path:"id" doc:"Brand ID"`
}

// BrandListItem represents a brand in a list response
type BrandListItem struct {
	ID        uuid.UUID `json:"id" doc:"Brand unique identifier"`
	Name      string    `json:"name" doc:"Brand name"`
	CreatedAt time.Time `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" doc:"Last update timestamp"`
}

// ListBrandsRequest defines the request for listing brands with optional filters
type ListBrandsRequest struct {
	CategoryIDs string `query:"category_ids" doc:"Comma-separated list of category UUIDs to filter brands by (optional)"`
}

// GetCategoryIDs parses the comma-separated category IDs and returns a slice of UUIDs
func (r *ListBrandsRequest) GetCategoryIDs() []uuid.UUID {
	if r.CategoryIDs == "" {
		return nil
	}

	parts := strings.Split(r.CategoryIDs, ",")
	categoryIDs := make([]uuid.UUID, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := uuid.Parse(part)
		if err == nil {
			categoryIDs = append(categoryIDs, id)
		}
	}

	return categoryIDs
}

// ListBrandsResponse defines the response for listing brands
type ListBrandsResponse struct {
	Body struct {
		Brands []BrandListItem `json:"brands" doc:"List of brands"`
	}
}
