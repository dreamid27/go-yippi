package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductVariantService is a mock implementation of ProductVariantService
type MockProductVariantService struct {
	mock.Mock
}

func (m *MockProductVariantService) CreateVariant(ctx context.Context, productID uuid.UUID, sku string, attributes map[string]string, stockQuantity int, priceAdjustment float64, isActive *bool) (*entities.ProductVariant, error) {
	args := m.Called(ctx, productID, sku, attributes, stockQuantity, priceAdjustment, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantService) UpdateVariant(ctx context.Context, variantID uuid.UUID, sku *string, attributes *map[string]string, stockQuantity *int, priceAdjustment *float64, isActive *bool) (*entities.ProductVariant, error) {
	args := m.Called(ctx, variantID, sku, attributes, stockQuantity, priceAdjustment, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantService) GetVariantsByProduct(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantService) UpdateStock(ctx context.Context, variantID uuid.UUID, newStockQuantity int) (*entities.ProductVariant, error) {
	args := m.Called(ctx, variantID, newStockQuantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantService) DeleteVariant(ctx context.Context, variantID uuid.UUID) error {
	args := m.Called(ctx, variantID)
	return args.Error(0)
}

func (m *MockProductVariantService) GetVariantByID(ctx context.Context, variantID uuid.UUID) (*entities.ProductVariant, error) {
	args := m.Called(ctx, variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

// ProductService mock for variant handler tests (reuse existing MockProductService)
// Already defined in product_handler_test.go

// Helper function to create test variant
func createTestVariant(id, productID uuid.UUID, sku string, stock int, adjustment float64, active bool) *entities.ProductVariant {
	return &entities.ProductVariant{
		ID:              id,
		ProductID:       productID,
		SKU:             sku,
		Attributes:      map[string]string{"size": "M", "color": "Red"},
		StockQuantity:   stock,
		PriceAdjustment: adjustment,
		IsActive:        active,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// TestListVariants_Success tests successful listing of product variants
func TestListVariants_Success(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variant1ID := uuid.New()
	variant2ID := uuid.New()

	product := &entities.Product{
		ID:        1,
		ProductID: productID,
		BasePrice: 100.00,
	}

	variants := []*entities.ProductVariant{
		createTestVariant(variant1ID, productID, "TSHIRT-M-RED", 50, 0, true),
		createTestVariant(variant2ID, productID, "TSHIRT-L-BLUE", 30, 10.00, true),
	}

	input := &dto.GetVariantsRequest{ProductID: productID.String()}

	mockProductService.On("GetProduct", ctx, productID).Return(product, nil)
	mockVariantService.On("GetVariantsByProduct", ctx, productID).Return(variants, nil)

	// Act
	response, err := handler.ListVariants(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Body.Data, 2)

	// Verify first variant mapping
	assert.Equal(t, variant1ID.String(), response.Body.Data[0].ID)
	assert.Equal(t, "TSHIRT-M-RED", response.Body.Data[0].SKU)
	assert.Equal(t, 50, response.Body.Data[0].StockQuantity)
	assert.Equal(t, 0, response.Body.Data[0].PriceAdjustment)
	assert.Equal(t, 100.00, response.Body.Data[0].FinalPrice)
	assert.True(t, response.Body.Data[0].IsInStock)

	// Verify second variant mapping
	assert.Equal(t, variant2ID.String(), response.Body.Data[1].ID)
	assert.Equal(t, "TSHIRT-L-BLUE", response.Body.Data[1].SKU)
	assert.Equal(t, 30, response.Body.Data[1].StockQuantity)
	assert.Equal(t, 10.00, response.Body.Data[1].PriceAdjustment)
	assert.Equal(t, 110.00, response.Body.Data[1].FinalPrice)

	mockProductService.AssertExpectations(t)
	mockVariantService.AssertExpectations(t)
}

// TestListVariants_InvalidUUID tests handling of invalid UUID
func TestListVariants_InvalidUUID(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	input := &dto.GetVariantsRequest{ProductID: "invalid-uuid"}

	// Act
	response, err := handler.ListVariants(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid product_id UUID format")

	mockVariantService.AssertNotCalled(t, "GetVariantsByProduct")
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestListVariants_ProductNotFound tests handling when product doesn't exist
func TestListVariants_ProductNotFound(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	input := &dto.GetVariantsRequest{ProductID: productID.String()}

	mockProductService.On("GetProduct", ctx, productID).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.ListVariants(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Product not found")

	mockProductService.AssertExpectations(t)
	mockVariantService.AssertNotCalled(t, "GetVariantsByProduct")
}

// TestCreateVariant_Success tests successful variant creation
func TestCreateVariant_Success(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = productID.String()
	input.Body.SKU = "TSHIRT-M-RED"
	input.Body.Attributes = map[string]string{"size": "M", "color": "Red"}
	input.Body.StockQuantity = 50
	input.Body.PriceAdjustment = 0
	isActive := true
	input.Body.IsActive = &isActive

	product := &entities.Product{
		ID:        1,
		ProductID: productID,
		BasePrice: 100.00,
	}

	variant := createTestVariant(variantID, productID, "TSHIRT-M-RED", 50, 0, true)

	mockVariantService.On("CreateVariant", ctx, productID, "TSHIRT-M-RED",
		mock.MatchedBy(func(attrs map[string]string) bool {
			return attrs["size"] == "M" && attrs["color"] == "Red"
		}), 50, 0.0, mock.Anything).Return(variant, nil)

	mockProductService.On("GetProduct", ctx, productID).Return(product, nil)

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, variantID.String(), response.Body.ID)
	assert.Equal(t, "TSHIRT-M-RED", response.Body.SKU)
	assert.Equal(t, 50, response.Body.StockQuantity)
	assert.Equal(t, 100.00, response.Body.FinalPrice)
	assert.True(t, response.Body.IsActive)
	assert.True(t, response.Body.IsInStock)

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertExpectations(t)
}

// TestCreateVariant_EmptyAttributes tests validation error when attributes are empty
func TestCreateVariant_EmptyAttributes(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = productID.String()
	input.Body.SKU = "TSHIRT-M"
	input.Body.Attributes = map[string]string{} // Empty attributes
	input.Body.StockQuantity = 50
	input.Body.PriceAdjustment = 0

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Attributes are required and cannot be empty")

	mockVariantService.AssertNotCalled(t, "CreateVariant")
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestCreateVariant_InvalidUUID tests handling of invalid product UUID
func TestCreateVariant_InvalidUUID(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = "invalid-uuid"
	input.Body.SKU = "TSHIRT-M"
	input.Body.Attributes = map[string]string{"size": "M"}
	input.Body.StockQuantity = 50
	input.Body.PriceAdjustment = 0

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid product_id UUID format")

	mockVariantService.AssertNotCalled(t, "CreateVariant")
}

// TestCreateVariant_ProductNotFound tests handling when product doesn't exist
func TestCreateVariant_ProductNotFound(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = productID.String()
	input.Body.SKU = "TSHIRT-M"
	input.Body.Attributes = map[string]string{"size": "M"}
	input.Body.StockQuantity = 50
	input.Body.PriceAdjustment = 0

	mockVariantService.On("CreateVariant", ctx, productID, "TSHIRT-M",
		mock.Anything, 50, 0.0, mock.Anything).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Product not found")

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestCreateVariant_DuplicateSKU tests handling of duplicate SKU error
func TestCreateVariant_DuplicateSKU(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = productID.String()
	input.Body.SKU = "DUPLICATE-SKU"
	input.Body.Attributes = map[string]string{"size": "M"}
	input.Body.StockQuantity = 50
	input.Body.PriceAdjustment = 0

	mockVariantService.On("CreateVariant", ctx, productID, "DUPLICATE-SKU",
		mock.Anything, 50, 0.0, mock.Anything).Return(nil, domainErrors.ErrDuplicateEntry)

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 409, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "SKU already exists")

	mockVariantService.AssertExpectations(t)
}

// TestCreateVariant_InvalidInput tests handling of invalid input errors
func TestCreateVariant_InvalidInput(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()

	input := &dto.CreateProductVariantRequest{}
	input.ProductID = productID.String()
	input.Body.SKU = "TSHIRT-M"
	input.Body.Attributes = map[string]string{"size": "M"}
	input.Body.StockQuantity = -10 // Negative stock
	input.Body.PriceAdjustment = 0

	mockVariantService.On("CreateVariant", ctx, productID, "TSHIRT-M",
		mock.Anything, -10, 0.0, mock.Anything).Return(nil, domainErrors.ErrInvalidInput)

	// Act
	response, err := handler.CreateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid input")

	mockVariantService.AssertExpectations(t)
}

// TestUpdateVariant_Success tests successful variant update
func TestUpdateVariant_Success(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	newSKU := "UPDATED-SKU"
	newStock := 75
	newAdjustment := 5.00
	newActive := false

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.SKU = &newSKU
	input.Body.StockQuantity = &newStock
	input.Body.PriceAdjustment = &newAdjustment
	input.Body.IsActive = &newActive

	product := &entities.Product{
		ID:        1,
		ProductID: productID,
		BasePrice: 100.00,
	}

	variant := createTestVariant(variantID, productID, "UPDATED-SKU", 75, 5.00, false)

	mockVariantService.On("UpdateVariant", ctx, variantID, &newSKU,
		mock.Anything, &newStock, &newAdjustment, &newActive).Return(variant, nil)

	mockProductService.On("GetProduct", ctx, productID).Return(product, nil)

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, variantID.String(), response.Body.ID)
	assert.Equal(t, "UPDATED-SKU", response.Body.SKU)
	assert.Equal(t, 75, response.Body.StockQuantity)
	assert.Equal(t, 5.00, response.Body.PriceAdjustment)
	assert.Equal(t, 105.00, response.Body.FinalPrice)
	assert.False(t, response.Body.IsActive)

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertExpectations(t)
}

// TestUpdateVariant_EmptyAttributesProvided tests validation when empty attributes are provided
func TestUpdateVariant_EmptyAttributesProvided(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	emptyAttrs := map[string]string{}

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.Attributes = &emptyAttrs // Empty attributes provided

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Attributes cannot be empty if provided")

	mockVariantService.AssertNotCalled(t, "UpdateVariant")
}

// TestUpdateVariant_InvalidProductUUID tests handling of invalid product UUID
func TestUpdateVariant_InvalidProductUUID(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	variantID := uuid.New()

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = "invalid-uuid"
	input.VariantID = variantID.String()

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid product_id UUID format")

	mockVariantService.AssertNotCalled(t, "UpdateVariant")
}

// TestUpdateVariant_InvalidVariantUUID tests handling of invalid variant UUID
func TestUpdateVariant_InvalidVariantUUID(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = productID.String()
	input.VariantID = "invalid-uuid"

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid variant_id UUID format")

	mockVariantService.AssertNotCalled(t, "UpdateVariant")
}

// TestUpdateVariant_VariantNotFound tests handling when variant doesn't exist
func TestUpdateVariant_VariantNotFound(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()

	mockVariantService.On("UpdateVariant", ctx, variantID,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Variant not found")

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestUpdateVariant_DuplicateSKU tests handling of duplicate SKU on update
func TestUpdateVariant_DuplicateSKU(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()
	duplicateSKU := "DUPLICATE-SKU"

	input := &dto.UpdateProductVariantRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.SKU = &duplicateSKU

	mockVariantService.On("UpdateVariant", ctx, variantID, &duplicateSKU,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainErrors.ErrDuplicateEntry)

	// Act
	response, err := handler.UpdateVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 409, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "SKU already exists")

	mockVariantService.AssertExpectations(t)
}

// TestUpdateStock_Success tests successful stock update
func TestUpdateStock_Success(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	newStock := 100

	input := &dto.UpdateStockRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.StockQuantity = newStock

	product := &entities.Product{
		ID:        1,
		ProductID: productID,
		BasePrice: 100.00,
	}

	variant := createTestVariant(variantID, productID, "TSHIRT-M-RED", 100, 0, true)

	mockVariantService.On("UpdateStock", ctx, variantID, newStock).Return(variant, nil)
	mockProductService.On("GetProduct", ctx, productID).Return(product, nil)

	// Act
	response, err := handler.UpdateStock(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, variantID.String(), response.Body.ID)
	assert.Equal(t, 100, response.Body.StockQuantity)
	assert.Equal(t, 100.00, response.Body.FinalPrice)
	assert.True(t, response.Body.IsInStock)

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertExpectations(t)
}

// TestUpdateStock_NegativeStock tests handling of negative stock quantity
func TestUpdateStock_NegativeStock(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	input := &dto.UpdateStockRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.StockQuantity = -10 // Negative stock

	mockVariantService.On("UpdateStock", ctx, variantID, -10).Return(nil, domainErrors.ErrInvalidInput)

	// Act
	response, err := handler.UpdateStock(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid stock quantity")

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestUpdateStock_VariantNotFound tests handling when variant doesn't exist
func TestUpdateStock_VariantNotFound(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	productID := uuid.New()
	variantID := uuid.New()

	input := &dto.UpdateStockRequest{}
	input.ProductID = productID.String()
	input.VariantID = variantID.String()
	input.Body.StockQuantity = 50

	mockVariantService.On("UpdateStock", ctx, variantID, 50).Return(nil, domainErrors.ErrNotFound)

	// Act
	response, err := handler.UpdateStock(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Variant not found")

	mockVariantService.AssertExpectations(t)
	mockProductService.AssertNotCalled(t, "GetProduct")
}

// TestDeleteVariant_Success tests successful variant deletion
func TestDeleteVariant_Success(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	variantID := uuid.New()

	input := &dto.DeleteVariantRequest{}
	input.ProductID = uuid.New().String()
	input.VariantID = variantID.String()

	mockVariantService.On("DeleteVariant", ctx, variantID).Return(nil)

	// Act
	response, err := handler.DeleteVariant(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)

	mockVariantService.AssertExpectations(t)
}

// TestDeleteVariant_InvalidUUID tests handling of invalid variant UUID
func TestDeleteVariant_InvalidUUID(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	input := &dto.DeleteVariantRequest{}
	input.ProductID = uuid.New().String()
	input.VariantID = "invalid-uuid"

	// Act
	response, err := handler.DeleteVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Invalid variant_id UUID format")

	mockVariantService.AssertNotCalled(t, "DeleteVariant")
}

// TestDeleteVariant_VariantNotFound tests handling when variant doesn't exist
func TestDeleteVariant_VariantNotFound(t *testing.T) {
	// Arrange
	mockVariantService := new(MockProductVariantService)
	mockProductService := new(MockProductService)
	handler := NewProductVariantHandler(mockVariantService, mockProductService)
	ctx := context.Background()

	variantID := uuid.New()

	input := &dto.DeleteVariantRequest{}
	input.ProductID = uuid.New().String()
	input.VariantID = variantID.String()

	mockVariantService.On("DeleteVariant", ctx, variantID).Return(domainErrors.ErrNotFound)

	// Act
	response, err := handler.DeleteVariant(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	assert.Contains(t, humaErr.Error(), "Variant not found")

	mockVariantService.AssertExpectations(t)
}
