package persistence

import (
	"context"

	"github.com/google/uuid"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/productvariant"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// StockValidatorRepository is an Ent adapter for stock validation operations
type StockValidatorRepository struct {
	client *ent.Client
}

// NewStockValidatorRepository creates a new StockValidatorRepository
func NewStockValidatorRepository(client *ent.Client) ports.StockValidatorService {
	return &StockValidatorRepository{
		client: client,
	}
}

// ValidateStock validates if variant has sufficient stock for requested quantity
func (r *StockValidatorRepository) ValidateStock(ctx context.Context, variantID uuid.UUID, requestedQty int) error {
	variant, err := r.client.ProductVariant.
		Query().
		Where(
			productvariant.ID(variantID),
			productvariant.IsActive(true),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return domainErrors.NewNotFoundError("ProductVariant", variantID.String())
		}
		return err
	}

	if !variant.IsActive {
		return domainErrors.NewValidationError("variant_id", "Product variant is inactive")
	}

	if variant.StockQuantity < requestedQty {
		return domainErrors.NewInsufficientStockError(
			variant.ID,
			requestedQty,
			variant.StockQuantity,
		)
	}

	return nil
}

// ValidateCartStock validates stock for all items in cart
func (r *StockValidatorRepository) ValidateCartStock(ctx context.Context, cartItems []*entities.CartItem) ([]ports.StockValidationError, error) {
	var validationErrors []ports.StockValidationError

	for _, item := range cartItems {
		variant, err := r.client.ProductVariant.
			Query().
			Where(
				productvariant.ID(item.ProductVariantID),
				productvariant.IsActive(true),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				validationErrors = append(validationErrors, ports.StockValidationError{
					CartItemID: item.ID,
					VariantID:  item.ProductVariantID,
					Available:  0,
					InCart:     item.Quantity,
					Message:    "Variant not found",
				})
			} else {
				validationErrors = append(validationErrors, ports.StockValidationError{
					CartItemID: item.ID,
					VariantID:  item.ProductVariantID,
					Available:  0,
					InCart:     item.Quantity,
					Message:    err.Error(),
				})
			}
			continue
		}

		if !variant.IsActive {
			validationErrors = append(validationErrors, ports.StockValidationError{
				CartItemID: item.ID,
				VariantID:  variant.ID,
				Available:  variant.StockQuantity,
				InCart:     item.Quantity,
				Message:    "Variant is inactive",
			})
			continue
		}

		if variant.StockQuantity < item.Quantity {
			validationErrors = append(validationErrors, ports.StockValidationError{
				CartItemID: item.ID,
				VariantID:  variant.ID,
				Available:  variant.StockQuantity,
				InCart:     item.Quantity,
				Message:    "Insufficient stock",
			})
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors, nil
	}

	return nil, nil
}
