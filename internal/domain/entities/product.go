package entities

import "time"

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusPublished ProductStatus = "published"
	ProductStatusArchived  ProductStatus = "archived"
)

// Product represents a product domain entity
type Product struct {
	ID          int
	SKU         string
	Slug        string
	Name        string
	Price       float64
	Description string
	Weight      int           // in grams
	Length      int           // in cm
	Width       int           // in cm
	Height      int           // in cm
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsValid checks if the product status is valid
func (p *Product) IsValid() bool {
	return p.Status == ProductStatusDraft ||
		p.Status == ProductStatusPublished ||
		p.Status == ProductStatusArchived
}

// IsPublished checks if the product is published
func (p *Product) IsPublished() bool {
	return p.Status == ProductStatusPublished
}
