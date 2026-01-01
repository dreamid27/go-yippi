package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCartService is a mock implementation of CartService
type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (*ports.CartDetail, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.CartDetail), args.Error(1)
}

func (m *MockCartService) AddItemToCart(ctx context.Context, userID uuid.UUID, variantID uuid.UUID, quantity int) (*entities.CartItem, error) {
	args := m.Called(ctx, userID, variantID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CartItem), args.Error(1)
}

func (m *MockCartService) UpdateCartItemQuantity(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID, newQuantity int) (*entities.CartItem, error) {
	args := m.Called(ctx, userID, cartItemID, newQuantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CartItem), args.Error(1)
}

func (m *MockCartService) RemoveCartItem(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID) error {
	args := m.Called(ctx, userID, cartItemID)
	return args.Error(0)
}

func (m *MockCartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartService) MergeGuestCart(ctx context.Context, userID uuid.UUID, guestItems []ports.GuestCartItem) (*ports.CartDetail, error) {
	args := m.Called(ctx, userID, guestItems)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.CartDetail), args.Error(1)
}

func (m *MockCartService) GetCartWithDetails(ctx context.Context, userID uuid.UUID) (*ports.CartDetail, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.CartDetail), args.Error(1)
}

// Helper functions to create test data
func createTestCartUser(id uuid.UUID) *entities.User {
	return &entities.User{
		ID:        id,
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      entities.UserRoleCustomer,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestCartProduct(id uuid.UUID, name string, basePrice float64) *entities.Product {
	return &entities.Product{
		ID:        id,
		Name:      name,
		BasePrice: basePrice,
		Slug:      "test-product",
		Status:    entities.ProductStatusPublished,
		ImageURLs: []string{"https://example.com/image.jpg"},
	}
}

func createTestCartVariant(productID, id uuid.UUID, sku string, attributes map[string]string, stock int, priceAdj float64, isActive bool) *entities.ProductVariant {
	return &entities.ProductVariant{
		ID:              id,
		ProductID:       productID,
		SKU:             sku,
		Attributes:      attributes,
		StockQuantity:   stock,
		PriceAdjustment: priceAdj,
		IsActive:        isActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func createTestCartItem(id, cartID, variantID uuid.UUID, quantity int, priceSnapshot float64) *entities.CartItem {
	return &entities.CartItem{
		ID:            id,
		CartID:        cartID,
		ProductVariantID: variantID,
		Quantity:      quantity,
		PriceSnapshot: priceSnapshot,
	}
}

func createTestCartDetail(id, userID uuid.UUID, items []*ports.CartItemDetail, subtotal float64, itemCount int) *ports.CartDetail {
	return &ports.CartDetail{
		ID:        id,
		UserID:    userID,
		Items:     items,
		Subtotal:  subtotal,
		ItemCount: itemCount,
	}
}

func createTestCartItemDetail(id, variantID, productID uuid.UUID, quantity int, priceSnapshot float64, variant *entities.ProductVariant, product *entities.Product) *ports.CartItemDetail {
	itemSubtotal := priceSnapshot * float64(quantity)
	return &ports.CartItemDetail{
		ID:             id,
		Quantity:       quantity,
		PriceSnapshot:  priceSnapshot,
		ItemSubtotal:   itemSubtotal,
		ProductVariant: variant,
		Product:        product,
	}
}

// Helper function to create context with user
func contextWithUser(ctx context.Context, user *entities.User) context.Context {
	return context.WithValue(ctx, "user", user)
}

// TestGetCart_Success tests successful retrieval of user's cart
func TestGetCart_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	product := createTestCartProduct(productID, "Test Product", 100.0)
	variant := createTestCartVariant(productID, variantID, "VAR-001", map[string]string{"size": "M"}, 50, 0, true)
	itemDetail := createTestCartItemDetail(itemID, variantID, productID, 2, 100.0, variant, product)
	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{itemDetail}, 200.0, 2)

	mockService.On("GetCartWithDetails", ctx, userID).Return(cartDetail, nil)

	// Act
	response, err := handler.GetCart(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, cartID.String(), response.Body.ID)
	assert.Equal(t, userID.String(), response.Body.UserID)
	assert.Len(t, response.Body.Items, 1)
	assert.Equal(t, 200.0, response.Body.Subtotal)
	assert.Equal(t, 2, response.Body.ItemCount)

	// Verify item details
	item := response.Body.Items[0]
	assert.Equal(t, itemID.String(), item.ID)
	assert.Equal(t, 2, item.Quantity)
	assert.Equal(t, 100.0, item.PriceSnapshot)
	assert.Equal(t, 200.0, item.Subtotal)
	assert.Equal(t, variantID.String(), item.ProductVariant.ID)
	assert.Equal(t, "VAR-001", item.ProductVariant.SKU)
	assert.Equal(t, productID.String(), item.Product.ID)
	assert.Equal(t, "Test Product", item.Product.Name)

	mockService.AssertExpectations(t)
}

// TestGetCart_EmptyCart tests successful retrieval of empty cart
func TestGetCart_EmptyCart(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{}, 0.0, 0)

	mockService.On("GetCartWithDetails", ctx, userID).Return(cartDetail, nil)

	// Act
	response, err := handler.GetCart(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, cartID.String(), response.Body.ID)
	assert.Len(t, response.Body.Items, 0)
	assert.Equal(t, 0.0, response.Body.Subtotal)
	assert.Equal(t, 0, response.Body.ItemCount)

	mockService.AssertExpectations(t)
}

// TestGetCart_Unauthorized tests handling of unauthenticated user
func TestGetCart_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	// Act
	response, err := handler.GetCart(ctx, &struct{}{})

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "User not authenticated")

	mockService.AssertNotCalled(t, "GetCartWithDetails")
}

