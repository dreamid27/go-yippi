package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrderItem holds the schema definition for the OrderItem entity.
type OrderItem struct {
	ent.Schema
}

// Fields of the OrderItem.
func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("order_id", uuid.UUID{}).
			Comment("Order ID"),

		field.UUID("product_id", uuid.UUID{}).
			Comment("Product ID (reference to Product)"),

		field.UUID("product_variant_id", uuid.UUID{}).
			Comment("Product Variant ID (reference to ProductVariant)"),

		field.String("product_name").
			NotEmpty().
			Comment("Product name at time of order (snapshot)"),

		field.JSON("product_variant_attributes", map[string]string{}).
			Comment("Variant attributes at time of order: {\"size\": \"M\", \"color\": \"Black\"}"),

		field.String("product_image_url").
			Optional().
			Comment("Main product image URL at time of order"),

		field.String("sku").
			NotEmpty().
			Comment("Variant SKU at time of order"),

		field.Int("quantity").
			Positive().
			Comment("Quantity ordered"),

		field.Float("price").
			Positive().
			Comment("Unit price at time of order (snapshot)"),

		field.Float("subtotal").
			Positive().
			Comment("Item subtotal: price * quantity"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OrderItem.
func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("items").
			Unique().
			Required().
			Field("order_id"),

		edge.From("product", Product.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("product_id"),

		edge.From("product_variant", ProductVariant.Type).
			Ref("order_items").
			Unique().
			Required().
			Field("product_variant_id"),
	}
}

// Indexes of the OrderItem.
func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("product_id"),
		index.Fields("product_variant_id"),
		index.Fields("order_id", "product_variant_id"),
	}
}
