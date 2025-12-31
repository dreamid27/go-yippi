package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/productvariant"
	"example.com/go-yippi/internal/domain/entities"
)

// ProductVariantRepository implements ports.ProductVariantRepository
type ProductVariantRepository struct {
	client *ent.Client
}

// NewProductVariantRepository creates a new product variant repository
func NewProductVariantRepository(client *ent.Client) *ProductVariantRepository {
	return &ProductVariantRepository{client: client}
}

// Create creates a new product variant
func (r *ProductVariantRepository) Create(ctx context.Context, variant *entities.ProductVariant) error {
	created, err := r.client.ProductVariant.
		Create().
		SetProductID(variant.ProductID).
		SetSku(variant.SKU).
		SetAttributes(variant.Attributes).
		SetStockQuantity(variant.StockQuantity).
		SetPriceAdjustment(variant.PriceAdjustment).
		SetIsActive(variant.IsActive).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create product variant: %w", err)
	}

	variant.ID = created.ID
	variant.CreatedAt = created.CreatedAt
	variant.UpdatedAt = created.UpdatedAt

	return nil
}

// FindByID finds a variant by ID
func (r *ProductVariantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error) {
	v, err := r.client.ProductVariant.
		Query().
		Where(productvariant.ID(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("product variant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find product variant: %w", err)
	}

	return toProductVariantEntity(v), nil
}

// FindBySKU finds a variant by SKU
func (r *ProductVariantRepository) FindBySKU(ctx context.Context, sku string) (*entities.ProductVariant, error) {
	v, err := r.client.ProductVariant.
		Query().
		Where(productvariant.Sku(sku)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("product variant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find product variant: %w", err)
	}

	return toProductVariantEntity(v), nil
}

// FindByProductID finds all variants for a product
func (r *ProductVariantRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error) {
	variants, err := r.client.ProductVariant.
		Query().
		Where(productvariant.ProductID(productID)).
		Order(ent.Asc(productvariant.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find product variants: %w", err)
	}

	result := make([]*entities.ProductVariant, len(variants))
	for i, v := range variants {
		result[i] = toProductVariantEntity(v)
	}

	return result, nil
}

// Update updates an existing variant
func (r *ProductVariantRepository) Update(ctx context.Context, variant *entities.ProductVariant) error {
	updated, err := r.client.ProductVariant.
		UpdateOneID(variant.ID).
		SetSku(variant.SKU).
		SetAttributes(variant.Attributes).
		SetStockQuantity(variant.StockQuantity).
		SetPriceAdjustment(variant.PriceAdjustment).
		SetIsActive(variant.IsActive).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product variant: %w", err)
	}

	variant.UpdatedAt = updated.UpdatedAt

	return nil
}

// UpdateStock updates stock quantity for a variant
func (r *ProductVariantRepository) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	_, err := r.client.ProductVariant.
		UpdateOneID(id).
		SetStockQuantity(quantity).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// Delete deletes a variant
func (r *ProductVariantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.ProductVariant.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	return nil
}

// toProductVariantEntity converts Ent model to domain entity
func toProductVariantEntity(v *ent.ProductVariant) *entities.ProductVariant {
	return &entities.ProductVariant{
		ID:              v.ID,
		ProductID:       v.ProductID,
		SKU:             v.Sku,
		Attributes:      v.Attributes,
		StockQuantity:   v.StockQuantity,
		PriceAdjustment: v.PriceAdjustment,
		IsActive:        v.IsActive,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}
