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
	"github.com/google/uuid"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service        ports.ProductService
	variantService ports.ProductVariantService
}

func NewProductHandler(service ports.ProductService, variantService ports.ProductVariantService) *ProductHandler {
	return &ProductHandler{
		service:        service,
		variantService: variantService,
	}
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

	// Search products with filters and pagination (Phase 2)
	huma.Register(api, huma.Operation{
		OperationID: "search-products",
		Method:      http.MethodGet,
		Path:        "/products",
		Summary:     "Search products with filtering and pagination",
		Description: "Search products with full-text search, category/brand/price filters, variant attribute filters, and pagination. Sort by relevance when search query provided.",
		Tags:        []string{"Products"},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.SearchProducts)

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
		Name:        input.Body.Name,
		BasePrice:   input.Body.BasePrice,
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

	// Handle image URLs
	if input.Body.ImageURLs != nil {
		product.ImageURLs = input.Body.ImageURLs
	}

	// Handle optional status (will default to draft in service if not provided)
	if input.Body.Status != nil {
		product.Status = entities.ProductStatus(*input.Body.Status)
	}

	// Handle optional category ID
	if input.Body.CategoryID != nil && *input.Body.CategoryID != "" {
		categoryUUID, err := uuid.Parse(*input.Body.CategoryID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid category_id UUID format", err)
		}
		product.CategoryID = &categoryUUID
	}

	// Handle optional brand ID
	if input.Body.BrandID != nil && *input.Body.BrandID != "" {
		brandUUID, err := uuid.Parse(*input.Body.BrandID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid brand_id UUID format", err)
		}
		product.BrandID = &brandUUID
	}

	// Handle optional stock quantity for Phase 2
	if input.Body.StockQuantity != nil {
		product.StockQuantity = input.Body.StockQuantity
	}

	// Create product
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

	// Phase 2: Auto-create default variant (REQ-9.5)
	stockQty := 0
	if product.StockQuantity != nil {
		stockQty = *product.StockQuantity
	}

	defaultSKU := "PRODUCT-" + product.Slug
	isActive := true
	_, err = h.variantService.CreateVariant(
		ctx,
		product.ID,
		defaultSKU,
		map[string]string{}, // empty attributes for default variant
		stockQty,
		0.0, // no price adjustment for default variant
		&isActive,
	)
	if err != nil {
		// If variant creation fails, we should probably rollback the product
		// For now, log the error and continue (product exists without variants)
		// TODO: Consider using transaction to rollback product creation
		return nil, huma.Error500InternalServerError("Failed to create default variant", err)
	}

	return h.mapToResponse(product), nil
}

// SearchProducts handles product search with filters (REQ-9.3)
func (h *ProductHandler) SearchProducts(ctx context.Context, input *dto.SearchProductsRequest) (*dto.ProductSearchResponse, error) {
	// Validate and parse UUID filters
	var categoryUUID, brandUUID *uuid.UUID

	if input.CategoryID != nil && *input.CategoryID != "" {
		parsed, err := uuid.Parse(*input.CategoryID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid category_id UUID format", err)
		}
		categoryUUID = &parsed
	}

	if input.BrandID != nil && *input.BrandID != "" {
		parsed, err := uuid.Parse(*input.BrandID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid brand_id UUID format", err)
		}
		brandUUID = &parsed
	}

	// Build search params
	params := ports.SearchProductsParams{
		Search:     input.Search,
		CategoryID: categoryUUID,
		BrandID:    brandUUID,
		MinPrice:   input.MinPrice,
		MaxPrice:   input.MaxPrice,
		Size:       input.Size,
		Color:      input.Color,
		Status:     input.Status,
		SortBy:     input.SortBy,
		SortOrder:  input.SortOrder,
		Page:       input.Page,
		Limit:      input.Limit,
	}

	// Apply defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// If search query provided, default to relevance sorting
	if params.Search != "" && input.SortBy == "" {
		params.SortBy = "relevance"
	}

	// Call service
	result, err := h.service.SearchProducts(ctx, params)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid search parameters", err)
		}
		return nil, huma.Error500InternalServerError("Failed to search products", err)
	}

	// Convert to DTO response
	resp := &dto.ProductSearchResponse{}
	resp.Body.Data = make([]dto.ProductListItemResponse, len(result.Products))

	for i, product := range result.Products {
		listItem := dto.ProductListItemResponse{
			ID:            product.ID.String(),
			Slug:          product.Slug,
			Name:          product.Name,
			BasePrice:     product.BasePrice,
			Description:   product.Description,
			ImageURLs:     product.ImageURLs,
			Status:        product.Status,
			VariantsCount: product.VariantsCount,
			MinPrice:      product.MinPrice,
			MaxPrice:      product.MaxPrice,
			HasStock:      product.HasStock,
		}

		// Map category summary
		if product.Category != nil {
			listItem.Category = &dto.CategorySummaryDTO{
				ID:   product.Category.ID.String(),
				Name: product.Category.Name,
				Slug: product.Category.Slug,
			}
		}

		// Map brand summary
		if product.Brand != nil {
			listItem.Brand = &dto.BrandSummaryDTO{
				ID:   product.Brand.ID.String(),
				Name: product.Brand.Name,
				Slug: product.Brand.Slug,
			}
		}

		resp.Body.Data[i] = listItem
	}

	// Convert pagination info
	resp.Body.Pagination = dto.PaginationResponse{
		Page:       result.Pagination.Page,
		Limit:      result.Pagination.Limit,
		Total:      result.Pagination.Total,
		TotalPages: result.Pagination.TotalPages,
	}

	return resp, nil
}

