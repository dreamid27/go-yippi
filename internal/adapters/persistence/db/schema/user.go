package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password_hash").
			Sensitive(). // Hide in logs
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("phone").
			Optional(),
		field.Enum("role").
			Values("customer", "admin").
			Default("customer"),
		field.Bool("is_active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
