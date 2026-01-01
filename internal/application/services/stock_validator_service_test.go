package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/infrastructure/logging"
)

func init() {
	// Initialize logging for tests
	logging.Initialize(logging.DefaultLogConfig())
}

// MockProductVariantRepository is a mock for ProductVariantRepository
type MockProductVariantRepository struct {
	mock.Mock
}

func (m *MockProductVariantRepository) Create(ctx context.Context, variant *entities.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) Update(ctx context.Context, variant *entities.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockProductVariantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestValidateStock_Success(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variantID := uuid.New()
	variant := &entities.ProductVariant{
		ID:            variantID,
		IsActive:      true,
		StockQuantity: 100,
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variantID).Return(variant, nil)

	err := service.ValidateStock(ctx, variantID, 50)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestValidateStock_VariantNotFound(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variantID := uuid.New()
	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variantID).Return(nil, domainErrors.ErrNotFound)

	err := service.ValidateStock(ctx, variantID, 50)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrNotFound))
	mockRepo.AssertExpectations(t)
}

func TestValidateStock_InsufficientStock(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variantID := uuid.New()
	variant := &entities.ProductVariant{
		ID:            variantID,
		IsActive:      true,
		StockQuantity: 10,
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variantID).Return(variant, nil)

	err := service.ValidateStock(ctx, variantID, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
	mockRepo.AssertExpectations(t)
}

func TestValidateStock_ExactStock(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variantID := uuid.New()
	variant := &entities.ProductVariant{
		ID:            variantID,
		IsActive:      true,
		StockQuantity: 50,
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variantID).Return(variant, nil)

	err := service.ValidateStock(ctx, variantID, 50)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestValidateStock_InactiveVariant(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variantID := uuid.New()
	variant := &entities.ProductVariant{
		ID:            variantID,
		IsActive:      false,
		StockQuantity: 100,
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variantID).Return(variant, nil)

	err := service.ValidateStock(ctx, variantID, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inactive")
	mockRepo.AssertExpectations(t)
}

func TestValidateCartStock_AllValid(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variant1ID := uuid.New()
	variant2ID := uuid.New()

	cartItems := []*entities.CartItem{
		{
			ID:               uuid.New(),
			ProductVariantID: variant1ID,
			Quantity:         10,
		},
		{
			ID:               uuid.New(),
			ProductVariantID: variant2ID,
			Quantity:         5,
		},
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant1ID).Return(&entities.ProductVariant{
		ID:            variant1ID,
		IsActive:      true,
		StockQuantity: 100,
	}, nil)

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant2ID).Return(&entities.ProductVariant{
		ID:            variant2ID,
		IsActive:      true,
		StockQuantity: 50,
	}, nil)

	validationErrors, err := service.ValidateCartStock(ctx, cartItems)
	assert.NoError(t, err)
	assert.Nil(t, validationErrors)
	mockRepo.AssertExpectations(t)
}

func TestValidateCartStock_MultipleErrors(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variant1ID := uuid.New()
	variant2ID := uuid.New()
	variant3ID := uuid.New()

	cartItems := []*entities.CartItem{
		{
			ID:               uuid.New(),
			ProductVariantID: variant1ID,
			Quantity:         10,
		},
		{
			ID:               uuid.New(),
			ProductVariantID: variant2ID,
			Quantity:         50,
		},
		{
			ID:               uuid.New(),
			ProductVariantID: variant3ID,
			Quantity:         5,
		},
	}

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant1ID).Return(&entities.ProductVariant{
		ID:            variant1ID,
		IsActive:      false,
		StockQuantity: 100,
	}, nil)

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant2ID).Return(&entities.ProductVariant{
		ID:            variant2ID,
		IsActive:      true,
		StockQuantity: 10,
	}, nil)

	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant3ID).Return(nil, domainErrors.ErrNotFound)

	validationErrors, err := service.ValidateCartStock(ctx, cartItems)
	assert.NoError(t, err)
	assert.NotNil(t, validationErrors)
	assert.Len(t, validationErrors, 3)
	mockRepo.AssertExpectations(t)
}

func TestValidateCartStock_EmptyCart(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	validationErrors, err := service.ValidateCartStock(ctx, []*entities.CartItem{})
	assert.NoError(t, err)
	assert.Nil(t, validationErrors)
}

func TestValidateStockForVariant_Success(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variant := &entities.ProductVariant{
		ID:            uuid.New(),
		IsActive:      true,
		StockQuantity: 100,
	}

	// Set up mock to return the variant
	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant.ID).Return(variant, nil)

	// This test validates the ValidateStock method with a valid variant ID
	err := service.ValidateStock(ctx, variant.ID, 50)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestValidateStockForVariant_Insufficient(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variant := &entities.ProductVariant{
		ID:            uuid.New(),
		IsActive:      true,
		StockQuantity: 10,
	}

	// Set up mock to return the variant
	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant.ID).Return(variant, nil)

	// This test validates the ValidateStock method with insufficient stock
	err := service.ValidateStock(ctx, variant.ID, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
	mockRepo.AssertExpectations(t)
}

func TestValidateStockForVariant_Inactive(t *testing.T) {
	mockRepo := new(MockProductVariantRepository)
	service := NewStockValidatorService(mockRepo)
	ctx := context.Background()

	variant := &entities.ProductVariant{
		ID:            uuid.New(),
		IsActive:      false,
		StockQuantity: 100,
	}

	// Set up mock to return the variant
	mockRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), variant.ID).Return(variant, nil)

	// This test validates the ValidateStock method with an inactive variant
	err := service.ValidateStock(ctx, variant.ID, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inactive")
	mockRepo.AssertExpectations(t)
}
