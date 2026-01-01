package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/infrastructure/logging"
)

func init() {
	// Initialize logging for tests
	logging.Initialize(logging.DefaultLogConfig())
}

func TestCalculateVariantPrice(t *testing.T) {
	service := NewPriceCalculatorService()

	tests := []struct {
		name            string
		basePrice       float64
		priceAdjustment float64
		expected        float64
	}{
		{
			name:            "zero adjustment",
			basePrice:       100000,
			priceAdjustment: 0,
			expected:        100000,
		},
		{
			name:            "positive adjustment",
			basePrice:       100000,
			priceAdjustment: 10000,
			expected:        110000,
		},
		{
			name:            "negative adjustment",
			basePrice:       100000,
			priceAdjustment: -5000,
			expected:        95000,
		},
		{
			name:            "zero base price",
			basePrice:       0,
			priceAdjustment: 10000,
			expected:        10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := service.CalculateVariantPrice(ctx, tt.basePrice, tt.priceAdjustment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCartSubtotal(t *testing.T) {
	service := NewPriceCalculatorService()

	tests := []struct {
		name     string
		items    []*entities.CartItem
		expected float64
	}{
		{
			name:     "empty cart",
			items:    []*entities.CartItem{},
			expected: 0.0,
		},
		{
			name: "single item",
			items: []*entities.CartItem{
				{
					Quantity:      2,
					PriceSnapshot: 100000,
				},
			},
			expected: 200000,
		},
		{
			name: "multiple items",
			items: []*entities.CartItem{
				{
					Quantity:      2,
					PriceSnapshot: 100000,
				},
				{
					Quantity:      1,
					PriceSnapshot: 50000,
				},
				{
					Quantity:      3,
					PriceSnapshot: 75000,
				},
			},
			expected: 475000,
		},
		{
			name: "zero quantity",
			items: []*entities.CartItem{
				{
					Quantity:      0,
					PriceSnapshot: 100000,
				},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := service.CalculateCartSubtotal(ctx, tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateItemSubtotal(t *testing.T) {
	service := NewPriceCalculatorService()

	tests := []struct {
		name          string
		priceSnapshot float64
		quantity      int
		expected      float64
	}{
		{
			name:          "single item",
			priceSnapshot: 100000,
			quantity:      1,
			expected:      100000,
		},
		{
			name:          "multiple quantity",
			priceSnapshot: 100000,
			quantity:      5,
			expected:      500000,
		},
		{
			name:          "zero quantity",
			priceSnapshot: 100000,
			quantity:      0,
			expected:      0.0,
		},
		{
			name:          "zero price",
			priceSnapshot: 0,
			quantity:      5,
			expected:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := service.CalculateItemSubtotal(ctx, tt.priceSnapshot, tt.quantity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPriceRange(t *testing.T) {
	service := NewPriceCalculatorService()

	tests := []struct {
		name       string
		variants   []*entities.ProductVariant
		basePrice  float64
		minPrice   float64
		maxPrice   float64
	}{
		{
			name:      "no variants",
			variants:  []*entities.ProductVariant{},
			basePrice: 100000,
			minPrice:  100000,
			maxPrice:  100000,
		},
		{
			name: "single active variant",
			variants: []*entities.ProductVariant{
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: 10000,
				},
			},
			basePrice: 100000,
			minPrice:  110000,
			maxPrice:  110000,
		},
		{
			name: "multiple active variants",
			variants: []*entities.ProductVariant{
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: -5000,
				},
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: 10000,
				},
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: 5000,
				},
			},
			basePrice: 100000,
			minPrice:  95000,
			maxPrice:  110000,
		},
		{
			name: "mix of active and inactive variants",
			variants: []*entities.ProductVariant{
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: -5000,
				},
				{
					ID:              uuid.New(),
					IsActive:        false,
					PriceAdjustment: 20000,
				},
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: 10000,
				},
			},
			basePrice: 100000,
			minPrice:  95000,
			maxPrice:  110000,
		},
		{
			name: "all inactive variants",
			variants: []*entities.ProductVariant{
				{
					ID:              uuid.New(),
					IsActive:        false,
					PriceAdjustment: -5000,
				},
				{
					ID:              uuid.New(),
					IsActive:        false,
					PriceAdjustment: 10000,
				},
			},
			basePrice: 100000,
			minPrice:  100000,
			maxPrice:  100000,
		},
		{
			name: "negative and positive adjustments",
			variants: []*entities.ProductVariant{
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: -10000,
				},
				{
					ID:              uuid.New(),
					IsActive:        true,
					PriceAdjustment: 20000,
				},
			},
			basePrice: 100000,
			minPrice:  90000,
			maxPrice:  120000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			min, max := service.GetPriceRange(ctx, tt.variants, tt.basePrice)
			assert.Equal(t, tt.minPrice, min)
			assert.Equal(t, tt.maxPrice, max)
		})
	}
}
