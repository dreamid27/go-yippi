package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBrandRepository is a mock implementation of ports.BrandRepository
type MockBrandRepository struct {
	mock.Mock
}

func (m *MockBrandRepository) Create(ctx context.Context, brand *entities.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Brand, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Brand), args.Error(1)
}

func (m *MockBrandRepository) GetByName(ctx context.Context, name string) (*entities.Brand, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Brand), args.Error(1)
}

func (m *MockBrandRepository) List(ctx context.Context) ([]*entities.Brand, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Brand), args.Error(1)
}

func (m *MockBrandRepository) Update(ctx context.Context, brand *entities.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateBrand_Success tests successful brand creation
func TestCreateBrand_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		Name: "Test Brand",
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(b *entities.Brand) bool {
		return b.Name == "Test Brand"
	})).Return(nil)

	// Act
	err := service.CreateBrand(ctx, brand)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestCreateBrand_EmptyName tests validation error for empty brand name
func TestCreateBrand_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		Name: "",
	}

	// Act
	err := service.CreateBrand(ctx, brand)

	// Assert
	require.Error(t, err)
	var validationErr *domainErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "name", validationErr.Field)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateBrand_WhitespaceOnlyName tests validation error for whitespace-only brand name
func TestCreateBrand_WhitespaceOnlyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		Name: "   ",
	}

	// Act
	err := service.CreateBrand(ctx, brand)

	// Assert
	require.Error(t, err)
	var validationErr *domainErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "name", validationErr.Field)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateBrand_NameTooLong tests validation error for brand name exceeding 255 characters
func TestCreateBrand_NameTooLong(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		Name: strings.Repeat("a", 256),
	}

	// Act
	err := service.CreateBrand(ctx, brand)

	// Assert
	require.Error(t, err)
	var validationErr *domainErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "name", validationErr.Field)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreateBrand_DuplicateName tests handling of duplicate brand name
func TestCreateBrand_DuplicateName(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		Name: "Existing Brand",
	}

	mockRepo.On("Create", ctx, brand).Return(domainErrors.NewDuplicateError("Brand", "name", "Existing Brand"))

	// Act
	err := service.CreateBrand(ctx, brand)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrDuplicateEntry))
	mockRepo.AssertExpectations(t)
}

// TestGetBrand_Success tests successful brand retrieval by ID
func TestGetBrand_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brandID := uuid.New()
	expectedBrand := &entities.Brand{
		ID:   brandID,
		Name: "Test Brand",
	}

	mockRepo.On("GetByID", ctx, brandID).Return(expectedBrand, nil)

	// Act
	result, err := service.GetBrand(ctx, brandID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBrand, result)
	mockRepo.AssertExpectations(t)
}

// TestGetBrand_NotFound tests handling of brand not found
func TestGetBrand_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brandID := uuid.New()
	mockRepo.On("GetByID", ctx, brandID).Return(nil, domainErrors.NewNotFoundError("Brand", brandID))

	// Act
	result, err := service.GetBrand(ctx, brandID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domainErrors.ErrNotFound))
	mockRepo.AssertExpectations(t)
}

// TestGetBrandByName_Success tests successful brand retrieval by name
func TestGetBrandByName_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brandName := "Test Brand"
	expectedBrand := &entities.Brand{
		ID:   uuid.New(),
		Name: brandName,
	}

	mockRepo.On("GetByName", ctx, brandName).Return(expectedBrand, nil)

	// Act
	result, err := service.GetBrandByName(ctx, brandName)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBrand, result)
	mockRepo.AssertExpectations(t)
}

// TestGetBrandByName_EmptyName tests validation error for empty name
func TestGetBrandByName_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	// Act
	result, err := service.GetBrandByName(ctx, "")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	var validationErr *domainErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	mockRepo.AssertNotCalled(t, "GetByName")
}

// TestListBrands_Success tests successful brand listing
func TestListBrands_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	expectedBrands := []*entities.Brand{
		{ID: uuid.New(), Name: "Brand 1"},
		{ID: uuid.New(), Name: "Brand 2"},
	}

	mockRepo.On("List", ctx).Return(expectedBrands, nil)

	// Act
	result, err := service.ListBrands(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBrands, result)
	mockRepo.AssertExpectations(t)
}

// TestUpdateBrand_Success tests successful brand update
func TestUpdateBrand_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		ID:   uuid.New(),
		Name: "Updated Brand",
	}

	mockRepo.On("Update", ctx, brand).Return(nil)

	// Act
	err := service.UpdateBrand(ctx, brand)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUpdateBrand_EmptyName tests validation error for empty brand name
func TestUpdateBrand_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brand := &entities.Brand{
		ID:   uuid.New(),
		Name: "",
	}

	// Act
	err := service.UpdateBrand(ctx, brand)

	// Assert
	require.Error(t, err)
	var validationErr *domainErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	mockRepo.AssertNotCalled(t, "Update")
}

// TestDeleteBrand_Success tests successful brand deletion
func TestDeleteBrand_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brandID := uuid.New()
	mockRepo.On("Delete", ctx, brandID).Return(nil)

	// Act
	err := service.DeleteBrand(ctx, brandID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteBrand_NotFound tests handling of brand not found during deletion
func TestDeleteBrand_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockBrandRepository)
	service := NewBrandService(mockRepo)
	ctx := context.Background()

	brandID := uuid.New()
	mockRepo.On("Delete", ctx, brandID).Return(domainErrors.NewNotFoundError("Brand", brandID))

	// Act
	err := service.DeleteBrand(ctx, brandID)

	// Assert
	require.Error(t, err)
	assert.True(t, errors.Is(err, domainErrors.ErrNotFound))
	mockRepo.AssertExpectations(t)
}
