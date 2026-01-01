package services

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/ports"
	"example.com/go-yippi/internal/infrastructure/observability"
)

// priceCalculatorService implements the PriceCalculatorService interface
type priceCalculatorService struct {
	observer *observability.ServiceObserver
}

// NewPriceCalculatorService creates a new price calculator service
func NewPriceCalculatorService() ports.PriceCalculatorService {
	return &priceCalculatorService{
		observer: observability.NewServiceObserver("PriceCalculatorService"),
	}
}

// CalculateVariantPrice calculates final price for a variant
// Formula: finalPrice = basePrice + priceAdjustment
func (s *priceCalculatorService) CalculateVariantPrice(ctx context.Context, basePrice float64, priceAdjustment float64) float64 {
	op := s.observer.StartOperation(ctx, "CalculateVariantPrice")
	defer op.End(nil)

	op.AddAttribute("base_price", basePrice)
	op.AddAttribute("price_adjustment", priceAdjustment)

	finalPrice := basePrice + priceAdjustment

	op.AddAttribute("final_price", finalPrice)
	op.End(nil)

	return finalPrice
}

// CalculateCartSubtotal calculates total price for cart items
// Formula: subtotal = sum(item.price_snapshot * item.quantity)
func (s *priceCalculatorService) CalculateCartSubtotal(ctx context.Context, cartItems []*entities.CartItem) float64 {
	op := s.observer.StartOperation(ctx, "CalculateCartSubtotal")
	defer op.End(nil)

	op.AddAttribute("items_count", len(cartItems))

	if len(cartItems) == 0 {
		op.LogInfo("Empty cart, returning 0.0 subtotal")
		op.End(nil)
		return 0.0
	}

	var subtotal float64
	for _, item := range cartItems {
		itemSubtotal := item.PriceSnapshot * float64(item.Quantity)
		subtotal += itemSubtotal
	}

	op.AddAttribute("subtotal", subtotal)
	op.RecordBusinessEvent("cart", "cart_subtotal_calculated")

	op.End(nil)
	return subtotal
}

// CalculateItemSubtotal calculates subtotal for a single item
// Formula: itemSubtotal = priceSnapshot * quantity
func (s *priceCalculatorService) CalculateItemSubtotal(ctx context.Context, priceSnapshot float64, quantity int) float64 {
	op := s.observer.StartOperation(ctx, "CalculateItemSubtotal")
	defer op.End(nil)

	op.AddAttribute("price_snapshot", priceSnapshot)
	op.AddAttribute("quantity", quantity)

	itemSubtotal := priceSnapshot * float64(quantity)

	op.AddAttribute("item_subtotal", itemSubtotal)
	op.End(nil)

	return itemSubtotal
}

// GetPriceRange returns min and max prices from variants
func (s *priceCalculatorService) GetPriceRange(ctx context.Context, variants []*entities.ProductVariant, basePrice float64) (minPrice float64, maxPrice float64) {
	op := s.observer.StartOperation(ctx, "GetPriceRange")
	defer op.End(nil)

	op.AddAttribute("variants_count", len(variants))
	op.AddAttribute("base_price", basePrice)

	if len(variants) == 0 {
		op.LogInfo("No variants, returning base price as both min and max")
		op.AddAttribute("min_price", basePrice)
		op.AddAttribute("max_price", basePrice)
		op.End(nil)
		return basePrice, basePrice
	}

	// Initialize with base price
	minPrice = basePrice
	maxPrice = basePrice
	foundActive := false

	// Check all active variants
	for _, variant := range variants {
		if !variant.IsActive {
			continue
		}

		finalPrice := s.CalculateVariantPrice(op.Context(), basePrice, variant.PriceAdjustment)

		if !foundActive {
			minPrice = finalPrice
			maxPrice = finalPrice
			foundActive = true
		} else {
			if finalPrice < minPrice {
				minPrice = finalPrice
			}
			if finalPrice > maxPrice {
				maxPrice = finalPrice
			}
		}
	}

	// If no active variants found, use base price
	if !foundActive {
		op.LogInfo("No active variants found, using base price")
		minPrice = basePrice
		maxPrice = basePrice
	}

	op.AddAttribute("min_price", minPrice)
	op.AddAttribute("max_price", maxPrice)
	op.End(nil)

	return minPrice, maxPrice
}
