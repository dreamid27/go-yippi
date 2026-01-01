package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Address holds the schema definition for the Address entity.
type Address struct {
	ent.Schema
}

// Fields of the Address.
func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("User ID - address owner"),

		field.String("label").
			NotEmpty().
			Comment("Address label: Home, Office, Parents, etc."),

		field.String("recipient_name").
			NotEmpty().
			Comment("Full name of recipient"),

		field.String("phone").
			NotEmpty().
			Comment("Recipient phone number (Indonesian format)"),

		field.String("address_line1").
			NotEmpty().
			Comment("Street address, building, floor, unit number"),

		field.String("address_line2").
			Optional().
			Comment("Additional address information"),

		field.String("province_id").
			NotEmpty().
			Comment("RajaOngkir province ID"),

		field.String("province_name").
			NotEmpty().
			Comment("Province name (cached from RajaOngkir)"),

		field.String("city_id").
			NotEmpty().
			Comment("RajaOngkir city ID"),

		field.String("city_name").
			NotEmpty().
			Comment("City name (cached from RajaOngkir)"),

		field.String("district").
			Optional().
			Comment("District/Sub-district (Kecamatan)"),

		field.String("postal_code").
			NotEmpty().
			Comment("Postal code"),

		field.Bool("is_default").
			Default(false).
			Comment("Is this default shipping address"),

		field.Bool("is_deleted").
			Default(false).
			Comment("Soft delete flag"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Address.
func (Address) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("addresses").
			Unique().
			Required().
			Field("user_id"),

		// TODO: Re-add after Order schema is created (Task 3)
		// edge.To("orders", Order.Type),
	}
}

// Indexes of the Address.
func (Address) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_default").Unique(), // Only one default per user
	}
}
