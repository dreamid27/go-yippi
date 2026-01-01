package services

import (
	"context"
	"fmt"

	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"example.com/go-yippi/internal/infrastructure/observability"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// stockValidatorService implements the StockValidatorService interface
type stockValidatorService struct {
	variantRepo ports.ProductVariantRepository
	observer    *observability.ServiceObserver
}

// NewStockValidatorService creates a new stock validator service
func NewStockValidatorService(variantRepo ports.ProductVariantRepository) ports.StockValidatorService {
	return &stockValidatorService{
		variantRepo: variantRepo,
		observer:    observability.NewServiceObserver("StockValidatorService"),
	}
}

// ValidateStock validates if requested quantity is available for a variant
func (s *stockValidatorService) ValidateStock(ctx context.Context, variantID uuid.UUID, requestedQty int) error {
	op := s.observer.StartOperation(ctx, "ValidateStock")
	defer op.End(nil)

	op.AddAttribute("variant_id", variantID.String())
	op.AddAttribute("requested_qty", requestedQty)

	// Validation: requestedQty must be positive
	if requestedQty <= 0 {
		err := errors.NewValidationError("quantity", "requested quantity must be greater than 0")
		observability.LogValidationError(op.Context(), "quantity", "requested quantity must be greater than 0")
		op.End(err)
		return err
	}

	// Find variant
	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		op.End(err)
		return err
	}

	op.AddAttribute("variant_sku", variant.SKU)
	op.AddAttribute("stock_available", variant.StockQuantity)

	// Check if variant is active
	if !variant.IsActive {
		err := errors.NewVariantInactiveError(variantID)
		observability.LogBusinessRule(op.Context(), "variant_must_be_active",
			fmt.Sprintf("Variant %s is inactive", variant.SKU))
		op.End(err)
		return err
	}

	// Check stock availability
	if variant.StockQuantity < requestedQty {
		err := errors.NewInsufficientStockError(variantID, variant.StockQuantity, requestedQty)
		op.LogWarn("Insufficient stock",
			zap.String("variant_sku", variant.SKU),
			zap.Int("available", variant.StockQuantity),
			zap.Int("requested", requestedQty))
		op.End(err)
		return err
	}

	op.LogInfo("Stock validation passed",
		zap.String("variant_sku", variant.SKU),
		zap.Int("available", variant.StockQuantity),
		zap.Int("requested", requestedQty))

	op.RecordBusinessEvent("stock", "stock_validated")
	op.End(nil)
	return nil
}

// ValidateCartStock validates stock for all items in a cart
func (s *stockValidatorService) ValidateCartStock(ctx context.Context, cartItems []*entities.CartItem) ([]ports.StockValidationError, error) {
	op := s.observer.StartOperation(ctx, "ValidateCartStock")
	defer op.End(nil)

	op.AddAttribute("cart_items_count", len(cartItems))

	if len(cartItems) == 0 {
		op.LogInfo("Empty cart, no validation needed")
		op.End(nil)
		return nil, nil
	}

	var validationErrors []ports.StockValidationError

	// Validate each cart item (don't fail fast - collect all errors)
	for _, item := range cartItems {
		// Find variant
		variant, err := s.variantRepo.FindByID(op.Context(), item.ProductVariantID)
		if err != nil {
			// Variant not found
			validationErrors = append(validationErrors, ports.StockValidationError{
				CartItemID: item.ID,
				VariantID:  item.ProductVariantID,
				Available:  0,
				InCart:     item.Quantity,
				Message:    "Product variant not found or has been removed",
			})
			op.LogWarn("Variant not found in cart validation",
				zap.String("variant_id", item.ProductVariantID.String()),
				zap.String("cart_item_id", item.ID.String()))
			continue
		}

		// Check if variant is active
		if !variant.IsActive {
			validationErrors = append(validationErrors, ports.StockValidationError{
				CartItemID: item.ID,
				VariantID:  item.ProductVariantID,
				Available:  variant.StockQuantity,
				InCart:     item.Quantity,
				Message:    fmt.Sprintf("Product variant '%s' is no longer available", variant.SKU),
			})
			op.LogWarn("Inactive variant in cart",
				zap.String("variant_sku", variant.SKU),
				zap.String("cart_item_id", item.ID.String()))
			continue
		}

		// Check stock availability
		if variant.StockQuantity < item.Quantity {
			validationErrors = append(validationErrors, ports.StockValidationError{
				CartItemID: item.ID,
				VariantID:  item.ProductVariantID,
				Available:  variant.StockQuantity,
				InCart:     item.Quantity,
				Message:    fmt.Sprintf("Insufficient stock for '%s': only %d available, %d in cart", variant.SKU, variant.StockQuantity, item.Quantity),
			})
			op.LogWarn("Insufficient stock in cart",
				zap.String("variant_sku", variant.SKU),
				zap.Int("available", variant.StockQuantity),
				zap.Int("in_cart", item.Quantity))
		}
	}

	if len(validationErrors) > 0 {
		op.AddAttribute("validation_errors_count", len(validationErrors))
		op.LogWarn("Cart stock validation failed",
			zap.Int("errors_count", len(validationErrors)))
		op.RecordBusinessEvent("stock", "cart_stock_validation_failed")
	} else {
		op.LogInfo("Cart stock validation passed")
		op.RecordBusinessEvent("stock", "cart_stock_validated")
	}

	op.End(nil)
	return validationErrors, nil
}
