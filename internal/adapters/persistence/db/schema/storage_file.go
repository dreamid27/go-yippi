package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// StorageFile holds the schema definition for the StorageFile entity.
type StorageFile struct {
	ent.Schema
}

// Fields of the StorageFile.
func (StorageFile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("filename").
			NotEmpty(),
		field.String("folder").
			NotEmpty(),
		field.String("original_filename").
			NotEmpty(),
		field.String("mime_type").
			NotEmpty(),
		field.Int64("file_size").
			Positive(),
		field.Bytes("file_data").
			NotEmpty(),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Default(map[string]interface{}{}),
		field.String("uploaded_by").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the StorageFile.
func (StorageFile) Edges() []ent.Edge {
	return nil
}

// Indexes of the StorageFile.
func (StorageFile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("folder"),
		index.Fields("filename"),
		index.Fields("created_at"),
	}
}