// TestAddCartItem_Success tests successful addition of item to cart
func TestAddCartItem_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = variantID.String()
	input.Body.Quantity = 2

	cartItem := createTestCartItem(itemID, cartID, variantID, 2, 100.0)
	product := createTestCartProduct(productID, "Test Product", 100.0)
	variant := createTestCartVariant(productID, variantID, "VAR-001", map[string]string{"size": "M"}, 50, 0, true)
	itemDetail := createTestCartItemDetail(itemID, variantID, productID, 2, 100.0, variant, product)
	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{itemDetail}, 200.0, 2)

	mockService.On("AddItemToCart", ctx, userID, variantID, 2).Return(cartItem, nil)
	mockService.On("GetCartWithDetails", ctx, userID).Return(cartDetail, nil)

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, itemID.String(), response.Body.ID)
	assert.Equal(t, 2, response.Body.Quantity)
	assert.Equal(t, 100.0, response.Body.PriceSnapshot)
	assert.Equal(t, 200.0, response.Body.Subtotal)

	mockService.AssertExpectations(t)
}

// TestAddCartItem_UpdateExistingQuantity tests updating quantity when item already exists
func TestAddCartItem_UpdateExistingQuantity(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = variantID.String()
	input.Body.Quantity = 3 // Adding 3 more to existing quantity

	cartItem := createTestCartItem(itemID, cartID, variantID, 5, 100.0) // Updated to 5
	product := createTestCartProduct(productID, "Test Product", 100.0)
	variant := createTestCartVariant(productID, variantID, "VAR-001", map[string]string{"size": "M"}, 50, 0, true)
	itemDetail := createTestCartItemDetail(itemID, variantID, productID, 5, 100.0, variant, product)
	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{itemDetail}, 500.0, 5)

	mockService.On("AddItemToCart", ctx, userID, variantID, 3).Return(cartItem, nil)
	mockService.On("GetCartWithDetails", ctx, userID).Return(cartDetail, nil)

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, itemID.String(), response.Body.ID)
	assert.Equal(t, 5, response.Body.Quantity) // Quantity updated

	mockService.AssertExpectations(t)
}

