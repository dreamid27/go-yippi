package persistence

import (
	"context"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/product"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
)

// ProductRepositoryImpl implements the ProductRepository interface using Ent
type ProductRepositoryImpl struct {
	client *ent.Client
}

func NewProductRepository(client *ent.Client) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{client: client}
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, prod *entities.Product) error {
	builder := r.client.Product.
		Create().
		SetSku(prod.SKU).
		SetSlug(prod.Slug).
		SetName(prod.Name).
		SetPrice(prod.Price).
		SetDescription(prod.Description).
		SetWeight(prod.Weight).
		SetLength(prod.Length).
		SetWidth(prod.Width).
		SetHeight(prod.Height)

	// Set image URLs if provided
	if prod.ImageURLs != nil {
		builder = builder.SetImageUrls(prod.ImageURLs)
	}

	// Set category ID if provided
	if prod.CategoryID != nil {
		builder = builder.SetCategoryID(*prod.CategoryID)
	}

	// Set brand ID if provided
	if prod.BrandID != nil {
		builder = builder.SetBrandID(*prod.BrandID)
	}

	created, err := builder.
		SetStatus(product.Status(prod.Status)).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Product", "sku or slug", prod.SKU)
		}
		return err
	}

	prod.ID = created.ID
	prod.CreatedAt = created.CreatedAt
	prod.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id int) (*entities.Product, error) {
	found, err := r.client.Product.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Product", id)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *ProductRepositoryImpl) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	found, err := r.client.Product.
		Query().
		Where(product.SkuEQ(sku)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Product", sku)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *ProductRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Product, error) {
	found, err := r.client.Product.
		Query().
		Where(product.Slug(slug)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("Product", slug)
		}
		return nil, err
	}

	return r.toEntity(found), nil
}

func (r *ProductRepositoryImpl) List(ctx context.Context) ([]*entities.Product, error) {
	list, err := r.client.Product.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, 0, len(list))
	for _, p := range list {
		products = append(products, r.toEntity(p))
	}

	return products, nil
}

func (r *ProductRepositoryImpl) ListByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error) {
	list, err := r.client.Product.
		Query().
		Where(product.StatusEQ(product.Status(status))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*entities.Product, 0, len(list))
	for _, p := range list {
		products = append(products, r.toEntity(p))
	}

	return products, nil
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, prod *entities.Product) error {
	builder := r.client.Product.
		UpdateOneID(prod.ID).
		SetSku(prod.SKU).
		SetSlug(prod.Slug).
		SetName(prod.Name).
		SetPrice(prod.Price).
		SetDescription(prod.Description).
		SetWeight(prod.Weight).
		SetLength(prod.Length).
		SetWidth(prod.Width).
		SetHeight(prod.Height)

	// Set image URLs if provided
	if prod.ImageURLs != nil {
		builder = builder.SetImageUrls(prod.ImageURLs)
	}

	// Set or clear category ID
	if prod.CategoryID != nil {
		builder = builder.SetCategoryID(*prod.CategoryID)
	} else {
		builder = builder.ClearCategory()
	}

	// Set or clear brand ID
	if prod.BrandID != nil {
		builder = builder.SetBrandID(*prod.BrandID)
	} else {
		builder = builder.ClearBrand()
	}

	_, err := builder.
		SetStatus(product.Status(prod.Status)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Product", prod.ID)
		}
		if ent.IsConstraintError(err) {
			return domainErrors.NewDuplicateError("Product", "sku or slug", prod.SKU)
		}
		return err
	}
	return nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id int) error {
	err := r.client.Product.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("Product", id)
		}
		return err
	}
	return nil
}

// toEntity converts Ent Product to domain entity
func (r *ProductRepositoryImpl) toEntity(p *ent.Product) *entities.Product {
	product := &entities.Product{
		ID:          p.ID,
		SKU:         p.Sku,
		Slug:        p.Slug,
		Name:        p.Name,
		Price:       p.Price,
		Description: p.Description,
		Weight:      p.Weight,
		Length:      p.Length,
		Width:       p.Width,
		Height:      p.Height,
		ImageURLs:   p.ImageUrls,
		Status:      entities.ProductStatus(p.Status),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Set category ID if it exists
	if p.Edges.Category != nil {
		product.CategoryID = &p.Edges.Category.ID
	} else if categoryID, exists := p.QueryCategory().OnlyID(context.Background()); exists == nil {
		product.CategoryID = &categoryID
	}

	// Set brand ID if it exists
	if p.Edges.Brand != nil {
		product.BrandID = &p.Edges.Brand.ID
	} else if brandID, exists := p.QueryBrand().OnlyID(context.Background()); exists == nil {
		product.BrandID = &brandID
	}

	return product
}
