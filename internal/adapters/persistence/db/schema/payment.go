package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Payment holds the schema definition for the Payment entity.
type Payment struct {
	ent.Schema
}

// Fields of the Payment.
func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("order_id", uuid.UUID{}).
			Comment("Order ID (one-to-one with Order)"),

		field.Enum("method").
			Values(
				"va_bca",
				"va_mandiri",
				"va_bni",
				"va_bri",
				"gopay",
				"ovo",
				"dana",
				"shopeepay",
				"qris",
				"credit_card",
				"cod",
			).
			Comment("Payment method"),

		field.Enum("status").
			Values(
				"pending",
				"processing",
				"success",
				"failed",
				"expired",
				"cancelled",
			).
			Default("pending").
			Comment("Payment status from Midtrans"),

		field.Float("amount").
			Positive().
			Comment("Payment amount (should match order.total)"),

		field.String("midtrans_transaction_id").
			Optional().
			Comment("Midtrans transaction ID"),

		field.String("midtrans_payment_type").
			Optional().
			Comment("Midtrans payment type: bank_transfer, e_wallet, qris, credit_card"),

		field.String("va_number").
			Optional().
			Comment("Virtual Account number (for VA payments)"),

		field.String("va_bank").
			Optional().
			Comment("Virtual Account bank: BCA, Mandiri, BNI, BRI"),

		field.String("payment_url").
			Optional().
			Comment("Midtrans Snap payment URL or redirect URL"),

		field.Time("expires_at").
			Optional().
			Comment("Payment expiration timestamp (24h from creation)"),

		field.Time("paid_at").
			Optional().
			Comment("Successful payment timestamp"),

		field.String("failure_reason").
			Optional().
			Comment("Reason for payment failure (from Midtrans)"),

		field.JSON("raw_midtrans_response", map[string]interface{}{}).
			Optional().
			Comment("Raw Midtrans webhook/response data for debugging"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Payment.
func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("payment").
			Unique().
			Required().
			Field("order_id"),
	}
}

// Indexes of the Payment.
func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id").Unique(),
		index.Fields("midtrans_transaction_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
	}
}
