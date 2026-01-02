package services

import (
	"context"
	"errors"
	"testing"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/google/uuid"
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

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	args := m.Called(ctx, id)
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

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
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

// MockPriceCalculatorService is a mock implementation of ports.PriceCalculatorService
type MockPriceCalculatorService struct {
	mock.Mock
}

func (m *MockPriceCalculatorService) CalculateVariantPrice(ctx context.Context, basePrice float64, adjustment float64) float64 {
	args := m.Called(ctx, basePrice, adjustment)
	return args.Get(0).(float64)
}

func (m *MockPriceCalculatorService) GetPriceRange(ctx context.Context, variants []*entities.ProductVariant, basePrice float64) (min float64, max float64) {
	args := m.Called(ctx, variants, basePrice)
	return args.Get(0).(float64), args.Get(1).(float64)
}

// TestGetProductFilters_Success tests successful retrieval of product filter attributes
func TestGetProductFilters_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockVariantRepo := new(MockProductVariantRepository)
	mockPriceCalc := new(MockPriceCalculatorService)
	service := NewProductService(mockRepo, mockCategoryRepo, mockVariantRepo, mockPriceCalc)
	ctx := context.Background()

	// Create test products
	productID1 := uuid.New()
	productID2 := uuid.New()
	productID3 := uuid.New()

	products := []*entities.Product{
		{ID: productID1, Name: "Product 1", Status: entities.ProductStatusPublished},
		{ID: productID2, Name: "Product 2", Status: entities.ProductStatusPublished},
		{ID: productID3, Name: "Product 3", Status: entities.ProductStatusPublished},
	}

	// Create test variants with attributes
	variants1 := []*entities.ProductVariant{
		{ID: uuid.New(), ProductID: productID1, Attributes: map[string]string{"size": "S", "color": "White"}},
		{ID: uuid.New(), ProductID: productID1, Attributes: map[string]string{"size": "M", "color": "Black"}},
	}
	variants2 := []*entities.ProductVariant{
		{ID: uuid.New(), ProductID: productID2, Attributes: map[string]string{"size": "S", "color": "White"}},
		{ID: uuid.New(), ProductID: productID2, Attributes: map[string]string{"size": "XS", "color": "Black"}},
	}
	variants3 := []*entities.ProductVariant{
		{ID: uuid.New(), ProductID: productID3, Attributes: map[string]string{"size": "L", "color": "Red"}},
	}

	queryResult := &entities.QueryResult{
		Products: products,
		PageInfo: entities.PageInfo{HasNextPage: false},
	}

	// Setup mock expectations
	mockCategoryRepo.On("GetDescendantIDs", ctx, mock.Anything).Return([]uuid.UUID{}, nil)
	mockRepo.On("Query", ctx, mock.Anything).Return(queryResult, nil)
	mockVariantRepo.On("FindByProductID", ctx, productID1).Return(variants1, nil)
	mockVariantRepo.On("FindByProductID", ctx, productID2).Return(variants2, nil)
	mockVariantRepo.On("FindByProductID", ctx, productID3).Return(variants3, nil)

	// Act
	params := ports.SearchProductsParams{}
	filters, err := service.GetProductFilters(ctx, params)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, filters)

	// Verify distinct attribute values
	assert.Contains(t, filters, "size")
	assert.Contains(t, filters, "color")

	// Check that values are sorted and distinct
	assert.Equal(t, []string{"L", "M", "S", "XS"}, filters["size"])
	assert.Equal(t, []string{"Black", "Red", "White"}, filters["color"])

	mockRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
}

// TestGetProductFilters_WithCategoryFilter tests filtering by category
func TestGetProductFilters_WithCategoryFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockVariantRepo := new(MockProductVariantRepository)
	mockPriceCalc := new(MockPriceCalculatorService)
	service := NewProductService(mockRepo, mockCategoryRepo, mockVariantRepo, mockPriceCalc)
	ctx := context.Background()

	categoryID := uuid.New()

	products := []*entities.Product{
		{ID: uuid.New(), Name: "Product 1", Status: entities.ProductStatusPublished},
	}

	variants := []*entities.ProductVariant{
		{ID: uuid.New(), ProductID: products[0].ID, Attributes: map[string]string{"size": "M", "material": "Cotton"}},
	}

	queryResult := &entities.QueryResult{
		Products: products,
		PageInfo: entities.PageInfo{HasNextPage: false},
	}

	mockCategoryRepo.On("GetDescendantIDs", ctx, []uuid.UUID{categoryID}).Return([]uuid.UUID{categoryID}, nil)
	mockRepo.On("Query", ctx, mock.Anything).Return(queryResult, nil)
	mockVariantRepo.On("FindByProductID", ctx, products[0].ID).Return(variants, nil)

	// Act
	params := ports.SearchProductsParams{CategoryID: &categoryID}
	filters, err := service.GetProductFilters(ctx, params)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, filters)
	assert.Equal(t, []string{"M"}, filters["size"])
	assert.Equal(t, []string{"Cotton"}, filters["material"])

	mockRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
	mockCategoryRepo.AssertExpectations(t)
}

// TestGetProductFilters_WithAttributePreFilter tests pre-filtering by size
func TestGetProductFilters_WithAttributePreFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockVariantRepo := new(MockProductVariantRepository)
	mockPriceCalc := new(MockPriceCalculatorService)
	service := NewProductService(mockRepo, mockCategoryRepo, mockVariantRepo, mockPriceCalc)
	ctx := context.Background()

	productID := uuid.New()
	products := []*entities.Product{
		{ID: productID, Name: "Product 1", Status: entities.ProductStatusPublished},
	}

	// Variants with different sizes and colors
	variants := []*entities.ProductVariant{
		{ID: uuid.New(), ProductID: productID, Attributes: map[string]string{"size": "S", "color": "White"}},
		{ID: uuid.New(), ProductID: productID, Attributes: map[string]string{"size": "S", "color": "Black"}},
		{ID: uuid.New(), ProductID: productID, Attributes: map[string]string{"size": "M", "color": "Red"}},
	}

	queryResult := &entities.QueryResult{
		Products: products,
		PageInfo: entities.PageInfo{HasNextPage: false},
	}

	size := "S"
	mockCategoryRepo.On("GetDescendantIDs", ctx, mock.Anything).Return([]uuid.UUID{}, nil)
	mockRepo.On("Query", ctx, mock.Anything).Return(queryResult, nil)
	mockVariantRepo.On("FindByProductID", ctx, productID).Return(variants, nil)

	// Act - pre-filter by size="S"
	params := ports.SearchProductsParams{Size: &size}
	filters, err := service.GetProductFilters(ctx, params)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, filters)

	// Should only get colors from size="S" variants
	assert.Equal(t, []string{"Black", "White"}, filters["color"])
	// Size should not be in filters since it's used for pre-filtering
	assert.NotContains(t, filters, "size")

	mockRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
}
