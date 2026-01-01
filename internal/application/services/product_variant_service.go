package services

import (
	"context"
	"fmt"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"example.com/go-yippi/internal/infrastructure/observability"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// productVariantService implements the ProductVariantService interface
type productVariantService struct {
	variantRepo ports.ProductVariantRepository
	productRepo ports.ProductRepository
	observer    *observability.ServiceObserver
}

// NewProductVariantService creates a new product variant service
func NewProductVariantService(
	variantRepo ports.ProductVariantRepository,
	productRepo ports.ProductRepository,
) ports.ProductVariantService {
	return &productVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
		observer:    observability.NewServiceObserver("ProductVariantService"),
	}
}

// CreateVariant creates a new product variant
func (s *productVariantService) CreateVariant(
	ctx context.Context,
	productID uuid.UUID,
	sku string,
	attributes map[string]string,
	stockQuantity int,
	priceAdjustment float64,
	isActive *bool,
) (*entities.ProductVariant, error) {
	op := s.observer.StartOperation(ctx, "CreateVariant")
	defer op.End(nil)

	op.AddAttribute("product_id", productID.String())
	op.AddAttribute("sku", sku)

	// Validate product exists
	product, err := s.productRepo.GetByID(op.Context(), productID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Product", productID)
		op.End(err)
		return nil, err
	}

	// Trim and validate SKU
	sku = strings.TrimSpace(sku)
	if sku == "" {
		err := errors.NewValidationError("sku", "SKU is required")
		observability.LogValidationError(op.Context(), "sku", "SKU is required")
		op.End(err)
		return nil, err
	}

	if len(sku) < 3 {
		err := errors.NewValidationError("sku", "SKU must be at least 3 characters")
		observability.LogValidationError(op.Context(), "sku", "SKU must be at least 3 characters")
		op.End(err)
		return nil, err
	}

	if len(sku) > 100 {
		err := errors.NewValidationError("sku", "SKU must not exceed 100 characters")
		observability.LogValidationError(op.Context(), "sku", "SKU must not exceed 100 characters")
		op.End(err)
		return nil, err
	}

	// Validate attributes
	if len(attributes) == 0 {
		err := errors.NewValidationError("attributes", "Attributes are required")
		observability.LogValidationError(op.Context(), "attributes", "Attributes are required")
		op.End(err)
		return nil, err
	}

	// Validate attributes have no empty keys or values
	for key, value := range attributes {
		if strings.TrimSpace(key) == "" {
			err := errors.NewValidationError("attributes", "Attribute keys cannot be empty")
			observability.LogValidationError(op.Context(), "attributes", "Attribute keys cannot be empty")
			op.End(err)
			return nil, err
		}
		if strings.TrimSpace(value) == "" {
			err := errors.NewValidationError("attributes", "Attribute values cannot be empty")
			observability.LogValidationError(op.Context(), "attributes", "Attribute values cannot be empty")
			op.End(err)
			return nil, err
		}
	}

	// Validate stock quantity
	if stockQuantity < 0 {
		err := errors.NewValidationError("stock_quantity", "Stock quantity must be >= 0")
		observability.LogValidationError(op.Context(), "stock_quantity", "Stock quantity must be >= 0")
		op.End(err)
		return nil, err
	}

	// Check SKU uniqueness
	existingVariant, err := s.variantRepo.FindBySKU(op.Context(), sku)
	if err == nil && existingVariant != nil {
		dupErr := errors.NewDuplicateError("ProductVariant", "sku", sku)
		observability.LogDuplicateError(op.Context(), "ProductVariant", "sku", sku)
		op.End(dupErr)
		return nil, dupErr
	}

	// Calculate final price (log business calculation)
	finalPrice := product.BasePrice + priceAdjustment
	op.AddAttribute("base_price", product.BasePrice)
	op.AddAttribute("price_adjustment", priceAdjustment)
	op.AddAttribute("final_price", finalPrice)

	// Determine active status (default to true if not specified)
	activeStatus := true
	if isActive != nil {
		activeStatus = *isActive
	}

	// Create variant entity
	variant := &entities.ProductVariant{
		ID:              uuid.New(),
		ProductID:       productID,
		SKU:             sku,
		Attributes:      attributes,
		StockQuantity:   stockQuantity,
		PriceAdjustment: priceAdjustment,
		IsActive:        activeStatus,
	}

	op.AddAttribute("variant_id", variant.ID.String())
	op.AddAttribute("is_active", activeStatus)

	// Persist variant
	err = s.variantRepo.Create(op.Context(), variant)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Record business event
	op.RecordBusinessEvent("product_variant", "variant_created",
		zap.String("variant_id", variant.ID.String()),
		zap.String("product_id", productID.String()),
		zap.String("sku", sku),
		zap.Float64("final_price", finalPrice),
	)

	op.End(nil)
	return variant, nil
}

