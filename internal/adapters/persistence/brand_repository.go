package persistence

import (
	"context"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/brand"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"github.com/google/uuid"
)

// BrandRepositoryImpl implements the BrandRepository interface using Ent
type BrandRepositoryImpl struct {
	client *ent.Client
}

func NewBrandRepository(client *ent.Client) *BrandRepositoryImpl {
	return &BrandRepositoryImpl{client: client}
}

func (r *BrandRepositoryImpl) Create(ctx context.Context, b *entities.Brand) error {
	created, err := r.client.Brand.
		Create().
		SetName(b.Name).
		Save(ctx)
	if err != nil {
		// Check for unique constraint violation
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Brand", "name", b.Name)
		}
		return err
	}

	b.ID = created.ID
	b.CreatedAt = created.CreatedAt
	b.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *BrandRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Brand, error) {
	found, err := r.client.Brand.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Brand", id)
		}
		return nil, err
	}

	return &entities.Brand{
		ID:        found.ID,
		Name:      found.Name,
		CreatedAt: found.CreatedAt,
		UpdatedAt: found.UpdatedAt,
	}, nil
}

func (r *BrandRepositoryImpl) GetByName(ctx context.Context, name string) (*entities.Brand, error) {
	found, err := r.client.Brand.
		Query().
		Where(brand.NameEQ(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Brand", name)
		}
		return nil, err
	}

	return &entities.Brand{
		ID:        found.ID,
		Name:      found.Name,
		CreatedAt: found.CreatedAt,
		UpdatedAt: found.UpdatedAt,
	}, nil
}

func (r *BrandRepositoryImpl) List(ctx context.Context) ([]*entities.Brand, error) {
	list, err := r.client.Brand.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	brands := make([]*entities.Brand, 0, len(list))
	for _, b := range list {
		brands = append(brands, &entities.Brand{
			ID:        b.ID,
			Name:      b.Name,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return brands, nil
}

func (r *BrandRepositoryImpl) Update(ctx context.Context, b *entities.Brand) error {
	updated, err := r.client.Brand.
		UpdateOneID(b.ID).
		SetName(b.Name).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Brand", b.ID)
		}
		// Check for unique constraint violation
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Brand", "name", b.Name)
		}
		return err
	}

	b.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *BrandRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Brand.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Brand", id)
		}
		return err
	}
	return nil
}