// TestAddCartItem_InvalidUUID tests handling of invalid variant UUID
func TestAddCartItem_InvalidUUID(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = "invalid-uuid"
	input.Body.Quantity = 2

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid product_variant_id UUID format")

	mockService.AssertNotCalled(t, "AddItemToCart")
}

// TestAddCartItem_Unauthorized tests handling of unauthenticated user
func TestAddCartItem_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = uuid.New().String()
	input.Body.Quantity = 2

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())

	mockService.AssertNotCalled(t, "AddItemToCart")
}

// TestAddCartItem_VariantNotFound tests handling of non-existent variant
func TestAddCartItem_VariantNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = variantID.String()
	input.Body.Quantity = 2

	mockService.On("AddItemToCart", ctx, userID, variantID, 2).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Product variant not found")

	mockService.AssertExpectations(t)
}

// TestAddCartItem_InsufficientStock tests handling of insufficient stock
func TestAddCartItem_InsufficientStock(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = variantID.String()
	input.Body.Quantity = 100

	stockErr := domainErrors.NewInsufficientStockError(variantID, 5, 100)
	mockService.On("AddItemToCart", ctx, userID, variantID, 100).Return(nil, stockErr)

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Insufficient stock")

	mockService.AssertExpectations(t)
}

// TestAddCartItem_InvalidInput tests handling of validation errors
func TestAddCartItem_InvalidInput(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.AddCartItemRequest{}
	input.Body.ProductVariantID = variantID.String()
	input.Body.Quantity = -5 // Negative quantity

	mockService.On("AddItemToCart", ctx, userID, variantID, -5).Return(nil, domainErrors.ErrInvalidInput)

	// Act
	response, err := handler.AddCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())

	mockService.AssertExpectations(t)
}

// TestUpdateCartItem_Success tests successful cart item quantity update
func TestUpdateCartItem_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = itemID.String()
	input.Body.Quantity = 5

	cartItem := createTestCartItem(itemID, cartID, variantID, 5, 100.0)
	product := createTestCartProduct(productID, "Test Product", 100.0)
	variant := createTestCartVariant(productID, variantID, "VAR-001", map[string]string{"size": "M"}, 50, 0, true)
	itemDetail := createTestCartItemDetail(itemID, variantID, productID, 5, 100.0, variant, product)
	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{itemDetail}, 500.0, 5)

	mockService.On("UpdateCartItemQuantity", ctx, userID, itemID, 5).Return(cartItem, nil)
	mockService.On("GetCartWithDetails", ctx, userID).Return(cartDetail, nil)

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, itemID.String(), response.Body.ID)
	assert.Equal(t, 5, response.Body.Quantity)
	assert.Equal(t, 500.0, response.Body.Subtotal)

	mockService.AssertExpectations(t)
}

// TestUpdateCartItem_DeleteItem tests deleting item when quantity is 0
func TestUpdateCartItem_DeleteItem(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = itemID.String()
	input.Body.Quantity = 0 // Delete item

	cartItem := createTestCartItem(itemID, cartID, uuid.New(), 0, 100.0)

	mockService.On("UpdateCartItemQuantity", ctx, userID, itemID, 0).Return(cartItem, nil)

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, response) // Returns nil for 204 No Content

	mockService.AssertExpectations(t)
}

// TestUpdateCartItem_InvalidUUID tests handling of invalid item UUID
func TestUpdateCartItem_InvalidUUID(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = "invalid-uuid"
	input.Body.Quantity = 5

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid item_id UUID format")

	mockService.AssertNotCalled(t, "UpdateCartItemQuantity")
}

