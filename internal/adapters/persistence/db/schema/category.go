package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
		Default(uuid.New).
		StorageKey("id").
		Comment("Category unique identifier"),
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Category name"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		// Foreign key ke parent category (nullable untuk root)
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		// Setiap category boleh punya 1 parent
		edge.From("parent", Category.Type).
			Ref("children").
			Unique().
			Field("parent_id"),

		// Satu parent punya banyak children
		edge.To("children", Category.Type),

		// One-to-many relationship dengan Product
		edge.To("products", Product.Type),
	}
}
