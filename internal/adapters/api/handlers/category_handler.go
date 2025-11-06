package handlers

import (
	"context"
	"errors"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/danielgtaylor/huma/v2"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	service ports.CategoryService
}

func NewCategoryHandler(service ports.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// RegisterRoutes registers all category routes with Huma
func (h *CategoryHandler) RegisterRoutes(api huma.API) {
	// Create category
	huma.Register(api, huma.Operation{
		OperationID: "create-category",
		Method:      http.MethodPost,
		Path:        "/categories",
		Summary:     "Create a new category",
		Description: "Creates a new category with a unique name and optional parent",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError},
	}, h.CreateCategory)

	// List all categories
	huma.Register(api, huma.Operation{
		OperationID: "list-categories",
		Method:      http.MethodGet,
		Path:        "/categories",
		Summary:     "List all categories",
		Description: "Retrieves a list of all categories",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.ListCategories)

	// Get category by ID
	huma.Register(api, huma.Operation{
		OperationID: "get-category",
		Method:      http.MethodGet,
		Path:        "/categories/{id}",
		Summary:     "Get a category by ID",
		Description: "Retrieves a category by its ID",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetCategory)

	// Get category by name
	huma.Register(api, huma.Operation{
		OperationID: "get-category-by-name",
		Method:      http.MethodGet,
		Path:        "/categories/name/{name}",
		Summary:     "Get a category by name",
		Description: "Retrieves a category by its name",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetCategoryByName)

	// List categories by parent
	huma.Register(api, huma.Operation{
		OperationID: "list-categories-by-parent",
		Method:      http.MethodGet,
		Path:        "/categories/by-parent",
		Summary:     "List categories by parent",
		Description: "Retrieves categories filtered by parent ID (omit parent_id for root categories)",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.ListCategoriesByParent)

	// Update category
	huma.Register(api, huma.Operation{
		OperationID: "update-category",
		Method:      http.MethodPut,
		Path:        "/categories/{id}",
		Summary:     "Update a category",
		Description: "Updates an existing category's information",
		Tags:        []string{"Categories"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.UpdateCategory)

	// Delete category
	huma.Register(api, huma.Operation{
		OperationID:   "delete-category",
		Method:        http.MethodDelete,
		Path:          "/categories/{id}",
		Summary:       "Delete a category",
		Description:   "Permanently deletes a category from the system (only if it has no children)",
		Tags:          []string{"Categories"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
	}, h.DeleteCategory)
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, input *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	category := &entities.Category{
		Name: input.Body.Name,
	}

	// Handle optional parent ID
	if input.Body.ParentID != nil {
		category.ParentID = input.Body.ParentID
	}

	err := h.service.CreateCategory(ctx, category)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Category with this name already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to create category", err)
	}

	return h.mapToResponse(category), nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, input *dto.GetCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := h.service.GetCategory(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Category not found")
		}
		return nil, huma.Error500InternalServerError("Failed to retrieve category", err)
	}

	return h.mapToResponse(category), nil
}

func (h *CategoryHandler) GetCategoryByName(ctx context.Context, input *dto.GetCategoryByNameRequest) (*dto.CategoryResponse, error) {
	category, err := h.service.GetCategoryByName(ctx, input.Name)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Category not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		return nil, huma.Error500InternalServerError("Failed to retrieve category", err)
	}

	return h.mapToResponse(category), nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, input *struct{}) (*dto.ListCategoriesResponse, error) {
	categories, err := h.service.ListCategories(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list categories", err)
	}

	response := &dto.ListCategoriesResponse{}
	response.Body.Categories = make([]dto.CategoryListItem, 0, len(categories))
	for _, c := range categories {
		response.Body.Categories = append(response.Body.Categories, h.mapToListItem(c))
	}

	return response, nil
}

func (h *CategoryHandler) ListCategoriesByParent(ctx context.Context, input *dto.ListCategoriesByParentRequest) (*dto.ListCategoriesResponse, error) {
	categories, err := h.service.ListCategoriesByParentID(ctx, input.ParentID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		return nil, huma.Error500InternalServerError("Failed to list categories", err)
	}

	response := &dto.ListCategoriesResponse{}
	response.Body.Categories = make([]dto.CategoryListItem, 0, len(categories))
	for _, c := range categories {
		response.Body.Categories = append(response.Body.Categories, h.mapToListItem(c))
	}

	return response, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, input *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category := &entities.Category{
		ID:   input.ID,
		Name: input.Body.Name,
	}

	// Handle optional parent ID
	if input.Body.ParentID != nil {
		category.ParentID = input.Body.ParentID
	}

	err := h.service.UpdateCategory(ctx, category)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Category not found")
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Category with this name already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to update category", err)
	}

	return h.mapToResponse(category), nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, input *dto.DeleteCategoryRequest) (*struct{}, error) {
	err := h.service.DeleteCategory(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Category not found")
		}
		return nil, huma.Error500InternalServerError("Failed to delete category", err)
	}

	return nil, nil
}

// mapToResponse maps domain entity to response DTO
func (h *CategoryHandler) mapToResponse(category *entities.Category) *dto.CategoryResponse {
	response := &dto.CategoryResponse{}
	response.Body.ID = category.ID
	response.Body.Name = category.Name
	response.Body.ParentID = category.ParentID
	response.Body.CreatedAt = category.CreatedAt
	response.Body.UpdatedAt = category.UpdatedAt
	return response
}

// mapToListItem maps domain entity to list item DTO
func (h *CategoryHandler) mapToListItem(category *entities.Category) dto.CategoryListItem {
	return dto.CategoryListItem{
		ID:        category.ID,
		Name:      category.Name,
		ParentID:  category.ParentID,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
