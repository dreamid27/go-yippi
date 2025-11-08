package persistence

import (
	"context"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/category"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/google/uuid"
)

// CategoryRepositoryImpl implements the CategoryRepository interface using Ent
type CategoryRepositoryImpl struct {
	client *ent.Client
}

func NewCategoryRepository(client *ent.Client) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{client: client}
}

func (r *CategoryRepositoryImpl) Create(ctx context.Context, cat *entities.Category) error {
	builder := r.client.Category.
		Create().
		SetName(cat.Name)

	// Set parent ID if provided
	if cat.ParentID != nil {
		builder = builder.SetParentID(*cat.ParentID)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Category", "name", cat.Name)
		}
		return err
	}

	cat.ID = created.ID
	cat.CreatedAt = created.CreatedAt
	cat.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *CategoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	found, err := r.client.Category.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Category", id)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *CategoryRepositoryImpl) GetByName(ctx context.Context, name string) (*entities.Category, error) {
	found, err := r.client.Category.
		Query().
		Where(category.NameEQ(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Category", name)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *CategoryRepositoryImpl) List(ctx context.Context) ([]*entities.Category, error) {
	list, err := r.client.Category.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]*entities.Category, 0, len(list))
	for _, c := range list {
		categories = append(categories, r.toEntity(c))
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) ListByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Category, error) {
	query := r.client.Category.Query()

	if parentID == nil {
		// Get root categories (no parent)
		query = query.Where(category.Not(category.HasParent()))
	} else {
		// Get categories with specific parent
		query = query.Where(category.HasParentWith(category.IDEQ(*parentID)))
	}

	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]*entities.Category, 0, len(list))
	for _, c := range list {
		categories = append(categories, r.toEntity(c))
	}

	return categories, nil
}

func (r *CategoryRepositoryImpl) Update(ctx context.Context, cat *entities.Category) error {
	builder := r.client.Category.
		UpdateOneID(cat.ID).
		SetName(cat.Name)

	// Update parent ID
	if cat.ParentID != nil {
		builder = builder.SetParentID(*cat.ParentID)
	} else {
		builder = builder.ClearParent()
	}

	_, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Category", cat.ID)
		}
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Category", "name", cat.Name)
		}
		return err
	}
	return nil
}

func (r *CategoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Category.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Category", id)
		}
		return err
	}
	return nil
}

// toEntity converts Ent Category to domain entity
func (r *CategoryRepositoryImpl) toEntity(c *ent.Category) *entities.Category {
	cat := &entities.Category{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	// Set parent ID if it exists (c.ParentID is already *uuid.UUID from Ent)
	if c.ParentID != nil {
		cat.ParentID = c.ParentID
	}

	return cat
}

// GetDescendantIDs returns all descendant category IDs for the given category IDs (including the given IDs)
func (r *CategoryRepositoryImpl) GetDescendantIDs(ctx context.Context, categoryIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(categoryIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	result := make(map[uuid.UUID]bool)
	for _, id := range categoryIDs {
		result[id] = true // Include the parent category itself
	}

	// Recursively find all descendants
	toProcess := categoryIDs
	for len(toProcess) > 0 {
		var nextBatch []uuid.UUID
		for _, parentID := range toProcess {
			// Find children of this parent
			children, err := r.ListByParentID(ctx, &parentID)
			if err != nil {
				return nil, err
			}

			for _, child := range children {
				if !result[child.ID] {
					result[child.ID] = true
					nextBatch = append(nextBatch, child.ID)
				}
			}
		}
		toProcess = nextBatch
	}

	// Convert map to slice
	ids := make([]uuid.UUID, 0, len(result))
	for id := range result {
		ids = append(ids, id)
	}

	return ids, nil
}