// GetProduct returns product detail with variants (REQ-9.4)
func (h *ProductHandler) GetProduct(ctx context.Context, input *dto.GetProductRequest) (*dto.ProductDetailResponse, error) {
	// Parse UUID from string
	productID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product ID UUID format", err)
	}

	// Get product with variants
	productDetail, err := h.service.GetProductWithVariants(ctx, productID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	return h.mapToDetailResponse(productDetail), nil
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
		listItem := dto.ProductListItem{
			ID:          product.ID.String(),
			Slug:        product.Slug,
			Name:        product.Name,
			BasePrice:   product.BasePrice,
			Description: product.Description,
			Weight:      product.Weight,
			Length:      product.Length,
			Width:       product.Width,
			Height:      product.Height,
			ImageURLs:   product.ImageURLs,
			Status:      string(product.Status),
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		}

		// Convert UUID pointers to string pointers
		if product.CategoryID != nil {
			categoryIDStr := product.CategoryID.String()
			listItem.CategoryID = &categoryIDStr
		}
		if product.BrandID != nil {
			brandIDStr := product.BrandID.String()
			listItem.BrandID = &brandIDStr
		}

		resp.Body.Products[i] = listItem
	}

	return resp, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, input *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	// Parse UUID from string
	productID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product ID UUID format", err)
	}

	product := &entities.Product{
		ID:          productID,
		Name:        input.Body.Name,
		BasePrice:   input.Body.BasePrice,
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

	// Handle image URLs
	if input.Body.ImageURLs != nil {
		product.ImageURLs = input.Body.ImageURLs
	}

	// Handle optional status (will default to draft in service if not provided)
	if input.Body.Status != nil {
		product.Status = entities.ProductStatus(*input.Body.Status)
	}

	// Handle optional category ID
	if input.Body.CategoryID != nil && *input.Body.CategoryID != "" {
		categoryUUID, err := uuid.Parse(*input.Body.CategoryID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid category_id UUID format", err)
		}
		product.CategoryID = &categoryUUID
	}

	// Handle optional brand ID
	if input.Body.BrandID != nil && *input.Body.BrandID != "" {
		brandUUID, err := uuid.Parse(*input.Body.BrandID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid brand_id UUID format", err)
		}
		product.BrandID = &brandUUID
	}

	err = h.service.UpdateProduct(ctx, product)
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
	// Parse UUID from string
	productID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product ID UUID format", err)
	}

	err = h.service.PublishProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Cannot publish this product", err)
		}
		return nil, huma.Error500InternalServerError("Failed to publish product", err)
	}

	product, err := h.service.GetProduct(ctx, productID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get updated product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) ArchiveProduct(ctx context.Context, input *dto.ArchiveProductRequest) (*dto.ProductResponse, error) {
	// Parse UUID from string
	productID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product ID UUID format", err)
	}

	err = h.service.ArchiveProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to archive product", err)
	}

	product, err := h.service.GetProduct(ctx, productID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get updated product", err)
	}

	return h.mapToResponse(product), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, input *dto.DeleteProductRequest) (*struct{}, error) {
	// Parse UUID from string
	productID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product ID UUID format", err)
	}

	err = h.service.DeleteProduct(ctx, productID)
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
	resp.Body.ID = product.ID.String()
	resp.Body.Slug = product.Slug
	resp.Body.Name = product.Name
	resp.Body.BasePrice = product.BasePrice
	resp.Body.Description = product.Description
	resp.Body.Weight = product.Weight
	resp.Body.Length = product.Length
	resp.Body.Width = product.Width
	resp.Body.Height = product.Height
	resp.Body.ImageURLs = product.ImageURLs
	resp.Body.Status = string(product.Status)

	// Convert UUID pointers to string pointers
	if product.CategoryID != nil {
		categoryIDStr := product.CategoryID.String()
		resp.Body.CategoryID = &categoryIDStr
	}
	if product.BrandID != nil {
		brandIDStr := product.BrandID.String()
		resp.Body.BrandID = &brandIDStr
	}
	resp.Body.CreatedAt = product.CreatedAt
	resp.Body.UpdatedAt = product.UpdatedAt
	return resp
}

