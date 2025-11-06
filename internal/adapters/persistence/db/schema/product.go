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
		field.String("sku").
			NotEmpty().
			Unique().
			Comment("Stock Keeping Unit"),
		field.String("slug").
			NotEmpty().
			Unique().
			Comment("URL-friendly identifier"),
		field.String("name").
			NotEmpty().
			Comment("Product name"),
		field.Float("price").
			Positive().
			Comment("Product price"),
		field.Text("description").
			Optional().
			Comment("Product description"),
		field.Int("weight").
			NonNegative().
			Default(0).
			Comment("Weight in grams for courier calculation"),
		field.Int("length").
			NonNegative().
			Default(0).
			Comment("Length in cm for courier calculation"),
		field.Int("width").
			NonNegative().
			Default(0).
			Comment("Width in cm for courier calculation"),
		field.Int("height").
			NonNegative().
			Default(0).
			Comment("Height in cm for courier calculation"),
		field.JSON("image_urls", []string{}).
			Optional().
			Comment("Access links to product images"),
		field.Enum("status").
			Values("draft", "published", "archived").
			Default("draft").
			Comment("Product status"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
