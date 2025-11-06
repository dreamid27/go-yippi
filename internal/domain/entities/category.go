package entities

import "time"

// Category represents a product category domain entity
type Category struct {
	ID        int
	Name      string
	ParentID  *int // nullable for root categories
	CreatedAt time.Time
	UpdatedAt time.Time
}