// TestUpdateCartItem_Unauthorized tests handling of unauthenticated user
func TestUpdateCartItem_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = uuid.New().String()
	input.Body.Quantity = 5

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())

	mockService.AssertNotCalled(t, "UpdateCartItemQuantity")
}

// TestUpdateCartItem_ItemNotFound tests handling of non-existent cart item
func TestUpdateCartItem_ItemNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = itemID.String()
	input.Body.Quantity = 5

	mockService.On("UpdateCartItemQuantity", ctx, userID, itemID, 5).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Cart item not found")

	mockService.AssertExpectations(t)
}

// TestUpdateCartItem_Forbidden tests handling when item doesn't belong to user
func TestUpdateCartItem_Forbidden(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = itemID.String()
	input.Body.Quantity = 5

	mockService.On("UpdateCartItemQuantity", ctx, userID, itemID, 5).Return(nil, domainErrors.ErrForbidden)

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 403, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "permission")

	mockService.AssertExpectations(t)
}

// TestUpdateCartItem_InsufficientStock tests handling of insufficient stock
func TestUpdateCartItem_InsufficientStock(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.UpdateCartItemRequest{}
	input.ItemID = itemID.String()
	input.Body.Quantity = 100

	stockErr := domainErrors.NewInsufficientStockError(variantID, 5, 100)
	mockService.On("UpdateCartItemQuantity", ctx, userID, itemID, 100).Return(nil, stockErr)

	// Act
	response, err := handler.UpdateCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Insufficient stock")

	mockService.AssertExpectations(t)
}

// TestRemoveCartItem_Success tests successful removal of cart item
func TestRemoveCartItem_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.RemoveCartItemRequest{}
	input.ItemID = itemID.String()

	mockService.On("RemoveCartItem", ctx, userID, itemID).Return(nil)

	// Act
	response, err := handler.RemoveCartItem(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)

	mockService.AssertExpectations(t)
}

// TestRemoveCartItem_InvalidUUID tests handling of invalid item UUID
func TestRemoveCartItem_InvalidUUID(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.RemoveCartItemRequest{}
	input.ItemID = "invalid-uuid"

	// Act
	response, err := handler.RemoveCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid item_id UUID format")

	mockService.AssertNotCalled(t, "RemoveCartItem")
}

// TestRemoveCartItem_Unauthorized tests handling of unauthenticated user
func TestRemoveCartItem_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	input := &dto.RemoveCartItemRequest{}
	input.ItemID = uuid.New().String()

	// Act
	response, err := handler.RemoveCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())

	mockService.AssertNotCalled(t, "RemoveCartItem")
}

// TestRemoveCartItem_ItemNotFound tests handling of non-existent cart item
func TestRemoveCartItem_ItemNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.RemoveCartItemRequest{}
	input.ItemID = itemID.String()

	mockService.On("RemoveCartItem", ctx, userID, itemID).Return(domainErrors.ErrNotFound)

	// Act
	response, err := handler.RemoveCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Cart item not found")

	mockService.AssertExpectations(t)
}

// TestRemoveCartItem_Forbidden tests handling when item doesn't belong to user
func TestRemoveCartItem_Forbidden(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.RemoveCartItemRequest{}
	input.ItemID = itemID.String()

	mockService.On("RemoveCartItem", ctx, userID, itemID).Return(domainErrors.ErrForbidden)

	// Act
	response, err := handler.RemoveCartItem(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 403, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "permission")

	mockService.AssertExpectations(t)
}

// TestClearCart_Success tests successful clearing of cart
func TestClearCart_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	mockService.On("ClearCart", ctx, userID).Return(nil)

	// Act
	response, err := handler.ClearCart(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)

	mockService.AssertExpectations(t)
}

// TestClearCart_Unauthorized tests handling of unauthenticated user
func TestClearCart_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	// Act
	response, err := handler.ClearCart(ctx, &struct{}{})

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())

	mockService.AssertNotCalled(t, "ClearCart")
}

