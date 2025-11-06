package entities

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	ImageURLs   []string      // access links to product images
	Status      ProductStatus
	BrandID     *uuid.UUID    // optional brand association
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

// GenerateSlug creates a URL-friendly slug from a given string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove any characters that are not alphanumeric or hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Replace multiple consecutive hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
