package services

import (
	"context"
	"testing"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCategoryRepository is a mock implementation of ports.CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *entities.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id int) (*entities.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByName(ctx context.Context, name string) (*entities.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Category), args.Error(1)
}

func (m *MockCategoryRepository) List(ctx context.Context) ([]*entities.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Category), args.Error(1)
}

func (m *MockCategoryRepository) ListByParentID(ctx context.Context, parentID *int) ([]*entities.Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateCategory_Success tests successful category creation
func TestCreateCategory_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	category := &entities.Category{
		Name: "Electronics",
	}

	mockRepo.On("Create", ctx, category).Return(nil)

	// Act
	err := service.CreateCategory(ctx, category)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestCreateCategory_WithParent tests successful category creation with parent
func TestCreateCategory_WithParent(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := 1
	category := &entities.Category{
		Name:     "Laptops",
		ParentID: &parentID,
	}

	parentCategory := &entities.Category{
		ID:   1,
		Name: "Electronics",
	}

	mockRepo.On("GetByID", ctx, parentID).Return(parentCategory, nil)
	mockRepo.On("Create", ctx, category).Return(nil)

	// Act
	err := service.CreateCategory(ctx, category)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestCreateCategory_ValidationError tests validation errors
func TestCreateCategory_ValidationError(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	category := &entities.Category{
		Name: "", // Empty name
	}

	// Act
	err := service.CreateCategory(ctx, category)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name is required")
	mockRepo.AssertExpectations(t)
}

// TestCreateCategory_InvalidParent tests creation with non-existent parent
func TestCreateCategory_InvalidParent(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := 999
	category := &entities.Category{
		Name:     "Laptops",
		ParentID: &parentID,
	}

	mockRepo.On("GetByID", ctx, parentID).Return(nil, domainErrors.NewNotFoundError("Category", parentID))

	// Act
	err := service.CreateCategory(ctx, category)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Parent category does not exist")
	mockRepo.AssertExpectations(t)
}

// TestUpdateCategory_Success tests successful category update
func TestUpdateCategory_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	category := &entities.Category{
		ID:   1,
		Name: "Updated Electronics",
	}

	mockRepo.On("Update", ctx, category).Return(nil)

	// Act
	err := service.UpdateCategory(ctx, category)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUpdateCategory_SelfParent tests preventing category from being its own parent
func TestUpdateCategory_SelfParent(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	categoryID := 1
	category := &entities.Category{
		ID:       categoryID,
		Name:     "Electronics",
		ParentID: &categoryID, // Same as ID
	}

	// Act
	err := service.UpdateCategory(ctx, category)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Category cannot be its own parent")
	mockRepo.AssertExpectations(t)
}

// TestDeleteCategory_Success tests successful category deletion
func TestDeleteCategory_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	categoryID := 1
	category := &entities.Category{
		ID:   categoryID,
		Name: "Electronics",
	}

	mockRepo.On("GetByID", ctx, categoryID).Return(category, nil)
	mockRepo.On("ListByParentID", ctx, &categoryID).Return([]*entities.Category{}, nil)
	mockRepo.On("Delete", ctx, categoryID).Return(nil)

	// Act
	err := service.DeleteCategory(ctx, categoryID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteCategory_WithChildren tests preventing deletion of category with children
func TestDeleteCategory_WithChildren(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	categoryID := 1
	category := &entities.Category{
		ID:   categoryID,
		Name: "Electronics",
	}

	children := []*entities.Category{
		{ID: 2, Name: "Laptops", ParentID: &categoryID},
	}

	mockRepo.On("GetByID", ctx, categoryID).Return(category, nil)
	mockRepo.On("ListByParentID", ctx, &categoryID).Return(children, nil)

	// Act
	err := service.DeleteCategory(ctx, categoryID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete category with children")
	mockRepo.AssertExpectations(t)
}

// TestListCategories_Success tests successful listing of categories
func TestListCategories_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	expectedCategories := []*entities.Category{
		{ID: 1, Name: "Electronics"},
		{ID: 2, Name: "Books"},
	}

	mockRepo.On("List", ctx).Return(expectedCategories, nil)

	// Act
	categories, err := service.ListCategories(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)
	mockRepo.AssertExpectations(t)
}

// TestListCategoriesByParentID_Success tests successful listing of categories by parent
func TestListCategoriesByParentID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockCategoryRepository)
	service := NewCategoryService(mockRepo)
	ctx := context.Background()

	parentID := 1
	parentCategory := &entities.Category{
		ID:   parentID,
		Name: "Electronics",
	}

	expectedCategories := []*entities.Category{
		{ID: 2, Name: "Laptops", ParentID: &parentID},
		{ID: 3, Name: "Phones", ParentID: &parentID},
	}

	mockRepo.On("GetByID", ctx, parentID).Return(parentCategory, nil)
	mockRepo.On("ListByParentID", ctx, &parentID).Return(expectedCategories, nil)

	// Act
	categories, err := service.ListCategoriesByParentID(ctx, &parentID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)
	mockRepo.AssertExpectations(t)
}
