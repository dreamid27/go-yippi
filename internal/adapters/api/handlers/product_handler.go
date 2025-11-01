package handlers

import (
	"context"
	"errors"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/danielgtaylor/huma/v2"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// RegisterRoutes registers all product routes with Huma
func (h *ProductHandler) RegisterRoutes(api huma.API) {
	// Create product
	huma.Register(api, huma.Operation{
		OperationID: "create-product",
		Method:      http.MethodPost,
		Path:        "/products",
		Summary:     "Create a new product",
		Description: "Creates a new product with SKU, name, price, and shipping details",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError},
	}, h.CreateProduct)

	// List all products
	huma.Register(api, huma.Operation{
		OperationID: "list-products",
		Method:      http.MethodGet,
		Path:        "/products",
		Summary:     "List all products",
		Description: "Retrieves a list of all products in the system",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.ListProducts)

	// Get product by ID
	huma.Register(api, huma.Operation{
		OperationID: "get-product",
		Method:      http.MethodGet,
		Path:        "/products/{id}",
		Summary:     "Get a product by ID",
		Description: "Retrieves a product by its ID",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetProduct)

	// Get product by SKU
	huma.Register(api, huma.Operation{
		OperationID: "get-product-by-sku",
		Method:      http.MethodGet,
		Path:        "/products/sku/{sku}",
		Summary:     "Get a product by SKU",
		Description: "Retrieves a product by its Stock Keeping Unit",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetProductBySKU)

	// Get product by slug
	huma.Register(api, huma.Operation{
		OperationID: "get-product-by-slug",
		Method:      http.MethodGet,
		Path:        "/products/slug/{slug}",
		Summary:     "Get a product by slug",
		Description: "Retrieves a product by its URL-friendly slug",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetProductBySlug)

	// List products by status
	huma.Register(api, huma.Operation{
		OperationID: "list-products-by-status",
		Method:      http.MethodGet,
		Path:        "/products/status/{status}",
		Summary:     "List products by status",
		Description: "Retrieves products filtered by status (draft, published, archived)",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.ListProductsByStatus)

	// Update product
	huma.Register(api, huma.Operation{
		OperationID: "update-product",
		Method:      http.MethodPut,
		Path:        "/products/{id}",
		Summary:     "Update a product",
		Description: "Updates an existing product's information",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.UpdateProduct)

	// Publish product
	huma.Register(api, huma.Operation{
		OperationID: "publish-product",
		Method:      http.MethodPost,
		Path:        "/products/{id}/publish",
		Summary:     "Publish a product",
		Description: "Changes product status from draft to published",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
	}, h.PublishProduct)

	// Archive product
	huma.Register(api, huma.Operation{
		OperationID: "archive-product",
		Method:      http.MethodPost,
		Path:        "/products/{id}/archive",
		Summary:     "Archive a product",
		Description: "Changes product status to archived",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.ArchiveProduct)

	// Delete product
	huma.Register(api, huma.Operation{
		OperationID:   "delete-product",
		Method:        http.MethodDelete,
		Path:          "/products/{id}",
		Summary:       "Delete a product",
		Description:   "Permanently deletes a product from the system",
		Tags:          []string{"Products"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.DeleteProduct)
}

func (h *ProductHandler) CreateProduct(ctx context.Context, input *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &entities.Product{
		SKU:         input.Body.SKU,
		Name:        input.Body.Name,
		Price:       input.Body.Price,
		Description: input.Body.Description,
	}

	// Handle optional slug (will be auto-generated in service if not provided)
	if input.Body.Slug != nil {
		product.Slug = *input.Body.Slug
	}

	// Handle optional dimensions
	if input.Body.Weight != nil {
		product.Weight = *input.Body.Weight
	}
	if input.Body.Length != nil {
		product.Length = *input.Body.Length
	}
	if input.Body.Width != nil {
		product.Width = *input.Body.Width
	}
	if input.Body.Height != nil {
		product.Height = *input.Body.Height
	}

	// Handle optional status (will default to draft in service if not provided)
	if input.Body.Status != nil {
		product.Status = entities.ProductStatus(*input.Body.Status)
	}

	err := h.service.CreateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Product with this SKU or slug already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to create product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, input *struct{}) (*dto.ListProductsResponse, error) {
	products, err := h.service.ListProducts(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list products", err)
	}

	resp := &dto.ListProductsResponse{}
	resp.Body.Products = make([]dto.ProductListItem, len(products))

	for i, product := range products {
		resp.Body.Products[i] = dto.ProductListItem{
			ID:          product.ID,
			SKU:         product.SKU,
			Slug:        product.Slug,
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			Weight:      product.Weight,
			Length:      product.Length,
			Width:       product.Width,
			Height:      product.Height,
			Status:      string(product.Status),
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		}
	}

	return resp, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, input *dto.GetProductRequest) (*dto.ProductResponse, error) {
	product, err := h.service.GetProduct(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) GetProductBySKU(ctx context.Context, input *dto.GetProductBySKURequest) (*dto.ProductResponse, error) {
	product, err := h.service.GetProductBySKU(ctx, input.SKU)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid SKU", err)
		}
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) GetProductBySlug(ctx context.Context, input *dto.GetProductBySlugRequest) (*dto.ProductResponse, error) {
	product, err := h.service.GetProductBySlug(ctx, input.Slug)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid slug", err)
		}
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) ListProductsByStatus(ctx context.Context, input *dto.ListProductsByStatusRequest) (*dto.ListProductsResponse, error) {
	products, err := h.service.ListProductsByStatus(ctx, entities.ProductStatus(input.Status))
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list products", err)
	}

	resp := &dto.ListProductsResponse{}
	resp.Body.Products = make([]dto.ProductListItem, len(products))

	for i, product := range products {
		resp.Body.Products[i] = dto.ProductListItem{
			ID:          product.ID,
			SKU:         product.SKU,
			Slug:        product.Slug,
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			Weight:      product.Weight,
			Length:      product.Length,
			Width:       product.Width,
			Height:      product.Height,
			Status:      string(product.Status),
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		}
	}

	return resp, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, input *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product := &entities.Product{
		ID:          input.ID,
		SKU:         input.Body.SKU,
		Name:        input.Body.Name,
		Price:       input.Body.Price,
		Description: input.Body.Description,
	}

	// Handle optional slug (will be auto-generated in service if not provided)
	if input.Body.Slug != nil {
		product.Slug = *input.Body.Slug
	}

	// Handle optional dimensions
	if input.Body.Weight != nil {
		product.Weight = *input.Body.Weight
	}
	if input.Body.Length != nil {
		product.Length = *input.Body.Length
	}
	if input.Body.Width != nil {
		product.Width = *input.Body.Width
	}
	if input.Body.Height != nil {
		product.Height = *input.Body.Height
	}

	// Handle optional status (will default to draft in service if not provided)
	if input.Body.Status != nil {
		product.Status = entities.ProductStatus(*input.Body.Status)
	}

	err := h.service.UpdateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Product with this SKU or slug already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to update product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) PublishProduct(ctx context.Context, input *dto.PublishProductRequest) (*dto.ProductResponse, error) {
	err := h.service.PublishProduct(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Cannot publish this product", err)
		}
		return nil, huma.Error500InternalServerError("Failed to publish product", err)
	}

	product, err := h.service.GetProduct(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get updated product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) ArchiveProduct(ctx context.Context, input *dto.ArchiveProductRequest) (*dto.ProductResponse, error) {
	err := h.service.ArchiveProduct(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to archive product", err)
	}

	product, err := h.service.GetProduct(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get updated product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, input *dto.DeleteProductRequest) (*struct{}, error) {
	err := h.service.DeleteProduct(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to delete product", err)
	}

	return &struct{}{}, nil
}

// mapToResponse converts domain entity to DTO response
func (h *ProductHandler) mapToResponse(product *entities.Product) *dto.ProductResponse {
	resp := &dto.ProductResponse{}
	resp.Body.ID = product.ID
	resp.Body.SKU = product.SKU
	resp.Body.Slug = product.Slug
	resp.Body.Name = product.Name
	resp.Body.Price = product.Price
	resp.Body.Description = product.Description
	resp.Body.Weight = product.Weight
	resp.Body.Length = product.Length
	resp.Body.Width = product.Width
	resp.Body.Height = product.Height
	resp.Body.Status = string(product.Status)
	resp.Body.CreatedAt = product.CreatedAt
	resp.Body.UpdatedAt = product.UpdatedAt
	return resp
}