// mapToDetailResponse converts ProductDetail to ProductDetailResponse (REQ-9.4)
func (h *ProductHandler) mapToDetailResponse(detail *ports.ProductDetail) *dto.ProductDetailResponse {
	resp := &dto.ProductDetailResponse{}

	// Map product fields
	resp.Body.ID = detail.Product.ID.String()
	resp.Body.Slug = detail.Product.Slug
	resp.Body.Name = detail.Product.Name
	resp.Body.BasePrice = detail.Product.BasePrice
	resp.Body.Description = detail.Product.Description
	resp.Body.ImageURLs = detail.Product.ImageURLs
	resp.Body.Status = string(detail.Product.Status)
	resp.Body.Weight = detail.Product.Weight
	resp.Body.CreatedAt = detail.Product.CreatedAt
	resp.Body.UpdatedAt = detail.Product.UpdatedAt

	// Map dimensions
	resp.Body.Dimensions = dto.DimensionsResponse{
		Length: detail.Product.Length,
		Width:  detail.Product.Width,
		Height: detail.Product.Height,
	}

	// Map category if exists
	if detail.Category != nil {
		categoryResp := &dto.CategoryResponse{}
		categoryResp.Body.ID = detail.Category.ID.String()
		categoryResp.Body.Name = detail.Category.Name
		categoryResp.Body.Slug = entities.GenerateSlug(detail.Category.Name)
		categoryResp.Body.Description = ""
		categoryResp.Body.CreatedAt = detail.Category.CreatedAt
		categoryResp.Body.UpdatedAt = detail.Category.UpdatedAt
		resp.Body.Category = categoryResp
	}

	// Map brand if exists
	if detail.Brand != nil {
		brandResp := &dto.BrandResponse{}
		brandResp.Body.ID = detail.Brand.ID
		brandResp.Body.Name = detail.Brand.Name
		brandResp.Body.Slug = entities.GenerateSlug(detail.Brand.Name)
		brandResp.Body.Description = ""
		brandResp.Body.CreatedAt = detail.Brand.CreatedAt
		brandResp.Body.UpdatedAt = detail.Brand.UpdatedAt
		resp.Body.Brand = brandResp
	}

	// Map variants
	resp.Body.Variants = make([]dto.ProductVariantResponse, len(detail.Variants))
	for i, variant := range detail.Variants {
		resp.Body.Variants[i] = dto.ProductVariantResponse{
			ID:              variant.ID.String(),
			ProductID:       detail.Product.ID.String(),
			SKU:             variant.SKU,
			Attributes:      variant.Attributes,
			StockQuantity:   variant.StockQuantity,
			PriceAdjustment: variant.PriceAdjustment,
			FinalPrice:      variant.FinalPrice,
			IsActive:        variant.IsActive,
			IsInStock:       variant.IsInStock,
			CreatedAt:       detail.Product.CreatedAt, // TODO: variants should have their own timestamps
			UpdatedAt:       detail.Product.UpdatedAt,
		}
	}

	return resp
}
