package entities

import (
	"time"

	"github.com/google/uuid"
)

// Brand represents a brand domain entity
type Brand struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
