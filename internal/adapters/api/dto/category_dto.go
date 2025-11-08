package dto

import "time"

// CreateCategoryRequest defines the request body for creating a category
type CreateCategoryRequest struct {
	Body struct {
		Name     string  `json:"name" minLength:"1" doc:"Category name (must be unique)"`
		ParentID *string `json:"parent_id,omitempty" doc:"Parent category ID (optional UUID, for subcategories)"`
	}
}

// CategoryResponse defines the response for category operations
type CategoryResponse struct {
	Body struct {
		ID        string    `json:"id" doc:"Category ID (UUID)"`
		Name      string    `json:"name" doc:"Category name"`
		ParentID  *string   `json:"parent_id,omitempty" doc:"Parent category ID (UUID)"`
		CreatedAt time.Time `json:"created_at" doc:"Creation timestamp"`
		UpdatedAt time.Time `json:"updated_at" doc:"Last update timestamp"`
	}
}

// GetCategoryRequest defines the request for getting a single category
type GetCategoryRequest struct {
	ID string `path:"id" doc:"Category ID (UUID)"`
}

// GetCategoryByNameRequest defines the request for getting a category by name
type GetCategoryByNameRequest struct {
	Name string `path:"name" doc:"Category name"`
}

// CategoryListItem represents a category in a list response
type CategoryListItem struct {
	ID        string    `json:"id" doc:"Category ID (UUID)"`
	Name      string    `json:"name" doc:"Category name"`
	ParentID  *string   `json:"parent_id,omitempty" doc:"Parent category ID (UUID)"`
	CreatedAt time.Time `json:"created_at" doc:"Creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" doc:"Last update timestamp"`
}

// ListCategoriesResponse defines the response for listing all categories
type ListCategoriesResponse struct {
	Body struct {
		Categories []CategoryListItem `json:"categories" doc:"List of categories"`
	}
}

// ListCategoriesByParentRequest defines the request for listing categories by parent
type ListCategoriesByParentRequest struct {
	ParentID string `query:"parent_id" default:"" doc:"Parent category ID (UUID for specific parent, empty for root categories)"`
}

// UpdateCategoryRequest defines the request for updating a category
type UpdateCategoryRequest struct {
	ID   string `path:"id" doc:"Category ID (UUID)"`
	Body struct {
		Name     string  `json:"name" minLength:"1" doc:"Category name (must be unique)"`
		ParentID *string `json:"parent_id,omitempty" doc:"Parent category ID (optional UUID)"`
	}
}

// DeleteCategoryRequest defines the request for deleting a category
type DeleteCategoryRequest struct {
	ID string `path:"id" doc:"Category ID (UUID)"`
}
