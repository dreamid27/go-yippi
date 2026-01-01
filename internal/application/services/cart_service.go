package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"example.com/go-yippi/internal/infrastructure/observability"
)

// CartService handles shopping cart business logic
type CartService struct {
	cartRepo          ports.CartRepository
	variantRepo       ports.ProductVariantRepository
	productRepo       ports.ProductRepository
	stockValidator    ports.StockValidatorService
	calcService       ports.PriceCalculatorService
	observer          *observability.ServiceObserver
}

// NewCartService creates a new CartService
func NewCartService(
	cartRepo ports.CartRepository,
	variantRepo ports.ProductVariantRepository,
	productRepo ports.ProductRepository,
	stockValidator ports.StockValidatorService,
	calcService ports.PriceCalculatorService,
) *CartService {
	return &CartService{
		cartRepo:       cartRepo,
		variantRepo:    variantRepo,
		productRepo:    productRepo,
		stockValidator: stockValidator,
		calcService:    calcService,
		observer:       observability.NewServiceObserver("CartService"),
	}
}

// GetOrCreateCart gets or creates a cart for a user
func (s *CartService) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (*ports.CartDetail, error) {
	op := s.observer.StartOperation(ctx, "GetOrCreateCart")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())

	// Get or create cart
	cart, err := s.cartRepo.GetOrCreate(op.Context(), userID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Get cart items
	items, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Build cart detail with enriched data
	cartDetail, err := s.buildCartDetail(op.Context(), cart, items)
	if err != nil {
		op.End(err)
		return nil, err
	}

	op.LogInfo("Cart retrieved", zap.Int("item_count", cartDetail.ItemCount), zap.Float64("subtotal", cartDetail.Subtotal))
	op.End(nil)
	return cartDetail, nil
}

