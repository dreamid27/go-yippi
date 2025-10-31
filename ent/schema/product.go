package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("description"),
		field.Float("price").Default(0),
        field.String("name").
            Default("unknown"),
		field.String("image_url"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
    }
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
