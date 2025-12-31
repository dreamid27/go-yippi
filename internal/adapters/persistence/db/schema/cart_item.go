package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CartItem holds the schema definition for the CartItem entity.
type CartItem struct {
	ent.Schema
}

// Fields of the CartItem.
func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("cart_id", uuid.UUID{}).
			Comment("Cart ID"),

		field.UUID("product_variant_id", uuid.UUID{}).
			Comment("Product Variant ID"),

		field.Int("quantity").
			Positive().
			Default(1).
			Comment("Quantity of this variant in cart"),

		field.Float("price_snapshot").
			Positive().
			Comment("Price snapshot at time of adding to cart"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the CartItem.
func (CartItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cart", Cart.Type).
			Ref("items").
			Unique().
			Required().
			Field("cart_id"),

		edge.From("product_variant", ProductVariant.Type).
			Ref("cart_items").
			Unique().
			Required().
			Field("product_variant_id"),
	}
}

// Indexes of the CartItem.
func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id"),
		index.Fields("product_variant_id"),
		// Prevent duplicate variant in same cart
		index.Fields("cart_id", "product_variant_id").Unique(),
	}
}
