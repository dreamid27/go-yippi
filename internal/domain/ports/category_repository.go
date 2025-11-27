package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *entities.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Category, error)
	GetByName(ctx context.Context, name string) (*entities.Category, error)
	List(ctx context.Context) ([]*entities.Category, error)
	ListByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetDescendantIDs(ctx context.Context, categoryIDs []uuid.UUID) ([]uuid.UUID, error)
}
