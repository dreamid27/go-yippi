package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// RefreshToken holds the schema definition for the RefreshToken entity.
type RefreshToken struct {
	ent.Schema
}

// Fields of the RefreshToken.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("token").
			Unique().
			NotEmpty().
			Sensitive(), // Hide in logs
		field.Time("expires_at").
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("revoked_at").
			Optional().
			Nillable(),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("refresh_tokens").
			Unique().
			Required(),
	}
}
