package dto

// AddCartItemRequest defines request body for adding an item to cart
type AddCartItemRequest struct {
	Body struct {
		ProductVariantID string `json:"product_variant_id" format:"uuid" minLength:"1" doc:"Product variant ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
		Quantity         int    `json:"quantity" minimum:"1" doc:"Quantity to add (minimum 1)" example:"2"`
	}
}

// UpdateCartItemRequest defines request body for updating cart item quantity
type UpdateCartItemRequest struct {
	ItemID string `path:"item_id" doc:"Cart item ID (UUID)"`
	Body   struct {
		Quantity int `json:"quantity" minimum:"0" doc:"New quantity (0 to delete item)" example:"3"`
	}
}

// RemoveCartItemRequest defines request for removing a cart item
type RemoveCartItemRequest struct {
	ItemID string `path:"item_id" doc:"Cart item ID (UUID)"`
}

// GuestCartItemDTO represents an item in guest's cart
type GuestCartItemDTO struct {
	ProductVariantID string `json:"product_variant_id" format:"uuid" minLength:"1" doc:"Product variant ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity         int    `json:"quantity" minimum:"1" doc:"Quantity (minimum 1)" example:"2"`
}

// MergeGuestCartRequest defines request body for merging guest cart
type MergeGuestCartRequest struct {
	Body struct {
		Items []GuestCartItemDTO `json:"items" minItems:"1" doc:"Guest cart items to merge"`
	}
}

// CartProductVariantDetail represents product variant information in cart responses
type CartProductVariantDetail struct {
	ID              string            `json:"id" doc:"Variant UUID" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID       string            `json:"product_id" doc:"Product UUID" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU             string            `json:"sku" doc:"Variant SKU" example:"TSHIRT-M-BLACK"`
	Attributes      map[string]string `json:"attributes" doc:"Variant attributes (e.g., size, color)" example:"{\"size\":\"M\",\"color\":\"Black\"}"`
	StockQuantity   int               `json:"stock_quantity" doc:"Available stock quantity" example:"50"`
	PriceAdjustment float64           `json:"price_adjustment" doc:"Price adjustment from base price" example:"0"`
	FinalPrice      float64           `json:"final_price" doc:"Final price (base + adjustment)" example:"100000"`
	IsActive        bool              `json:"is_active" doc:"Whether variant is active" example:"true"`
	IsInStock       bool              `json:"is_in_stock" doc:"Whether variant is in stock" example:"true"`
	CreatedAt       string            `json:"created_at" doc="Creation timestamp (ISO 8601)" example:"2026-01-01T10:00:00Z"`
	UpdatedAt       string            `json:"updated_at" doc="Last update timestamp (ISO 8601)" example:"2026-01-01T10:00:00Z"`
}

// ProductSummaryResponse represents minimal product data in cart responses
type ProductSummaryResponse struct {
	ID        string   `json:"id" doc:"Product ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string   `json:"name" doc:"Product name" example:"Cotton T-Shirt"`
	Slug      string   `json:"slug" doc:"URL-friendly identifier" example:"cotton-tshirt"`
	ImageURLs []string `json:"image_urls" doc:"Product image URLs" example:"[\"https://cdn.example.com/img1.jpg\"]"`
}

// CartItemResponse represents a cart item with product and variant details
type CartItemResponse struct {
	ID             string                      `json:"id" doc:"Cart item ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity       int                         `json:"quantity" doc:"Item quantity" example:"2"`
	PriceSnapshot  float64                     `json:"price_snapshot" doc:"Price at time of adding to cart" example:"100000"`
	Subtotal       float64                     `json:"subtotal" doc:"Item subtotal (price_snapshot * quantity)" example:"200000"`
	ProductVariant *CartProductVariantDetail    `json:"product_variant" doc:"Product variant details"`
	Product        *ProductSummaryResponse     `json:"product" doc:"Product summary"`
}

// CartResponse represents user's shopping cart
type CartResponse struct {
	Body struct {
		ID        string              `json:"id" doc:"Cart ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
		UserID    string              `json:"user_id" doc:"User ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
		Items     []*CartItemResponse `json:"items" doc:"Cart items"`
		Subtotal  float64             `json:"subtotal" doc:"Cart subtotal (sum of all items)" example:"500000"`
		ItemCount int                 `json:"item_count" doc:"Total number of items in cart" example:"5"`
	}
}

// StockValidationErrorResponse represents a stock validation error
type StockValidationErrorResponse struct {
	CartItemID string `json:"cart_item_id,omitempty" doc:"Cart item ID if applicable (UUID)"`
	VariantID  string `json:"variant_id" doc:"Variant ID (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
	Available  int    `json:"available" doc:"Available stock quantity" example:"5"`
	InCart     int    `json:"in_cart" doc:"Requested quantity in cart" example:"10"`
	Message    string `json:"message" doc:"Error message" example:"Insufficient stock for variant 'TSHIRT-XL-WHITE'"`
}

// CartItemResponseSingle defines the response for single cart item operations (add/update)
type CartItemResponseSingle struct {
	Body CartItemResponse `json:"body"`
}
