package handlers

import (
	"context"
	"errors"
	"testing"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductService is a mock implementation of ProductService
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *entities.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) GetProduct(ctx context.Context, id int) (*entities.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductService) GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductService) GetProductBySlug(ctx context.Context, slug string) (*entities.Product, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductService) ListProducts(ctx context.Context) ([]*entities.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Product), args.Error(1)
}

func (m *MockProductService) ListPublishedProducts(ctx context.Context) ([]*entities.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Product), args.Error(1)
}

func (m *MockProductService) ListProductsByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, product *entities.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) PublishProduct(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) ArchiveProduct(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.QueryResult), args.Error(1)
}

// TestCreateProduct_Success tests successful product creation with all required fields
func TestCreateProduct_Success(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "TEST-001"
	input.Body.Name = "Test Product"
	input.Body.Price = 99.99
	input.Body.Description = "A test product"

	mockService.On("CreateProduct", ctx, mock.MatchedBy(func(p *entities.Product) bool {
		// Verify DTO to entity mapping
		return p.SKU == "TEST-001" &&
			p.Name == "Test Product" &&
			p.Price == 99.99 &&
			p.Description == "A test product"
	})).Run(func(args mock.Arguments) {
		// Simulate service setting ID and timestamps
		product := args.Get(1).(*entities.Product)
		product.ID = 1
		product.Slug = "test-product"
		product.Status = entities.ProductStatusDraft
	}).Return(nil)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, response.Body.ID)
	assert.Equal(t, "TEST-001", response.Body.SKU)
	assert.Equal(t, "Test Product", response.Body.Name)
	assert.Equal(t, 99.99, response.Body.Price)
	assert.Equal(t, "A test product", response.Body.Description)
	assert.Equal(t, "test-product", response.Body.Slug)
	assert.Equal(t, "draft", response.Body.Status)
	mockService.AssertExpectations(t)
}

// TestCreateProduct_WithOptionalFields tests product creation with all optional fields
func TestCreateProduct_WithOptionalFields(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	customSlug := "custom-product-slug"
	weight := 500
	length := 20
	width := 15
	height := 10
	status := "published"

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "TEST-002"
	input.Body.Name = "Test Product with Options"
	input.Body.Price = 199.99
	input.Body.Description = "A test product with all fields"
	input.Body.Slug = &customSlug
	input.Body.Weight = &weight
	input.Body.Length = &length
	input.Body.Width = &width
	input.Body.Height = &height
	input.Body.Status = &status

	mockService.On("CreateProduct", ctx, mock.MatchedBy(func(p *entities.Product) bool {
		return p.SKU == "TEST-002" &&
			p.Slug == "custom-product-slug" &&
			p.Weight == 500 &&
			p.Length == 20 &&
			p.Width == 15 &&
			p.Height == 10 &&
			p.Status == entities.ProductStatusPublished
	})).Run(func(args mock.Arguments) {
		product := args.Get(1).(*entities.Product)
		product.ID = 2
	}).Return(nil)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 2, response.Body.ID)
	assert.Equal(t, "custom-product-slug", response.Body.Slug)
	assert.Equal(t, 500, response.Body.Weight)
	assert.Equal(t, 20, response.Body.Length)
	assert.Equal(t, 15, response.Body.Width)
	assert.Equal(t, 10, response.Body.Height)
	assert.Equal(t, "published", response.Body.Status)
	mockService.AssertExpectations(t)
}

// TestCreateProduct_ValidationError tests handling of validation errors from service
func TestCreateProduct_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "" // Empty SKU should cause validation error
	input.Body.Name = "Test Product"
	input.Body.Price = 99.99

	validationErr := domainErrors.NewValidationError("sku", "SKU is required")
	mockService.On("CreateProduct", ctx, mock.Anything).Return(validationErr)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Check that error is a Huma error with correct status code
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr), "Error should be a Huma status error")
	assert.Equal(t, 400, humaErr.GetStatus(), "Should return 400 Bad Request")
	mockService.AssertExpectations(t)
}

