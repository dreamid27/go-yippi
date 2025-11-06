package services

import (
	"context"
	"errors"
	"testing"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductRepository is a mock implementation of ports.ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *entities.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id int) (*entities.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySlug(ctx context.Context, slug string) (*entities.Product, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *entities.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) Query(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.QueryResult), args.Error(1)
}

func (m *MockProductRepository) List(ctx context.Context) ([]*entities.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Product), args.Error(1)
}

func (m *MockProductRepository) ListByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Product), args.Error(1)
}

// TestCreateProduct_Success tests successful product creation with all required fields
func TestCreateProduct_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:         "TEST-001",
		Name:        "Test Product",
		Price:       99.99,
		Description: "A test product",
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(p *entities.Product) bool {
		// Verify that the service sets default status and slug
		return p.SKU == "TEST-001" &&
			p.Name == "Test Product" &&
			p.Price == 99.99 &&
			p.Status == entities.ProductStatusDraft &&
			p.Slug == "test-product"
	})).Return(nil)

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entities.ProductStatusDraft, product.Status, "Status should default to draft")
	assert.Equal(t, "test-product", product.Slug, "Slug should be auto-generated")
	mockRepo.AssertExpectations(t)
}

// TestCreateProduct_WithCustomSlug tests product creation with a custom slug
func TestCreateProduct_WithCustomSlug(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-002",
		Name:  "Test Product",
		Slug:  "custom-slug",
		Price: 49.99,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(p *entities.Product) bool {
		return p.Slug == "custom-slug"
	})).Return(nil)

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "custom-slug", product.Slug, "Custom slug should be preserved")
	mockRepo.AssertExpectations(t)
}

// TestCreateProduct_WithOptionalFields tests product creation with optional dimension fields
func TestCreateProduct_WithOptionalFields(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:    "TEST-003",
		Name:   "Test Product with Dimensions",
		Price:  199.99,
		Weight: 500,  // grams
		Length: 20,   // cm
		Width:  15,   // cm
		Height: 10,   // cm
		Status: entities.ProductStatusPublished,
	}

	mockRepo.On("Create", ctx, product).Return(nil)

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 500, product.Weight)
	assert.Equal(t, 20, product.Length)
	assert.Equal(t, 15, product.Width)
	assert.Equal(t, 10, product.Height)
	assert.Equal(t, entities.ProductStatusPublished, product.Status)
	mockRepo.AssertExpectations(t)
}

// TestCreateProduct_EmptySKU tests validation error when SKU is empty
func TestCreateProduct_EmptySKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "",
		Name:  "Test Product",
		Price: 99.99,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput), "Should return validation error")
	assert.Contains(t, err.Error(), "SKU is required")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_WhitespaceSKU tests validation error when SKU contains only whitespace
func TestCreateProduct_WhitespaceSKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "   ",
		Name:  "Test Product",
		Price: 99.99,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "SKU is required")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_EmptyName tests validation error when name is empty
func TestCreateProduct_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-004",
		Name:  "",
		Price: 99.99,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_ZeroPrice tests validation error when price is zero
func TestCreateProduct_ZeroPrice(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-005",
		Name:  "Test Product",
		Price: 0,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Price must be greater than 0")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_NegativePrice tests validation error when price is negative
func TestCreateProduct_NegativePrice(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-006",
		Name:  "Test Product",
		Price: -10.00,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Price must be greater than 0")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_NegativeWeight tests validation error when weight is negative
func TestCreateProduct_NegativeWeight(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:    "TEST-007",
		Name:   "Test Product",
		Price:  99.99,
		Weight: -100,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Weight cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_NegativeLength tests validation error when length is negative
func TestCreateProduct_NegativeLength(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:    "TEST-008",
		Name:   "Test Product",
		Price:  99.99,
		Length: -10,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Length cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_NegativeWidth tests validation error when width is negative
func TestCreateProduct_NegativeWidth(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-009",
		Name:  "Test Product",
		Price: 99.99,
		Width: -5,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Width cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_NegativeHeight tests validation error when height is negative
func TestCreateProduct_NegativeHeight(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:    "TEST-010",
		Name:   "Test Product",
		Price:  99.99,
		Height: -8,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Height cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_InvalidStatus tests validation error when status is invalid
func TestCreateProduct_InvalidStatus(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:    "TEST-011",
		Name:   "Test Product",
		Price:  99.99,
		Status: entities.ProductStatus("invalid-status"),
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrInvalidInput))
	assert.Contains(t, err.Error(), "Invalid product status")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateProduct_RepositoryDuplicateError tests handling of duplicate entry from repository
func TestCreateProduct_RepositoryDuplicateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-DUPLICATE",
		Name:  "Test Product",
		Price: 99.99,
	}

	duplicateErr := domainErrors.NewDuplicateError("Product", "sku", "TEST-DUPLICATE")
	mockRepo.On("Create", ctx, mock.Anything).Return(duplicateErr)

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrDuplicateEntry))
	mockRepo.AssertExpectations(t)
}

// TestCreateProduct_RepositoryGenericError tests handling of generic repository error
func TestCreateProduct_RepositoryGenericError(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	service := NewProductService(mockRepo, mockCategoryRepo)
	ctx := context.Background()

	product := &entities.Product{
		SKU:   "TEST-012",
		Name:  "Test Product",
		Price: 99.99,
	}

	genericErr := errors.New("database connection failed")
	mockRepo.On("Create", ctx, mock.Anything).Return(genericErr)

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	require.Error(t, err)
	assert.Equal(t, "database connection failed", err.Error())
	mockRepo.AssertExpectations(t)
}
