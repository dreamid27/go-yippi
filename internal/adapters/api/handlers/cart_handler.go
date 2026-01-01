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

// CartHandler handles HTTP requests for shopping cart operations
type CartHandler struct {
	service ports.CartService
}

// NewCartHandler creates a new CartHandler instance
func NewCartHandler(service ports.CartService) *CartHandler {
	return &CartHandler{service: service}
}

// RegisterRoutes registers all cart routes with Huma
func (h *CartHandler) RegisterRoutes(api huma.API) {
	// Get user's cart
	huma.Register(api, huma.Operation{
		OperationID: "get-cart",
		Method:      http.MethodGet,
		Path:        "/cart",
		Summary:     "Get user's cart",
		Description: "Retrieves the authenticated user's shopping cart with all items",
		Tags:        []string{"Cart"},
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.GetCart)

	// Add item to cart
	huma.Register(api, huma.Operation{
		OperationID: "add-cart-item",
		Method:      http.MethodPost,
		Path:        "/cart/items",
		Summary:     "Add item to cart",
		Description: "Adds a product variant to the authenticated user's cart",
		Tags:        []string{"Cart"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, h.AddCartItem)

	// Update cart item quantity
	huma.Register(api, huma.Operation{
		OperationID: "update-cart-item",
		Method:      http.MethodPut,
		Path:        "/cart/items/{item_id}",
		Summary:     "Update cart item quantity",
		Description: "Updates the quantity of a cart item. If quantity is 0, the item is removed and 204 No Content is returned.",
		Tags:        []string{"Cart"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
	}, h.UpdateCartItem)

	// Remove cart item
	huma.Register(api, huma.Operation{
		OperationID:   "remove-cart-item",
		Method:        http.MethodDelete,
		Path:          "/cart/items/{item_id}",
		Summary:       "Remove item from cart",
		Description:   "Removes a specific item from the authenticated user's cart",
		Tags:          []string{"Cart"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError},
	}, h.RemoveCartItem)

	// Clear cart
	huma.Register(api, huma.Operation{
		OperationID:   "clear-cart",
		Method:        http.MethodDelete,
		Path:          "/cart",
		Summary:       "Clear cart",
		Description:   "Removes all items from the authenticated user's cart",
		Tags:          []string{"Cart"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.ClearCart)

	// Merge guest cart
	huma.Register(api, huma.Operation{
		OperationID: "merge-guest-cart",
		Method:      http.MethodPost,
		Path:        "/cart/merge",
		Summary:     "Merge guest cart",
		Description: "Merges guest cart items into the authenticated user's cart after login",
		Tags:        []string{"Cart"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
	}, h.MergeGuestCart)
}

// GetCart retrieves the authenticated user's cart
func (h *CartHandler) GetCart(ctx context.Context, input *struct{}) (*dto.CartResponse, error) {
	// Extract user from context (set by auth middleware)
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Get cart with details
	cartDetail, err := h.service.GetCartWithDetails(ctx, user.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve cart", err)
	}

	// Build response
	resp := &dto.CartResponse{}
	resp.Body.ID = cartDetail.ID.String()
	resp.Body.UserID = cartDetail.UserID.String()
	resp.Body.Items = convertCartItemsToDTO(cartDetail.Items)
	resp.Body.Subtotal = cartDetail.Subtotal
	resp.Body.ItemCount = cartDetail.ItemCount

	return resp, nil
}

// AddCartItem adds an item to the user's cart
func (h *CartHandler) AddCartItem(ctx context.Context, input *dto.AddCartItemRequest) (*dto.CartItemResponseSingle, error) {
	// Extract user from context
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Parse variant ID
	variantID, err := uuid.Parse(input.Body.ProductVariantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid product_variant_id UUID format", err)
	}

	// Add item to cart
	cartItem, err := h.service.AddItemToCart(ctx, user.ID, variantID, input.Body.Quantity)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		// Check for not found errors
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product variant not found or inactive")
		}
		// Check for stock validation errors
		if isStockError(err) {
			return nil, huma.Error400BadRequest("Insufficient stock", err)
		}
		return nil, huma.Error500InternalServerError("Failed to add item to cart", err)
	}

	// Get full cart details to return enriched item
	cartDetail, err := h.service.GetCartWithDetails(ctx, user.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve cart details", err)
	}

	// Find the added/updated item in cart details
	var itemDetail *ports.CartItemDetail
	for _, item := range cartDetail.Items {
		if item.ID == cartItem.ID {
			itemDetail = item
			break
		}
	}

	if itemDetail == nil {
		return nil, huma.Error500InternalServerError("Failed to find added item in cart")
	}

	// Build response
	resp := &dto.CartItemResponseSingle{}
	itemDTO := convertCartItemDetailToDTO(itemDetail)
	resp.Body = *itemDTO
	return resp, nil
}

// UpdateCartItem updates the quantity of a cart item
func (h *CartHandler) UpdateCartItem(ctx context.Context, input *dto.UpdateCartItemRequest) (*dto.CartItemResponseSingle, error) {
	// Extract user from context
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Parse cart item ID
	cartItemID, err := uuid.Parse(input.ItemID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid item_id UUID format", err)
	}

	// Update cart item quantity
	updatedItem, err := h.service.UpdateCartItemQuantity(ctx, user.ID, cartItemID, input.Body.Quantity)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		// Check for not found errors
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Cart item not found")
		}
		// Check for authorization errors (item doesn't belong to user)
		if errors.Is(err, domainErrors.ErrForbidden) {
			return nil, huma.Error403Forbidden("You don't have permission to modify this cart item")
		}
		// Check for stock validation errors
		if isStockError(err) {
			return nil, huma.Error400BadRequest("Insufficient stock", err)
		}
		return nil, huma.Error500InternalServerError("Failed to update cart item", err)
	}

	// If quantity was 0, item was deleted - return 204 No Content
	if input.Body.Quantity == 0 {
		// For Huma v2, we need to return an empty response with 204 status
		// This is handled by Huma's DefaultStatus in operation config
		return nil, nil
	}

	// Get full cart details to return enriched item
	cartDetail, err := h.service.GetCartWithDetails(ctx, user.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve cart details", err)
	}

	// Find the updated item in cart details
	var itemDetail *ports.CartItemDetail
	for _, item := range cartDetail.Items {
		if item.ID == updatedItem.ID {
			itemDetail = item
			break
		}
	}

	if itemDetail == nil {
		return nil, huma.Error500InternalServerError("Failed to find updated item in cart")
	}

	// Build response
	resp := &dto.CartItemResponseSingle{}
	itemDTO := convertCartItemDetailToDTO(itemDetail)
	resp.Body = *itemDTO
	return resp, nil
}

// RemoveCartItem removes an item from the cart
func (h *CartHandler) RemoveCartItem(ctx context.Context, input *dto.RemoveCartItemRequest) (*struct{}, error) {
	// Extract user from context
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Parse cart item ID
	cartItemID, err := uuid.Parse(input.ItemID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid item_id UUID format", err)
	}

	// Remove cart item
	err = h.service.RemoveCartItem(ctx, user.ID, cartItemID)
	if err != nil {
		// Check for not found errors
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Cart item not found")
		}
		// Check for authorization errors
		if errors.Is(err, domainErrors.ErrForbidden) {
			return nil, huma.Error403Forbidden("You don't have permission to remove this cart item")
		}
		return nil, huma.Error500InternalServerError("Failed to remove cart item", err)
	}

	return &struct{}{}, nil
}

// ClearCart clears all items from the user's cart
func (h *CartHandler) ClearCart(ctx context.Context, input *struct{}) (*struct{}, error) {
	// Extract user from context
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Clear cart
	err := h.service.ClearCart(ctx, user.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to clear cart", err)
	}

	return &struct{}{}, nil
}

// MergeGuestCart merges guest cart items into user's cart
func (h *CartHandler) MergeGuestCart(ctx context.Context, input *dto.MergeGuestCartRequest) (*dto.CartResponse, error) {
	// Extract user from context
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Convert DTOs to domain GuestCartItems
	guestItems := make([]ports.GuestCartItem, len(input.Body.Items))
	for i, item := range input.Body.Items {
		variantID, err := uuid.Parse(item.ProductVariantID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid product_variant_id UUID format", err)
		}

		guestItems[i] = ports.GuestCartItem{
			ProductVariantID: variantID,
			Quantity:         item.Quantity,
		}
	}

	// Merge guest cart
	cartDetail, err := h.service.MergeGuestCart(ctx, user.ID, guestItems)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		// Check for not found errors
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("Product variant not found")
		}
		// Check for stock validation errors
		if isStockError(err) {
			return nil, huma.Error400BadRequest("Stock validation failed", err)
		}
		return nil, huma.Error500InternalServerError("Failed to merge guest cart", err)
	}

	// Build response
	resp := &dto.CartResponse{}
	resp.Body.ID = cartDetail.ID.String()
	resp.Body.UserID = cartDetail.UserID.String()
	resp.Body.Items = convertCartItemsToDTO(cartDetail.Items)
	resp.Body.Subtotal = cartDetail.Subtotal
	resp.Body.ItemCount = cartDetail.ItemCount

	return resp, nil
}