// AddItemToCart adds a product variant to user's cart
func (s *CartService) AddItemToCart(
	ctx context.Context,
	userID uuid.UUID,
	variantID uuid.UUID,
	quantity int,
) (*entities.CartItem, error) {
	op := s.observer.StartOperation(ctx, "AddItemToCart")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())
	op.AddAttribute("variant_id", variantID.String())
	op.AddAttribute("quantity", quantity)

	// Validate quantity
	if quantity < 1 {
		err := errors.NewValidationError("quantity", "Quantity must be at least 1")
		observability.LogValidationError(op.Context(), "quantity", "Quantity must be at least 1")
		op.End(err)
		return nil, err
	}

	// Validate variant exists and is active
	variant, err := s.variantRepo.FindByID(op.Context(), variantID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "ProductVariant", variantID)
		err = errors.NewNotFoundError("ProductVariant", variantID)
		op.End(err)
		return nil, err
	}

	if !variant.IsActive {
		err := errors.NewValidationError("variant", "Variant is not active")
		observability.LogBusinessRule(op.Context(), "add_active_variant_only", "Cannot add inactive variant to cart")
		op.End(err)
		return nil, err
	}

	// Check stock availability
	err = s.stockValidator.ValidateStock(op.Context(), variantID, quantity)
	if err != nil {
		observability.LogBusinessRule(op.Context(), "stock_sufficient", err.Error())
		op.End(err)
		return nil, err
	}

	// Get user's cart
	cart, err := s.cartRepo.GetOrCreate(op.Context(), userID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Check if variant already exists in cart
	existingItem, err := s.cartRepo.GetCartItemByVariant(op.Context(), cart.ID, variantID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Get product for price calculation
	product, err := s.productRepo.GetByID(op.Context(), variant.ProductID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Calculate final price for price snapshot
	finalPrice := s.calcService.CalculateVariantPrice(op.Context(), product.BasePrice, variant.PriceAdjustment)

	if existingItem != nil {
		// Update existing item quantity
		newQuantity := existingItem.Quantity + quantity

		// Validate stock for new total quantity
		if err := s.stockValidator.ValidateStock(op.Context(), variantID, newQuantity); err != nil {
			observability.LogBusinessRule(op.Context(), "stock_validation_update",
				fmt.Sprintf("Cannot add %d more items. %s", quantity, err.Error()))
			op.End(err)
			return nil, err
		}

		err = s.cartRepo.UpdateItemQuantity(op.Context(), existingItem.ID, newQuantity)
		if err != nil {
			op.End(err)
			return nil, err
		}

		// Fetch updated item
		items, err := s.cartRepo.GetItems(op.Context(), cart.ID)
		if err != nil {
			op.End(err)
			return nil, err
		}

		var updatedItem *entities.CartItem
		for _, item := range items {
			if item.ID == existingItem.ID {
				updatedItem = item
				break
			}
		}

		op.RecordBusinessEvent("cart", "item_quantity_increased",
			zap.String("cart_id", cart.ID.String()),
			zap.String("variant_id", variantID.String()),
			zap.Int("old_quantity", existingItem.Quantity),
			zap.Int("new_quantity", newQuantity),
		)

		op.End(nil)
		return updatedItem, nil
	}

	// Create new cart item
	cartItem := &entities.CartItem{
		ID:               uuid.New(),
		CartID:           cart.ID,
		ProductVariantID: variantID,
		Quantity:         quantity,
		PriceSnapshot:    finalPrice,
	}

	err = s.cartRepo.AddItem(op.Context(), cartItem)
	if err != nil {
		op.End(err)
		return nil, err
	}

	op.LogInfo("Added item to cart",
		zap.String("cart_item_id", cartItem.ID.String()),
		zap.String("variant_id", variantID.String()),
		zap.Int("quantity", quantity),
		zap.Float64("price_snapshot", finalPrice),
	)

	op.RecordBusinessEvent("cart", "cart_item_added",
		zap.String("cart_id", cart.ID.String()),
		zap.String("variant_id", variantID.String()),
		zap.Int("quantity", quantity),
		zap.Float64("price_snapshot", finalPrice),
	)

	op.End(nil)
	return cartItem, nil
}

// UpdateCartItemQuantity updates quantity of a cart item
func (s *CartService) UpdateCartItemQuantity(
	ctx context.Context,
	userID uuid.UUID,
	cartItemID uuid.UUID,
	newQuantity int,
) (*entities.CartItem, error) {
	op := s.observer.StartOperation(ctx, "UpdateCartItemQuantity")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())
	op.AddAttribute("cart_item_id", cartItemID.String())
	op.AddAttribute("new_quantity", newQuantity)

	// Validate quantity
	if newQuantity < 0 {
		err := errors.NewValidationError("quantity", "Quantity must be >= 0")
		observability.LogValidationError(op.Context(), "quantity", "Quantity must be >= 0")
		op.End(err)
		return nil, err
	}

	// If quantity is 0, delete item
	if newQuantity == 0 {
		err := s.RemoveCartItem(op.Context(), userID, cartItemID)
		if err != nil {
			op.End(err)
			return nil, err
		}
		op.End(nil)
		return nil, nil
	}

	// Get cart
	cart, err := s.cartRepo.FindByUserID(op.Context(), userID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Cart", userID)
		err = errors.NewNotFoundError("Cart", userID)
		op.End(err)
		return nil, err
	}

	// Get cart items
	items, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Find cart item and verify ownership
	var cartItem *entities.CartItem
	for _, item := range items {
		if item.ID == cartItemID {
			cartItem = item
			break
		}
	}

	if cartItem == nil {
		err := errors.NewNotFoundError("CartItem", cartItemID)
		observability.LogNotFoundError(op.Context(), "CartItem", cartItemID)
		op.End(err)
		return nil, err
	}

	// Authorization check: ensure cart item belongs to user's cart
	if cartItem.CartID != cart.ID {
		err := fmt.Errorf("cart item does not belong to user")
		observability.LogBusinessRule(op.Context(), "cart_item_ownership",
			fmt.Sprintf("User %s attempted to modify cart item %s not in their cart", userID, cartItemID))
		op.End(err)
		return nil, err
	}

	// Validate stock for new quantity
	err = s.stockValidator.ValidateStock(op.Context(), cartItem.ProductVariantID, newQuantity)
	if err != nil {
		observability.LogBusinessRule(op.Context(), "stock_sufficient", err.Error())
		op.End(err)
		return nil, err
	}

	// Update quantity
	oldQuantity := cartItem.Quantity
	err = s.cartRepo.UpdateItemQuantity(op.Context(), cartItemID, newQuantity)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Fetch updated item
	updatedItems, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	var updatedItem *entities.CartItem
	for _, item := range updatedItems {
		if item.ID == cartItemID {
			updatedItem = item
			break
		}
	}

	op.LogInfo("Updated cart item quantity",
		zap.String("cart_item_id", cartItemID.String()),
		zap.Int("old_quantity", oldQuantity),
		zap.Int("new_quantity", newQuantity),
	)

	op.RecordBusinessEvent("cart", "cart_item_quantity_updated",
		zap.String("cart_id", cart.ID.String()),
		zap.String("variant_id", cartItem.ProductVariantID.String()),
		zap.Int("old_quantity", oldQuantity),
		zap.Int("new_quantity", newQuantity),
	)

	op.End(nil)
	return updatedItem, nil
}

// RemoveCartItem removes an item from cart
func (s *CartService) RemoveCartItem(ctx context.Context, userID uuid.UUID, cartItemID uuid.UUID) error {
	op := s.observer.StartOperation(ctx, "RemoveCartItem")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())
	op.AddAttribute("cart_item_id", cartItemID.String())

	// Get cart
	cart, err := s.cartRepo.FindByUserID(op.Context(), userID)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Cart", userID)
		err = errors.NewNotFoundError("Cart", userID)
		op.End(err)
		return err
	}

	// Get cart items
	items, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return err
	}

	// Find cart item and verify ownership
	var cartItem *entities.CartItem
	for _, item := range items {
		if item.ID == cartItemID {
			cartItem = item
			break
		}
	}

	if cartItem == nil {
		err := errors.NewNotFoundError("CartItem", cartItemID)
		observability.LogNotFoundError(op.Context(), "CartItem", cartItemID)
		op.End(err)
		return err
	}

	// Authorization check: ensure cart item belongs to user's cart
	if cartItem.CartID != cart.ID {
		err := fmt.Errorf("cart item does not belong to user")
		observability.LogBusinessRule(op.Context(), "cart_item_ownership",
			fmt.Sprintf("User %s attempted to remove cart item %s not in their cart", userID, cartItemID))
		op.End(err)
		return err
	}

	// Delete item
	err = s.cartRepo.RemoveItem(op.Context(), cartItemID)
	if err != nil {
		op.End(err)
		return err
	}

	op.LogInfo("Removed item from cart",
		zap.String("cart_item_id", cartItemID.String()),
		zap.String("variant_id", cartItem.ProductVariantID.String()),
	)

	op.RecordBusinessEvent("cart", "cart_item_removed",
		zap.String("cart_id", cart.ID.String()),
		zap.String("variant_id", cartItem.ProductVariantID.String()),
	)

	op.End(nil)
	return nil
}

