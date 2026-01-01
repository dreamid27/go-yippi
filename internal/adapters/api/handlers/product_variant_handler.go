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

// ProductVariantHandler handles HTTP requests for product variants
type ProductVariantHandler struct {
	variantService ports.ProductVariantService
	productService ports.ProductService
}

// NewProductVariantHandler creates a new product variant handler
func NewProductVariantHandler(
	variantService ports.ProductVariantService,
	productService ports.ProductService,
) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantService: variantService,
		productService: productService,
	}
}

// RegisterRoutes registers all product variant routes with Huma
func (h *ProductVariantHandler) RegisterRoutes(api huma.API) {
	// List variants for a product (Public)
	huma.Register(api, huma.Operation{
		OperationID: "list-product-variants",
		Method:      http.MethodGet,
		Path:        "/products/{product_id}/variants",
		Summary:     "List product variants",
		Description: "Retrieves all variants for a specific product",
		Tags:        []string{"Product Variants"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
	}, h.ListVariants)

	// Create variant (Admin only)
	huma.Register(api, huma.Operation{
		OperationID: "create-product-variant",
		Method:      http.MethodPost,
		Path:        "/products/{product_id}/variants",
		Summary:     "Create product variant",
		Description: "Creates a new variant for a product (Admin only)",
		Tags:        []string{"Product Variants"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.CreateVariant)

	// Update variant (Admin only)
	huma.Register(api, huma.Operation{
		OperationID: "update-product-variant",
		Method:      http.MethodPut,
		Path:        "/products/{product_id}/variants/{variant_id}",
		Summary:     "Update product variant",
		Description: "Updates an existing product variant (Admin only)",
		Tags:        []string{"Product Variants"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.UpdateVariant)

	// Update stock (Admin only)
	huma.Register(api, huma.Operation{
		OperationID: "update-variant-stock",
		Method:      http.MethodPatch,
		Path:        "/products/{product_id}/variants/{variant_id}/stock",
		Summary:     "Update variant stock",
		Description: "Updates stock quantity for a variant (Admin only)",
		Tags:        []string{"Product Variants"},
		Errors:      []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError},
	}, h.UpdateStock)

	// Delete variant (Admin only)
	huma.Register(api, huma.Operation{
		OperationID:   "delete-product-variant",
		Method:        http.MethodDelete,
		Path:          "/products/{product_id}/variants/{variant_id}",
		Summary:       "Delete product variant",
		Description:   "Soft deletes a product variant (Admin only)",
		Tags:          []string{"Product Variants"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError},
	}, h.DeleteVariant)
}

// ListVariants returns all variants for a product
func (h *ProductVariantHandler) ListVariants(ctx context.Context, input *dto.GetVariantsRequest) (*dto.ProductVariantListResponse, error) {
	// Parse product ID
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product_id UUID format", err)
	}

	// Get product to retrieve base price
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	// Get variants
	variants, err := h.variantService.GetVariantsByProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get variants", err)
	}

	// Map to response
	resp := &dto.ProductVariantListResponse{}
	resp.Body.Data = make([]dto.ProductVariantDetail, len(variants))

	for i, variant := range variants {
		resp.Body.Data[i] = dto.ProductVariantDetail{
			ID:              variant.ID.String(),
			ProductID:       variant.ProductID.String(),
			SKU:             variant.SKU,
			Attributes:      variant.Attributes,
			StockQuantity:   variant.StockQuantity,
			PriceAdjustment: variant.PriceAdjustment,
			FinalPrice:      variant.CalculateFinalPrice(product.BasePrice),
			IsActive:        variant.IsActive,
			IsInStock:       variant.IsInStock(),
			CreatedAt:       variant.CreatedAt,
			UpdatedAt:       variant.UpdatedAt,
		}
	}

	return resp, nil
}

// CreateVariant creates a new product variant
func (h *ProductVariantHandler) CreateVariant(ctx context.Context, input *dto.CreateProductVariantRequest) (*dto.VariantDetailResponse, error) {
	// Parse product ID
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product_id UUID format", err)
	}

	// Validate attributes are not empty
	if len(input.Body.Attributes) == 0 {
		return nil, huma.Error400BadRequest("Attributes are required and cannot be empty")
	}

	// Create variant via service
	variant, err := h.variantService.CreateVariant(
		ctx,
		productID,
		input.Body.SKU,
		input.Body.Attributes,
		input.Body.StockQuantity,
		input.Body.PriceAdjustment,
		input.Body.IsActive,
	)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("SKU already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to create variant", err)
	}

	// Get product to calculate final price
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	// Map to response
	return h.mapToResponse(variant, product.BasePrice), nil
}

