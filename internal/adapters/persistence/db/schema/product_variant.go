package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProductVariant holds the schema definition for the ProductVariant entity.
type ProductVariant struct {
	ent.Schema
}

// Fields of the ProductVariant.
func (ProductVariant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.String("sku").
			NotEmpty().
			Unique().
			Comment("Variant SKU (unique across all variants)"),

		field.JSON("attributes", map[string]string{}).
			Comment("Variant attributes: {\"size\": \"M\", \"color\": \"Red\"}"),

		field.Int("stock_quantity").
			NonNegative().
			Default(0).
			Comment("Stock quantity for this variant"),

		field.Float("price_adjustment").
			Default(0).
			Comment("Price adjustment from base_price (can be negative)"),

		field.Bool("is_active").
			Default(true).
			Comment("Is this variant available for sale"),

		field.UUID("product_id", uuid.UUID{}).
			Comment("Product ID"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ProductVariant.
func (ProductVariant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("variants").
			Unique().
			Required().
			Field("product_id"),

		edge.To("cart_items", CartItem.Type),

		// One-to-many to OrderItem (order items referencing this variant)
		edge.To("order_items", OrderItem.Type),
	}
}

// Indexes of the ProductVariant.
func (ProductVariant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("sku").Unique(),
	}
}