// UpdateVariant updates an existing product variant
func (s *productVariantService) UpdateVariant(
	ctx context.Context,
	variantID uuid.UUID,
	sku *string,
	attributes *map[string]string,
	stockQuantity *int,
	priceAdjustment *float64,
	isActive *bool,
) (*entities.ProductVariant, error) {
	op := s.observer.StartOperation(ctx, "UpdateVariant")
	defer op.End(nil)

	op.AddAttribute("variant_id", variantID.String())

	// Find variant
	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		op.End(err)
		return nil, err
	}

	// Update SKU if provided
	if sku != nil {
		skuVal := strings.TrimSpace(*sku)
		if skuVal == "" {
			err := errors.NewValidationError("sku", "SKU cannot be empty")
			observability.LogValidationError(op.Context(), "sku", "SKU cannot be empty")
			op.End(err)
			return nil, err
		}

		if len(skuVal) < 3 || len(skuVal) > 100 {
			err := errors.NewValidationError("sku", "SKU must be between 3 and 100 characters")
			observability.LogValidationError(op.Context(), "sku", "SKU must be between 3 and 100 characters")
			op.End(err)
			return nil, err
		}

		// Check SKU uniqueness if changed
		if skuVal != variant.SKU {
			existingVariant, err := s.variantRepo.FindBySKU(op.Context(), skuVal)
			if err == nil && existingVariant != nil && existingVariant.ID != variantID {
				dupErr := errors.NewDuplicateError("ProductVariant", "sku", skuVal)
				observability.LogDuplicateError(op.Context(), "ProductVariant", "sku", skuVal)
				op.End(dupErr)
				return nil, dupErr
			}
		}

		variant.SKU = skuVal
		op.AddAttribute("updated_sku", skuVal)
	}

	// Update attributes if provided
	if attributes != nil {
		if len(*attributes) == 0 {
			err := errors.NewValidationError("attributes", "Attributes cannot be empty")
			observability.LogValidationError(op.Context(), "attributes", "Attributes cannot be empty")
			op.End(err)
			return nil, err
		}

		// Validate attribute keys and values
		for key, value := range *attributes {
			if strings.TrimSpace(key) == "" {
				err := errors.NewValidationError("attributes", "Attribute keys cannot be empty")
				observability.LogValidationError(op.Context(), "attributes", "Attribute keys cannot be empty")
				op.End(err)
				return nil, err
			}
			if strings.TrimSpace(value) == "" {
				err := errors.NewValidationError("attributes", "Attribute values cannot be empty")
				observability.LogValidationError(op.Context(), "attributes", "Attribute values cannot be empty")
				op.End(err)
				return nil, err
			}
		}

		variant.Attributes = *attributes
	}

	// Update stock quantity if provided
	if stockQuantity != nil {
		if *stockQuantity < 0 {
			err := errors.NewValidationError("stock_quantity", "Stock quantity must be >= 0")
			observability.LogValidationError(op.Context(), "stock_quantity", "Stock quantity must be >= 0")
			op.End(err)
			return nil, err
		}

		oldStock := variant.StockQuantity
		variant.StockQuantity = *stockQuantity
		op.LogInfo("Stock quantity updated",
			zap.Int("old_stock", oldStock),
			zap.Int("new_stock", *stockQuantity))
	}

	// Update price adjustment if provided
	if priceAdjustment != nil {
		variant.PriceAdjustment = *priceAdjustment
		op.AddAttribute("price_adjustment", *priceAdjustment)
	}

	// Update active status if provided
	if isActive != nil {
		if variant.IsActive != *isActive {
			observability.LogBusinessRule(op.Context(), "variant_status_change",
				fmt.Sprintf("Variant %s status changed from %v to %v", variant.SKU, variant.IsActive, *isActive))
		}
		variant.IsActive = *isActive
	}

	// Persist updates
	err = s.variantRepo.Update(op.Context(), variant)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Record business event
	op.RecordBusinessEvent("product_variant", "variant_updated",
		zap.String("variant_id", variant.ID.String()),
		zap.String("sku", variant.SKU),
	)

	op.End(nil)
	return variant, nil
}

