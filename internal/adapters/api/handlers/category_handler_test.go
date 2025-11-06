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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCategoryService is a mock implementation of CategoryService
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, category *entities.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryService) GetCategory(ctx context.Context, id int) (*entities.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryByName(ctx context.Context, name string) (*entities.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Category), args.Error(1)
}

func (m *MockCategoryService) ListCategories(ctx context.Context) ([]*entities.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Category), args.Error(1)
}

func (m *MockCategoryService) ListCategoriesByParentID(ctx context.Context, parentID *int) ([]*entities.Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Category), args.Error(1)
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, category *entities.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestCreateCategory_Success tests successful category creation
func TestCreateCategory_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateCategoryRequest{}
	input.Body.Name = "Electronics"

	mockService.On("CreateCategory", ctx, mock.MatchedBy(func(c *entities.Category) bool {
		return c.Name == "Electronics" && c.ParentID == nil
	})).Return(nil).Run(func(args mock.Arguments) {
		// Simulate ID assignment by service/repository
		cat := args.Get(1).(*entities.Category)
		cat.ID = 1
		cat.CreatedAt = time.Now()
		cat.UpdatedAt = time.Now()
	})

	// Act
	response, err := handler.CreateCategory(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, response.Body.ID)
	assert.Equal(t, "Electronics", response.Body.Name)
	assert.Nil(t, response.Body.ParentID)
	mockService.AssertExpectations(t)
}

// TestCreateCategory_WithParent tests successful category creation with parent
func TestCreateCategory_WithParent(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	parentID := 1
	input := &dto.CreateCategoryRequest{}
	input.Body.Name = "Laptops"
	input.Body.ParentID = &parentID

	mockService.On("CreateCategory", ctx, mock.MatchedBy(func(c *entities.Category) bool {
		return c.Name == "Laptops" && c.ParentID != nil && *c.ParentID == 1
	})).Return(nil).Run(func(args mock.Arguments) {
		cat := args.Get(1).(*entities.Category)
		cat.ID = 2
		cat.CreatedAt = time.Now()
		cat.UpdatedAt = time.Now()
	})

	// Act
	response, err := handler.CreateCategory(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 2, response.Body.ID)
	assert.Equal(t, "Laptops", response.Body.Name)
	assert.Equal(t, 1, *response.Body.ParentID)
	mockService.AssertExpectations(t)
}

// TestCreateCategory_ValidationError tests validation error handling
func TestCreateCategory_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateCategoryRequest{}
	input.Body.Name = ""

	validationErr := domainErrors.NewValidationError("name", "Name is required")
	mockService.On("CreateCategory", ctx, mock.Anything).Return(validationErr)

	// Act
	response, err := handler.CreateCategory(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestCreateCategory_DuplicateError tests duplicate entry error handling
func TestCreateCategory_DuplicateError(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.CreateCategoryRequest{}
	input.Body.Name = "Electronics"

	duplicateErr := domainErrors.NewDuplicateError("Category", "name", "Electronics")
	mockService.On("CreateCategory", ctx, mock.Anything).Return(duplicateErr)

	// Act
	response, err := handler.CreateCategory(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 409, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestGetCategory_Success tests successful category retrieval
func TestGetCategory_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.GetCategoryRequest{ID: 1}
	category := &entities.Category{
		ID:        1,
		Name:      "Electronics",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.On("GetCategory", ctx, 1).Return(category, nil)

	// Act
	response, err := handler.GetCategory(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, response.Body.ID)
	assert.Equal(t, "Electronics", response.Body.Name)
	mockService.AssertExpectations(t)
}

// TestGetCategory_NotFound tests not found error handling
func TestGetCategory_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.GetCategoryRequest{ID: 999}
	notFoundErr := domainErrors.NewNotFoundError("Category", 999)

	mockService.On("GetCategory", ctx, 999).Return(nil, notFoundErr)

	// Act
	response, err := handler.GetCategory(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 404, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestListCategories_Success tests successful listing of categories
func TestListCategories_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	categories := []*entities.Category{
		{ID: 1, Name: "Electronics", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Name: "Books", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockService.On("ListCategories", ctx).Return(categories, nil)

	// Act
	response, err := handler.ListCategories(ctx, &struct{}{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Body.Categories, 2)
	assert.Equal(t, "Electronics", response.Body.Categories[0].Name)
	assert.Equal(t, "Books", response.Body.Categories[1].Name)
	mockService.AssertExpectations(t)
}

// TestUpdateCategory_Success tests successful category update
func TestUpdateCategory_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.UpdateCategoryRequest{ID: 1}
	input.Body.Name = "Updated Electronics"

	mockService.On("UpdateCategory", ctx, mock.MatchedBy(func(c *entities.Category) bool {
		return c.ID == 1 && c.Name == "Updated Electronics"
	})).Return(nil).Run(func(args mock.Arguments) {
		cat := args.Get(1).(*entities.Category)
		cat.CreatedAt = time.Now()
		cat.UpdatedAt = time.Now()
	})

	// Act
	response, err := handler.UpdateCategory(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 1, response.Body.ID)
	assert.Equal(t, "Updated Electronics", response.Body.Name)
	mockService.AssertExpectations(t)
}

// TestDeleteCategory_Success tests successful category deletion
func TestDeleteCategory_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.DeleteCategoryRequest{ID: 1}
	mockService.On("DeleteCategory", ctx, 1).Return(nil)

	// Act
	response, err := handler.DeleteCategory(ctx, input)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

// TestDeleteCategory_WithChildren tests deletion error when category has children
func TestDeleteCategory_WithChildren(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	input := &dto.DeleteCategoryRequest{ID: 1}
	validationErr := domainErrors.NewValidationError("category", "Cannot delete category with children")
	mockService.On("DeleteCategory", ctx, 1).Return(validationErr)

	// Act
	response, err := handler.DeleteCategory(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	var humaErr huma.StatusError
	require.True(t, errors.As(err, &humaErr))
	assert.Equal(t, 400, humaErr.GetStatus())
	mockService.AssertExpectations(t)
}

// TestListCategoriesByParent_Success tests listing categories by parent
func TestListCategoriesByParent_Success(t *testing.T) {
	// Arrange
	mockService := new(MockCategoryService)
	handler := NewCategoryHandler(mockService)
	ctx := context.Background()

	parentID := 1
	input := &dto.ListCategoriesByParentRequest{ParentID: &parentID}

	categories := []*entities.Category{
		{ID: 2, Name: "Laptops", ParentID: &parentID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 3, Name: "Phones", ParentID: &parentID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockService.On("ListCategoriesByParentID", ctx, &parentID).Return(categories, nil)

	// Act
	response, err := handler.ListCategoriesByParent(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Body.Categories, 2)
	assert.Equal(t, "Laptops", response.Body.Categories[0].Name)
	assert.Equal(t, "Phones", response.Body.Categories[1].Name)
	mockService.AssertExpectations(t)
}