// TestCreateProduct_DuplicateError tests handling of duplicate entry errors
func TestCreateProduct_DuplicateError(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "DUPLICATE-SKU"
	input.Body.Name = "Test Product"
	input.Body.Price = 99.99

	duplicateErr := domainErrors.NewDuplicateError("Product", "sku", "DUPLICATE-SKU")
	mockService.On("CreateProduct", ctx, mock.Anything).Return(duplicateErr)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Check that error is a Huma 409 Conflict error
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr), "Error should be a Huma status error")
	assert.Equal(t, 409, humaErr.GetStatus(), "Should return 409 Conflict")
	assert.Contains(t, humaErr.Error(), "Product with this SKU or slug already exists")
	mockService.AssertExpectations(t)
}

// TestCreateProduct_InternalServerError tests handling of generic service errors
func TestCreateProduct_InternalServerError(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "TEST-003"
	input.Body.Name = "Test Product"
	input.Body.Price = 99.99

	// Generic error that's not a domain error
	genericErr := errors.New("database connection failed")
	mockService.On("CreateProduct", ctx, mock.Anything).Return(genericErr)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	// Check that error is a Huma 500 error
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr), "Error should be a Huma status error")
	assert.Equal(t, 500, humaErr.GetStatus(), "Should return 500 Internal Server Error")
	assert.Contains(t, humaErr.Error(), "Failed to create product")
	mockService.AssertExpectations(t)
}

// TestCreateProduct_NilOptionalFields tests that nil optional fields are handled correctly
func TestCreateProduct_NilOptionalFields(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "TEST-004"
	input.Body.Name = "Test Product"
	input.Body.Price = 99.99
	// All optional fields are nil

	mockService.On("CreateProduct", ctx, mock.MatchedBy(func(p *entities.Product) bool {
		// Verify that optional fields are zero values
		return p.SKU == "TEST-004" &&
			p.Slug == "" && // Should be empty, service will auto-generate
			p.Weight == 0 &&
			p.Length == 0 &&
			p.Width == 0 &&
			p.Height == 0 &&
			p.Status == ""
	})).Run(func(args mock.Arguments) {
		product := args.Get(1).(*entities.Product)
		product.ID = 4
		product.Slug = "test-product"
		product.Status = entities.ProductStatusDraft
	}).Return(nil)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 4, response.Body.ID)
	assert.Equal(t, "test-product", response.Body.Slug)
	assert.Equal(t, "draft", response.Body.Status)
	mockService.AssertExpectations(t)
}

// TestCreateProduct_EntityMappingComplete tests that all fields are properly mapped
func TestCreateProduct_EntityMappingComplete(t *testing.T) {
	// Arrange
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)
	ctx := context.Background()

	slug := "complete-product"
	weight := 1000
	length := 30
	width := 20
	height := 15
	status := "draft"

	input := &dto.CreateProductRequest{}
	input.Body.SKU = "COMPLETE-001"
	input.Body.Name = "Complete Product"
	input.Body.Price = 299.99
	input.Body.Description = "A product with all fields populated"
	input.Body.Slug = &slug
	input.Body.Weight = &weight
	input.Body.Length = &length
	input.Body.Width = &width
	input.Body.Height = &height
	input.Body.Status = &status

	var capturedProduct *entities.Product
	mockService.On("CreateProduct", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedProduct = args.Get(1).(*entities.Product)
		capturedProduct.ID = 100
	}).Return(nil)

	// Act
	response, err := handler.CreateProduct(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify all fields were mapped correctly from DTO to entity
	assert.Equal(t, "COMPLETE-001", capturedProduct.SKU)
	assert.Equal(t, "Complete Product", capturedProduct.Name)
	assert.Equal(t, 299.99, capturedProduct.Price)
	assert.Equal(t, "A product with all fields populated", capturedProduct.Description)
	assert.Equal(t, "complete-product", capturedProduct.Slug)
	assert.Equal(t, 1000, capturedProduct.Weight)
	assert.Equal(t, 30, capturedProduct.Length)
	assert.Equal(t, 20, capturedProduct.Width)
	assert.Equal(t, 15, capturedProduct.Height)
	assert.Equal(t, entities.ProductStatusDraft, capturedProduct.Status)

	// Verify response mapping
	assert.Equal(t, 100, response.Body.ID)
	assert.Equal(t, "COMPLETE-001", response.Body.SKU)
	assert.Equal(t, "Complete Product", response.Body.Name)
	assert.Equal(t, 299.99, response.Body.Price)
	mockService.AssertExpectations(t)
}
