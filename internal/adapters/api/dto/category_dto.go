package dto

import "time"

// CreateCategoryRequest defines the request body for creating a category
type CreateCategoryRequest struct {
	Body struct {
		Name     string `json:"name" minLength:"1" doc:"Category name (must be unique)"`
		ParentID *int   `json:"parent_id,omitempty" doc:"Parent category ID (optional, for subcategories)"`
	}
}

// CategoryResponse defines the response for category operations
type CategoryResponse struct {
	Body struct {
		ID        int       `json:"id" doc:"Category ID"`
		Name      string    `json:"name" doc:"Category name"`
		ParentID  *int      `json:"parent_id,omitempty" doc:"Parent category ID"`
		CreatedAt time.Time `json:"created_at" doc:"Creation timestamp"`
		UpdatedAt time.Time `json:"updated_at" doc:"Last update timestamp"`
	}
}

// GetCategoryRequest defines the request for getting a single category
type GetCategoryRequest struct {
	ID int `path:"id" doc:"Category ID"`
}

// GetCategoryByNameRequest defines the request for getting a category by name
type GetCategoryByNameRequest struct {
	Name string `path:"name" doc:"Category name"`
}

// CategoryListItem represents a category in a list response
type CategoryListItem struct {
	ID        int       `json:"id" doc:"Category ID"`
	Name      string    `json:"name" doc:"Category name"`
	ParentID  *int      `json:"parent_id,omitempty" doc:"Parent category ID"`
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
	ParentID *int `query:"parent_id" doc:"Parent category ID (omit or null for root categories)"`
}

// UpdateCategoryRequest defines the request for updating a category
type UpdateCategoryRequest struct {
	ID   int `path:"id" doc:"Category ID"`
	Body struct {
		Name     string `json:"name" minLength:"1" doc:"Category name (must be unique)"`
		ParentID *int   `json:"parent_id,omitempty" doc:"Parent category ID (optional)"`
	}
}

// DeleteCategoryRequest defines the request for deleting a category
type DeleteCategoryRequest struct {
	ID int `path:"id" doc:"Category ID"`
}
