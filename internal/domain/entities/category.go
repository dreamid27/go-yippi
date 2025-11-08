package entities

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category domain entity
type Category struct {
	ID        uuid.UUID
	Name      string
	ParentID  *uuid.UUID // nullable for root categories
	CreatedAt time.Time
	UpdatedAt time.Time
}