// GetVariantsByProduct returns all variants for a product
func (s *productVariantService) GetVariantsByProduct(ctx context.Context, productID uuid.UUID) ([]*entities.ProductVariant, error) {
	op := s.observer.StartOperation(ctx, "GetVariantsByProduct")
	defer op.End(nil)

	op.AddAttribute("product_id", productID.String())

	// Validate product exists
	_, err := s.productRepo.GetByID(op.Context(), productID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Product", productID)
		op.End(err)
		return nil, err
	}

	// Get variants for product (repository should order by: is_active DESC, created_at ASC)
	variants, err := s.variantRepo.FindByProductID(op.Context(), productID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	op.AddAttribute("variants_count", len(variants))
	op.LogInfo("Retrieved variants for product",
		zap.Int("count", len(variants)))

	op.End(nil)
	return variants, nil
}

// UpdateStock updates stock quantity for a variant (admin only)
func (s *productVariantService) UpdateStock(ctx context.Context, variantID uuid.UUID, newStockQuantity int) (*entities.ProductVariant, error) {
	op := s.observer.StartOperation(ctx, "UpdateStock")
	defer op.End(nil)

	op.AddAttribute("variant_id", variantID.String())
	op.AddAttribute("new_stock", newStockQuantity)

	// Validate stock quantity
	if newStockQuantity < 0 {
		err := errors.NewValidationError("stock_quantity", "Stock quantity must be >= 0")
		observability.LogValidationError(op.Context(), "stock_quantity", "Stock quantity must be >= 0")
		op.End(err)
		return nil, err
	}

	// Find variant
	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		op.End(err)
		return nil, err
	}

	oldStock := variant.StockQuantity

	// Update stock atomically
	err = s.variantRepo.UpdateStock(op.Context(), variantID, newStockQuantity)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Fetch updated variant
	updatedVariant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Log stock change for audit trail
	op.LogInfo("Stock updated",
		zap.String("variant_sku", updatedVariant.SKU),
		zap.Int("old_stock", oldStock),
		zap.Int("new_stock", newStockQuantity),
		zap.Int("difference", newStockQuantity-oldStock),
	)

	// Record business event
	op.RecordBusinessEvent("product_variant", "stock_updated",
		zap.String("variant_id", variantID.String()),
		zap.String("variant_sku", updatedVariant.SKU),
		zap.Int("old_stock", oldStock),
		zap.Int("new_stock", newStockQuantity),
	)

	op.End(nil)
	return updatedVariant, nil
}

// DeleteVariant deletes a variant (soft delete if in use, hard delete otherwise)
func (s *productVariantService) DeleteVariant(ctx context.Context, variantID uuid.UUID) error {
	op := s.observer.StartOperation(ctx, "DeleteVariant")
	defer op.End(nil)

	op.AddAttribute("variant_id", variantID.String())

	// Find variant
	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		op.End(err)
		return err
	}

	// TODO: Check if variant is in any active cart/order
	// For Phase 2, we implement soft delete by default
	// Hard delete only for already inactive variants

	if variant.IsActive {
		// Soft delete: mark as inactive
		variant.IsActive = false
		err = s.variantRepo.Update(op.Context(), variant)
		if err != nil {
			op.End(err)
			return err
		}

		observability.LogBusinessRule(op.Context(), "soft_delete_variant",
			fmt.Sprintf("Variant %s soft deleted (marked as inactive)", variant.SKU))

		op.RecordBusinessEvent("product_variant", "variant_soft_deleted",
			zap.String("variant_id", variantID.String()),
			zap.String("variant_sku", variant.SKU),
		)

		op.LogInfo("Variant soft deleted (deactivated)",
			zap.String("variant_sku", variant.SKU))
	} else {
		// Hard delete: variant is already inactive and not referenced
		err = s.variantRepo.Delete(op.Context(), variantID)
		if err != nil {
			op.End(err)
			return err
		}

		op.RecordBusinessEvent("product_variant", "variant_hard_deleted",
			zap.String("variant_id", variantID.String()),
			zap.String("variant_sku", variant.SKU),
		)

		op.LogInfo("Variant hard deleted",
			zap.String("variant_sku", variant.SKU))
	}

	op.End(nil)
	return nil
}

// GetVariantByID returns a variant by ID
func (s *productVariantService) GetVariantByID(ctx context.Context, variantID uuid.UUID) (*entities.ProductVariant, error) {
	op := s.observer.StartOperation(ctx, "GetVariantByID")
	defer op.End(nil)

	op.AddAttribute("variant_id", variantID.String())

	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		op.End(err)
		return nil, err
	}

	op.AddAttribute("variant_sku", variant.SKU)
	op.End(nil)
	return variant, nil
}
