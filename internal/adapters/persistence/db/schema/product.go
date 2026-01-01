package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			StorageKey("id"),

		field.String("slug").
			NotEmpty().
			Unique().
			Comment("URL-friendly identifier"),

		field.String("name").
			NotEmpty().
			Comment("Product name"),

		field.Float("base_price").
			Positive().
			Comment("Base price for price calculation with variants"),

		field.Text("description").
			Optional().
			Comment("Product description"),

		field.Int("stock_quantity").
			NonNegative().
			Optional().
			Nillable().
			Comment("Stock for non-variant products, NULL for variant products"),

		field.Int("low_stock_threshold").
			NonNegative().
			Default(10).
			Comment("Alert when stock below this number"),

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

		// FK ke Category
		// Foreign keys as UUID
		field.UUID("category_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Category ID"),

		// FK ke Brand
		field.UUID("brand_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Brand ID"),

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
	return []ent.Edge{
		// Many-to-one ke Category (pakai kolom category_id, bukan join table category_products)
		edge.From("category", Category.Type).
			Ref("products").
			Unique().
			Field("category_id"),

		// Many-to-one ke Brand (pakai kolom brand_id)
		edge.From("brand", Brand.Type).
			Ref("products").
			Unique().
			Field("brand_id"),

		// One-to-many to ProductVariant
		edge.To("variants", ProductVariant.Type),

		// One-to-many to OrderItem (order items referencing this product)
		edge.To("order_items", OrderItem.Type),
	}
}
