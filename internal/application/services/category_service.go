package services

import (
	"context"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/google/uuid"
)

// CategoryService handles business logic for categories
type CategoryService struct {
	repo ports.CategoryRepository
}

func NewCategoryService(repo ports.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *entities.Category) error {
	// Validate required fields
	if strings.TrimSpace(category.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}

	// Validate parent exists if provided
	if category.ParentID != nil {
		_, err := s.repo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return domainErrors.NewValidationError("parent_id", "Parent category does not exist")
		}
	}

	return s.repo.Create(ctx, category)
}

func (s *CategoryService) GetCategory(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) GetCategoryByName(ctx context.Context, name string) (*entities.Category, error) {
	if strings.TrimSpace(name) == "" {
		return nil, domainErrors.NewValidationError("name", "Name is required")
	}
	return s.repo.GetByName(ctx, name)
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]*entities.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) ListCategoriesByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Category, error) {
	// Validate parent exists if provided
	if parentID != nil {
		_, err := s.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, domainErrors.NewValidationError("parent_id", "Parent category does not exist")
		}
	}

	return s.repo.ListByParentID(ctx, parentID)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *entities.Category) error {
	// Validate required fields
	if strings.TrimSpace(category.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}

	// Validate parent exists if provided and not the same as category ID
	if category.ParentID != nil {
		if *category.ParentID == category.ID {
			return domainErrors.NewValidationError("parent_id", "Category cannot be its own parent")
		}
		_, err := s.repo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return domainErrors.NewValidationError("parent_id", "Parent category does not exist")
		}
	}

	return s.repo.Update(ctx, category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Check if category exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if category has children
	children, err := s.repo.ListByParentID(ctx, &id)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return domainErrors.NewValidationError("category", "Cannot delete category with children")
	}

	return s.repo.Delete(ctx, id)
}
