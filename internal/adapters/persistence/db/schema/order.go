package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Order holds the schema definition for the Order entity.
type Order struct {
	ent.Schema
}

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("User ID - customer who placed the order"),

		field.UUID("shipping_address_id", uuid.UUID{}).
			Comment("Shipping address ID (reference to Address)"),

		field.String("order_number").
			Unique().
			NotEmpty().
			Comment("Human-readable order number: ORD-20250101-12345"),

		field.Enum("status").
			Values(
				"pending_payment",
				"payment_verified",
				"processing",
				"shipped",
				"delivered",
				"completed",
				"cancelled",
			).
			Default("pending_payment").
			Comment("Order status workflow"),

		// Shipping details
		field.String("courier").
			Optional().
			Comment("Courier name: JNE, J&T, SiCepat, Anteraja"),

		field.String("courier_service").
			Optional().
			Comment("Service type: OKE, REG, YES, CARGO"),

		field.Float("shipping_cost").
			Default(0).
			Comment("Shipping cost in IDR"),

		field.String("tracking_number").
			Optional().
			Comment("Courier tracking number (resi)"),

		// Order totals
		field.Float("subtotal").
			Positive().
			Comment("Sum of all items (price_snapshot * quantity)"),

		field.Float("discount").
			Default(0).
			Comment("Total discount amount (from vouchers/promos)"),

		field.Float("tax").
			Default(0).
			Comment("Tax amount (11% PPN in Indonesia)"),

		field.Float("total").
			Positive().
			Comment("Final total: subtotal + shipping_cost + tax - discount"),

		// Metadata
		field.String("notes").
			Optional().
			Comment("Customer order notes"),

		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional metadata: source_device, referrer, etc."),

		field.Time("cancelled_at").
			Optional().
			Comment("Order cancellation timestamp"),

		field.String("cancellation_reason").
			Optional().
			Comment("Reason for cancellation"),

		field.Time("shipped_at").
			Optional().
			Comment("Shipment timestamp"),

		field.Time("delivered_at").
			Optional().
			Comment("Delivery timestamp"),

		field.Time("completed_at").
			Optional().
			Comment("Order completion timestamp"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("orders").
			Unique().
			Required().
			Field("user_id"),

		edge.From("shipping_address", Address.Type).
			Ref("orders").
			Unique().
			Required().
			Field("shipping_address_id"),

		// TODO: Uncomment when OrderItem schema is created
		// edge.To("items", OrderItem.Type),

		// TODO: Uncomment when Payment schema is created
		// edge.To("payment", Payment.Type).
		// 	Unique(),
	}
}

// Indexes of the Order.
func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("order_number").Unique(),
		index.Fields("status"),
		index.Fields("created_at"),
		// Composite index for user's pending orders
		index.Fields("user_id", "status", "created_at"),
	}
}
