package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// CategoryService defines the interface for category business logic operations
type CategoryService interface {
	CreateCategory(ctx context.Context, category *entities.Category) error
	GetCategory(ctx context.Context, id uuid.UUID) (*entities.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*entities.Category, error)
	ListCategories(ctx context.Context) ([]*entities.Category, error)
	ListCategoriesByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
