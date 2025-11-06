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

// BrandHandler handles HTTP requests for brands
type BrandHandler struct {
	service ports.BrandService
}

func NewBrandHandler(service ports.BrandService) *BrandHandler {
	return &BrandHandler{service: service}
}

// RegisterRoutes registers all brand routes with Huma
func (h *BrandHandler) RegisterRoutes(api huma.API) {
	// Create brand
	huma.Register(api, huma.Operation{
		OperationID: "create-brand",
		Method:      http.MethodPost,
		Path:        "/brands",
		Summary:     "Create a new brand",
		Description: "Creates a new brand with a unique name",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError},
	}, h.CreateBrand)

	// List all brands
	huma.Register(api, huma.Operation{
		OperationID: "list-brands",
		Method:      http.MethodGet,
		Path:        "/brands",
		Summary:     "List all brands",
		Description: "Retrieves a list of all brands",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.ListBrands)

	// Get brand by ID
	huma.Register(api, huma.Operation{
		OperationID: "get-brand",
		Method:      http.MethodGet,
		Path:        "/brands/{id}",
		Summary:     "Get a brand by ID",
		Description: "Retrieves a brand by its unique identifier",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetBrand)

	// Get brand by name
	huma.Register(api, huma.Operation{
		OperationID: "get-brand-by-name",
		Method:      http.MethodGet,
		Path:        "/brands/name/{name}",
		Summary:     "Get a brand by name",
		Description: "Retrieves a brand by its name",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetBrandByName)

	// Update brand
	huma.Register(api, huma.Operation{
		OperationID: "update-brand",
		Method:      http.MethodPut,
		Path:        "/brands/{id}",
		Summary:     "Update a brand",
		Description: "Updates an existing brand's information",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.UpdateBrand)

	// Delete brand
	huma.Register(api, huma.Operation{
		OperationID: "delete-brand",
		Method:      http.MethodDelete,
		Path:        "/brands/{id}",
		Summary:     "Delete a brand",
		Description: "Deletes a brand by its unique identifier",
		Tags:        []string{"Brands"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.DeleteBrand)
}

// CreateBrand handles POST /brands
func (h *BrandHandler) CreateBrand(ctx context.Context, input *dto.CreateBrandRequest) (*dto.BrandResponse, error) {
	brand := &entities.Brand{
		Name: input.Body.Name,
	}

	err := h.service.CreateBrand(ctx, brand)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Brand with this name already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to create brand", err)
	}

	return h.mapToResponse(brand), nil
}

// ListBrands handles GET /brands
func (h *BrandHandler) ListBrands(ctx context.Context, _ *struct{}) (*dto.ListBrandsResponse, error) {
	brands, err := h.service.ListBrands(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list brands", err)
	}

	response := &dto.ListBrandsResponse{}
	response.Body.Brands = make([]dto.BrandListItem, 0, len(brands))

	for _, brand := range brands {
		response.Body.Brands = append(response.Body.Brands, dto.BrandListItem{
			ID:        brand.ID,
			Name:      brand.Name,
			CreatedAt: brand.CreatedAt,
			UpdatedAt: brand.UpdatedAt,
		})
	}

	return response, nil
}

// GetBrand handles GET /brands/{id}
func (h *BrandHandler) GetBrand(ctx context.Context, input *dto.GetBrandRequest) (*dto.BrandResponse, error) {
	brand, err := h.service.GetBrand(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Brand not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get brand", err)
	}

	return h.mapToResponse(brand), nil
}

// GetBrandByName handles GET /brands/name/{name}
func (h *BrandHandler) GetBrandByName(ctx context.Context, input *dto.GetBrandByNameRequest) (*dto.BrandResponse, error) {
	brand, err := h.service.GetBrandByName(ctx, input.Name)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Brand not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get brand", err)
	}

	return h.mapToResponse(brand), nil
}

// UpdateBrand handles PUT /brands/{id}
func (h *BrandHandler) UpdateBrand(ctx context.Context, input *dto.UpdateBrandRequest) (*dto.BrandResponse, error) {
	brand := &entities.Brand{
		ID:   input.ID,
		Name: input.Body.Name,
	}

	err := h.service.UpdateBrand(ctx, brand)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Brand not found")
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Brand with this name already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to update brand", err)
	}

	return h.mapToResponse(brand), nil
}

// DeleteBrand handles DELETE /brands/{id}
func (h *BrandHandler) DeleteBrand(ctx context.Context, input *dto.DeleteBrandRequest) (*struct{}, error) {
	err := h.service.DeleteBrand(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Brand not found")
		}
		return nil, huma.Error500InternalServerError("Failed to delete brand", err)
	}

	return &struct{}{}, nil
}

// mapToResponse converts a domain Brand entity to a BrandResponse DTO
func (h *BrandHandler) mapToResponse(brand *entities.Brand) *dto.BrandResponse {
	response := &dto.BrandResponse{}
	response.Body.ID = brand.ID
	response.Body.Name = brand.Name
	response.Body.CreatedAt = brand.CreatedAt
	response.Body.UpdatedAt = brand.UpdatedAt
	return response
}
