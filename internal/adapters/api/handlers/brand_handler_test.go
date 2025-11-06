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

// MockBrandService is a mock implementation of BrandService
type MockBrandService struct {
	mock.Mock
}

func (m *MockBrandService) CreateBrand(ctx context.Context, brand *entities.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandService) GetBrand(ctx context.Context, id uuid.UUID) (*entities.Brand, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Brand), args.Error(1)
}

func (m *MockBrandService) GetBrandByName(ctx context.Context, name string) (*entities.Brand, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Brand), args.Error(1)
}

func (m *MockBrandService) ListBrands(ctx context.Context) ([]*entities.Brand, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Brand), args.Error(1)
}

func (m *MockBrandService) UpdateBrand(ctx context.Context, brand *entities.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandService) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateBrand_Success tests successful brand creation
func TestCreateBrand_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateBrandRequest{}
	input.Body.Name = "Test Brand"

	mockService.On("CreateBrand", ctx, mock.MatchedBy(func(b *entities.Brand) bool {
		return b.Name == "Test Brand"
	})).Run(func(args mock.Arguments) {
		// Simulate service setting ID and timestamps
		brand := args.Get(1).(*entities.Brand)
		brand.ID = uuid.New()
		brand.CreatedAt = time.Now()
		brand.UpdatedAt = time.Now()
	}).Return(nil)

	// Act
	response, err := handler.CreateBrand(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Brand", response.Body.Name)
	assert.NotEqual(t, uuid.Nil, response.Body.ID)
	mockService.AssertExpectations(t)
}

// TestCreateBrand_ValidationError tests handling of validation errors
func TestCreateBrand_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateBrandRequest{}
	input.Body.Name = "" // Empty name

	validationErr := domainErrors.NewValidationError("name", "Name is required")
	mockService.On("CreateBrand", ctx, mock.Anything).Return(validationErr)

	// Act
	response, err := handler.CreateBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestCreateBrand_DuplicateError tests handling of duplicate brand name
func TestCreateBrand_DuplicateError(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateBrandRequest{}
	input.Body.Name = "Existing Brand"

	duplicateErr := domainErrors.NewDuplicateError("Brand", "name", "Existing Brand")
	mockService.On("CreateBrand", ctx, mock.Anything).Return(duplicateErr)

	// Act
	response, err := handler.CreateBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 409, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestListBrands_Success tests successful brand listing
func TestListBrands_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	now := time.Now()
	brands := []*entities.Brand{
		{ID: uuid.New(), Name: "Brand 1", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Name: "Brand 2", CreatedAt: now, UpdatedAt: now},
	}

	mockService.On("ListBrands", ctx).Return(brands, nil)

	// Act
	response, err := handler.ListBrands(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Body.Brands, 2)
	assert.Equal(t, "Brand 1", response.Body.Brands[0].Name)
	assert.Equal(t, "Brand 2", response.Body.Brands[1].Name)
	mockService.AssertExpectations(t)
}

// TestListBrands_Empty tests listing when no brands exist
func TestListBrands_Empty(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	mockService.On("ListBrands", ctx).Return([]*entities.Brand{}, nil)

	// Act
	response, err := handler.ListBrands(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Empty(t, response.Body.Brands)
	mockService.AssertExpectations(t)
}

// TestGetBrand_Success tests successful brand retrieval by ID
func TestGetBrand_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	now := time.Now()
	brand := &entities.Brand{
		ID:        brandID,
		Name:      "Test Brand",
		CreatedAt: now,
		UpdatedAt: now,
	}

	input := &dto.GetBrandRequest{ID: brandID}
	mockService.On("GetBrand", ctx, brandID).Return(brand, nil)

	// Act
	response, err := handler.GetBrand(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, brandID, response.Body.ID)
	assert.Equal(t, "Test Brand", response.Body.Name)
	mockService.AssertExpectations(t)
}

// TestGetBrand_NotFound tests handling of brand not found
func TestGetBrand_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.GetBrandRequest{ID: brandID}

	notFoundErr := domainErrors.NewNotFoundError("Brand", brandID)
	mockService.On("GetBrand", ctx, brandID).Return(nil, notFoundErr)

	// Act
	response, err := handler.GetBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestGetBrandByName_Success tests successful brand retrieval by name
func TestGetBrandByName_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	now := time.Now()
	brand := &entities.Brand{
		ID:        uuid.New(),
		Name:      "Test Brand",
		CreatedAt: now,
		UpdatedAt: now,
	}

	input := &dto.GetBrandByNameRequest{Name: "Test Brand"}
	mockService.On("GetBrandByName", ctx, "Test Brand").Return(brand, nil)

	// Act
	response, err := handler.GetBrandByName(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Brand", response.Body.Name)
	mockService.AssertExpectations(t)
}

// TestGetBrandByName_NotFound tests handling of brand not found by name
func TestGetBrandByName_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	input := &dto.GetBrandByNameRequest{Name: "Nonexistent Brand"}

	notFoundErr := domainErrors.NewNotFoundError("Brand", "Nonexistent Brand")
	mockService.On("GetBrandByName", ctx, "Nonexistent Brand").Return(nil, notFoundErr)

	// Act
	response, err := handler.GetBrandByName(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestUpdateBrand_Success tests successful brand update
func TestUpdateBrand_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.UpdateBrandRequest{ID: brandID}
	input.Body.Name = "Updated Brand"

	mockService.On("UpdateBrand", ctx, mock.MatchedBy(func(b *entities.Brand) bool {
		return b.ID == brandID && b.Name == "Updated Brand"
	})).Run(func(args mock.Arguments) {
		// Simulate service updating timestamps
		brand := args.Get(1).(*entities.Brand)
		brand.UpdatedAt = time.Now()
	}).Return(nil)

	// Act
	response, err := handler.UpdateBrand(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, brandID, response.Body.ID)
	assert.Equal(t, "Updated Brand", response.Body.Name)
	mockService.AssertExpectations(t)
}

// TestUpdateBrand_ValidationError tests handling of validation errors during update
func TestUpdateBrand_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.UpdateBrandRequest{ID: brandID}
	input.Body.Name = "" // Empty name

	validationErr := domainErrors.NewValidationError("name", "Name is required")
	mockService.On("UpdateBrand", ctx, mock.Anything).Return(validationErr)

	// Act
	response, err := handler.UpdateBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestUpdateBrand_NotFound tests handling of brand not found during update
func TestUpdateBrand_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.UpdateBrandRequest{ID: brandID}
	input.Body.Name = "Updated Brand"

	notFoundErr := domainErrors.NewNotFoundError("Brand", brandID)
	mockService.On("UpdateBrand", ctx, mock.Anything).Return(notFoundErr)

	// Act
	response, err := handler.UpdateBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestDeleteBrand_Success tests successful brand deletion
func TestDeleteBrand_Success(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.DeleteBrandRequest{ID: brandID}

	mockService.On("DeleteBrand", ctx, brandID).Return(nil)

	// Act
	response, err := handler.DeleteBrand(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

// TestDeleteBrand_NotFound tests handling of brand not found during deletion
func TestDeleteBrand_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockBrandService)
	handler := NewBrandHandler(mockService)
	ctx := context.Background()

	brandID := uuid.New()
	input := &dto.DeleteBrandRequest{ID: brandID}

	notFoundErr := domainErrors.NewNotFoundError("Brand", brandID)
	mockService.On("DeleteBrand", ctx, brandID).Return(notFoundErr)

	// Act
	response, err := handler.DeleteBrand(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)

	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}