// Helper functions

// convertCartItemsToDTO converts cart item details to DTOs
func convertCartItemsToDTO(items []*ports.CartItemDetail) []*dto.CartItemResponse {
	result := make([]*dto.CartItemResponse, len(items))
	for i, item := range items {
		result[i] = convertCartItemDetailToDTO(item)
	}
	return result
}

// convertCartItemDetailToDTO converts a single cart item detail to DTO
func convertCartItemDetailToDTO(item *ports.CartItemDetail) *dto.CartItemResponse {
	// Calculate final price for variant
	finalPrice := item.ProductVariant.CalculateFinalPrice(item.Product.BasePrice)

	return &dto.CartItemResponse{
		ID:            item.ID.String(),
		Quantity:      item.Quantity,
		PriceSnapshot: item.PriceSnapshot,
		Subtotal:      item.ItemSubtotal,
		ProductVariant: &dto.CartProductVariantDetail{
			ID:              item.ProductVariant.ID.String(),
			ProductID:       item.ProductVariant.ProductID.String(),
			SKU:             item.ProductVariant.SKU,
			Attributes:      item.ProductVariant.Attributes,
			StockQuantity:   item.ProductVariant.StockQuantity,
			PriceAdjustment: item.ProductVariant.PriceAdjustment,
			FinalPrice:      finalPrice,
			IsActive:        item.ProductVariant.IsActive,
			IsInStock:       item.ProductVariant.IsInStock(),
			CreatedAt:       item.ProductVariant.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       item.ProductVariant.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		Product: &dto.ProductSummaryResponse{
			ID:        item.Product.ID.String(),
			Name:      item.Product.Name,
			Slug:      item.Product.Slug,
			ImageURLs: item.Product.ImageURLs,
	},
	}
}

// isStockError checks if the error is a stock validation error
func isStockError(err error) bool {
	// Check for common stock error messages
	errMsg := err.Error()
	return errMsg == "insufficient stock" ||
		errMsg == "variant is not active" ||
		errMsg == "variant not found" ||
		errors.Is(err, domainErrors.ErrInsufficientStock) ||
		errors.Is(err, domainErrors.ErrVariantInactive)
}