// UpdateVariant updates an existing product variant
func (h *ProductVariantHandler) UpdateVariant(ctx context.Context, input *dto.UpdateProductVariantRequest) (*dto.VariantDetailResponse, error) {
	// Parse product ID
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product_id UUID format", err)
	}

	// Parse variant ID
	variantID, err := uuid.Parse(input.VariantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid variant_id UUID format", err)
	}

	// Validate attributes if provided
	if input.Body.Attributes != nil && len(*input.Body.Attributes) == 0 {
		return nil, huma.Error400BadRequest("Attributes cannot be empty if provided")
	}

	// Update variant via service
	variant, err := h.variantService.UpdateVariant(
		ctx,
		variantID,
		input.Body.SKU,
		input.Body.Attributes,
		input.Body.StockQuantity,
		input.Body.PriceAdjustment,
		input.Body.IsActive,
	)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Variant not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("SKU already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to update variant", err)
	}

	// Get product to calculate final price
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	// Map to response
	return h.mapToResponse(variant, product.BasePrice), nil
}

// UpdateStock updates stock quantity for a variant
func (h *ProductVariantHandler) UpdateStock(ctx context.Context, input *dto.UpdateStockRequest) (*dto.VariantDetailResponse, error) {
	// Parse product ID
	productID, err := uuid.Parse(input.ProductID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product_id UUID format", err)
	}

	// Parse variant ID
	variantID, err := uuid.Parse(input.VariantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid variant_id UUID format", err)
	}

	// Update stock via service
	variant, err := h.variantService.UpdateStock(ctx, variantID, input.Body.StockQuantity)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Variant not found")
		}
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid stock quantity", err)
		}
		return nil, huma.Error500InternalServerError("Failed to update stock", err)
	}

	// Get product to calculate final price
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get product", err)
	}

	// Map to response
	return h.mapToResponse(variant, product.BasePrice), nil
}

// DeleteVariant deletes a product variant
func (h *ProductVariantHandler) DeleteVariant(ctx context.Context, input *dto.DeleteVariantRequest) (*struct{}, error) {
	// Parse variant ID
	variantID, err := uuid.Parse(input.VariantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid variant_id UUID format", err)
	}

	// Delete variant via service
	err = h.variantService.DeleteVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Variant not found")
		}
		// TODO: Check for ErrVariantInUse when cart/order system is implemented
		return nil, huma.Error500InternalServerError("Failed to delete variant", err)
	}

	return &struct{}{}, nil
}

// mapToResponse converts domain entity to DTO response
func (h *ProductVariantHandler) mapToResponse(variant *entities.ProductVariant, basePrice float64) *dto.VariantDetailResponse {
	resp := &dto.VariantDetailResponse{}
	resp.Body = dto.ProductVariantDetail{
		ID:              variant.ID.String(),
		ProductID:       variant.ProductID.String(),
		SKU:             variant.SKU,
		Attributes:      variant.Attributes,
		StockQuantity:   variant.StockQuantity,
		PriceAdjustment: variant.PriceAdjustment,
		FinalPrice:      variant.CalculateFinalPrice(basePrice),
		IsActive:        variant.IsActive,
		IsInStock:       variant.IsInStock(),
		CreatedAt:       variant.CreatedAt,
		UpdatedAt:       variant.UpdatedAt,
	}
	return resp
}