// ClearCart removes all items from user's cart
func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	op := s.observer.StartOperation(ctx, "ClearCart")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())

	// Get cart
	cart, err := s.cartRepo.GetOrCreate(op.Context(), userID)
	if err != nil {
		op.End(err)
		return err
	}

	// Clear all items
	err = s.cartRepo.ClearCart(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return err
	}

	op.LogInfo("Cart cleared", zap.String("cart_id", cart.ID.String()))
	op.RecordBusinessEvent("cart", "cart_cleared", zap.String("cart_id", cart.ID.String()))
	op.End(nil)
	return nil
}

// MergeGuestCart merges guest cart items into user's cart
func (s *CartService) MergeGuestCart(ctx context.Context, userID uuid.UUID, guestItems []ports.GuestCartItem) (*ports.CartDetail, error) {
	op := s.observer.StartOperation(ctx, "MergeGuestCart")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())
	op.AddAttribute("guest_items_count", len(guestItems))

	if len(guestItems) == 0 {
		// No items to merge, just return current cart
		return s.GetOrCreateCart(op.Context(), userID)
	}

	// Validate all guest items and collect variants
	variantsMap := make(map[uuid.UUID]*entities.ProductVariant)
	for _, guestItem := range guestItems {
		variant, err := s.variantRepo.FindByID(op.Context(), guestItem.ProductVariantID)
		if err != nil {
			observability.LogNotFoundError(op.Context(), "ProductVariant", guestItem.ProductVariantID)
			err = errors.NewNotFoundError("ProductVariant", guestItem.ProductVariantID)
			op.End(err)
			return nil, err
		}

		// Validate quantity
		if guestItem.Quantity <= 0 {
			err := errors.NewValidationError("quantity",
				fmt.Sprintf("Invalid quantity %d for variant %s", guestItem.Quantity, guestItem.ProductVariantID))
			observability.LogValidationError(op.Context(), "quantity", err.Error())
			op.End(err)
			return nil, err
		}

		variantsMap[guestItem.ProductVariantID] = variant
	}

	// Get or create cart
	cart, err := s.cartRepo.GetOrCreate(op.Context(), userID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Get existing cart items
	existingItems, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Create a map for quick lookup of existing items
	existingItemsMap := make(map[uuid.UUID]*entities.CartItem)
	for _, item := range existingItems {
		existingItemsMap[item.ProductVariantID] = item
	}

	// Merge guest items
	var mergedCount int

	for _, guestItem := range guestItems {
		variant := variantsMap[guestItem.ProductVariantID]

		// Check if variant already exists in cart
		if existingItem, exists := existingItemsMap[guestItem.ProductVariantID]; exists {
			// Sum quantities
			newQuantity := existingItem.Quantity + guestItem.Quantity

			// Validate stock for merged quantity
			if err := s.stockValidator.ValidateStock(op.Context(), guestItem.ProductVariantID, newQuantity); err != nil {
				observability.LogBusinessRule(op.Context(), "stock_validation_merge",
					fmt.Sprintf("Insufficient stock when merging variant %s: %s", variant.ID, err.Error()))
				op.End(err)
				return nil, err
			}

			// Update quantity
			err = s.cartRepo.UpdateItemQuantity(op.Context(), existingItem.ID, newQuantity)
			if err != nil {
				op.End(err)
				return nil, err
			}
			mergedCount++
		} else {
			// Validate stock for new item
			if err := s.stockValidator.ValidateStock(op.Context(), guestItem.ProductVariantID, guestItem.Quantity); err != nil {
				observability.LogBusinessRule(op.Context(), "stock_validation_new", err.Error())
				op.End(err)
				return nil, err
			}

			// Get product for price calculation
			product, err := s.productRepo.GetByID(op.Context(), variant.ProductID)
			if err != nil {
				op.End(err)
				return nil, err
			}

			// Calculate final price for price snapshot
			finalPrice := s.calcService.CalculateVariantPrice(op.Context(), product.BasePrice, variant.PriceAdjustment)

			// Add new item
			newItem := &entities.CartItem{
				ID:               uuid.New(),
				CartID:           cart.ID,
				ProductVariantID: guestItem.ProductVariantID,
				Quantity:         guestItem.Quantity,
				PriceSnapshot:    finalPrice,
			}

			err = s.cartRepo.AddItem(op.Context(), newItem)
			if err != nil {
				op.End(err)
				return nil, err
			}
			mergedCount++
		}
	}

	op.RecordBusinessEvent("cart", "guest_cart_merged",
		zap.String("cart_id", cart.ID.String()),
		zap.Int("items_merged", mergedCount),
	)

	// Return updated cart with details
	cartDetail, err := s.GetCartWithDetails(op.Context(), userID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	op.End(nil)
	return cartDetail, nil
}

// GetCartWithDetails gets cart with all items and details
func (s *CartService) GetCartWithDetails(ctx context.Context, userID uuid.UUID) (*ports.CartDetail, error) {
	op := s.observer.StartOperation(ctx, "GetCartWithDetails")
	defer op.End(nil)

	op.AddAttribute("user_id", userID.String())

	// Get or create cart
	cart, err := s.cartRepo.GetOrCreate(op.Context(), userID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Get cart items
	items, err := s.cartRepo.GetItems(op.Context(), cart.ID)
	if err != nil {
		op.End(err)
		return nil, err
	}

	// Build cart detail with enriched items
	cartDetail, err := s.buildCartDetail(op.Context(), cart, items)
	if err != nil {
		op.End(err)
		return nil, err
	}

	op.LogInfo("Cart with details retrieved",
		zap.String("cart_id", cartDetail.ID.String()),
		zap.Int("item_count", cartDetail.ItemCount),
		zap.Float64("subtotal", cartDetail.Subtotal),
	)

	op.End(nil)
	return cartDetail, nil
}

// buildCartDetail builds a CartDetail from cart and items by enriching with product/variant data
func (s *CartService) buildCartDetail(ctx context.Context, cart *entities.Cart, items []*entities.CartItem) (*ports.CartDetail, error) {
	cartItemDetails := make([]*ports.CartItemDetail, 0, len(items))

	for _, item := range items {
		// Get variant
		variant, err := s.variantRepo.FindByID(ctx, item.ProductVariantID)
		if err != nil {
			return nil, err
		}

		// Get product
		product, err := s.productRepo.GetByID(ctx, variant.ProductID)
		if err != nil {
			return nil, err
		}

		// Calculate item subtotal
		itemSubtotal := s.calcService.CalculateItemSubtotal(ctx, item.PriceSnapshot, item.Quantity)

		cartItemDetail := &ports.CartItemDetail{
			ID:             item.ID,
			Quantity:       item.Quantity,
			PriceSnapshot:  item.PriceSnapshot,
			ItemSubtotal:   itemSubtotal,
			ProductVariant: variant,
			Product:        product,
		}

		cartItemDetails = append(cartItemDetails, cartItemDetail)
	}

	// Calculate totals
	subtotal := s.calcService.CalculateCartSubtotal(ctx, items)
	itemCount := 0
	for _, item := range items {
		itemCount += item.Quantity
	}

	return &ports.CartDetail{
		ID:        cart.ID,
		UserID:    cart.UserID,
		Items:     cartItemDetails,
		Subtotal:  subtotal,
		ItemCount: itemCount,
	}, nil
}