// TestMergeGuestCart_Success tests successful merging of guest cart
func TestMergeGuestCart_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	cartID := uuid.New()
	productID := uuid.New()
	variantID1 := uuid.New()
	variantID2 := uuid.New()
	itemID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: variantID1.String(), Quantity: 2},
		{ProductVariantID: variantID2.String(), Quantity: 3},
	}

	product := createTestCartProduct(productID, "Test Product", 100.0)
	variant1 := createTestCartVariant(productID, variantID1, "VAR-001", map[string]string{"size": "M"}, 50, 0, true)
	_ = createTestCartVariant(productID, variantID2, "VAR-002", map[string]string{"size": "L"}, 30, 10.0, true) // Used for validation only

	itemDetail := createTestCartItemDetail(itemID, variantID1, productID, 2, 100.0, variant1, product)
	cartDetail := createTestCartDetail(cartID, userID, []*ports.CartItemDetail{itemDetail}, 200.0, 2)

	mockService.On("MergeGuestCart", ctx, userID, mock.MatchedBy(func(items []ports.GuestCartItem) bool {
		return len(items) == 2 &&
			items[0].ProductVariantID == variantID1 &&
			items[0].Quantity == 2 &&
			items[1].ProductVariantID == variantID2 &&
			items[1].Quantity == 3
	})).Return(cartDetail, nil)

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, cartID.String(), response.Body.ID)
	assert.Equal(t, userID.String(), response.Body.UserID)
	assert.Len(t, response.Body.Items, 1)

	mockService.AssertExpectations(t)
}

// TestMergeGuestCart_InvalidUUID tests handling of invalid variant UUID
func TestMergeGuestCart_InvalidUUID(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: "invalid-uuid", Quantity: 2},
	}

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid product_variant_id UUID format")

	mockService.AssertNotCalled(t, "MergeGuestCart")
}

// TestMergeGuestCart_Unauthorized tests handling of unauthenticated user
func TestMergeGuestCart_Unauthorized(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()
	// No user in context

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: uuid.New().String(), Quantity: 2},
	}

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 401, humaErr.GetStatus())

	mockService.AssertNotCalled(t, "MergeGuestCart")
}

// TestMergeGuestCart_VariantNotFound tests handling of non-existent variant
func TestMergeGuestCart_VariantNotFound(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: variantID.String(), Quantity: 2},
	}

	_ = []ports.GuestCartItem{
		{ProductVariantID: variantID, Quantity: 2},
	}

	mockService.On("MergeGuestCart", ctx, userID, mock.Anything).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Product variant not found")

	mockService.AssertExpectations(t)
}

// TestMergeGuestCart_InsufficientStock tests handling of insufficient stock
func TestMergeGuestCart_InsufficientStock(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: variantID.String(), Quantity: 100},
	}

	stockErr := domainErrors.NewInsufficientStockError(variantID, 5, 100)
	mockService.On("MergeGuestCart", ctx, userID, mock.Anything).Return(nil, stockErr)

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Stock validation failed")

	mockService.AssertExpectations(t)
}

// TestMergeGuestCart_InvalidInput tests handling of validation errors
func TestMergeGuestCart_InvalidInput(t *testing.T) {
	// Arrange
	mockService := new(MockCartService)
	handler := NewCartHandler(mockService)
	ctx := context.Background()

	userID := uuid.New()
	variantID := uuid.New()

	user := createTestCartUser(userID)
	ctx = contextWithUser(ctx, user)

	input := &dto.MergeGuestCartRequest{}
	input.Body.Items = []dto.GuestCartItemDTO{
		{ProductVariantID: variantID.String(), Quantity: -5}, // Negative quantity
	}

	mockService.On("MergeGuestCart", ctx, userID, mock.Anything).Return(nil, domainErrors.ErrInvalidInput)

	// Act
	response, err := handler.MergeGuestCart(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())

	mockService.AssertExpectations(t)
}
