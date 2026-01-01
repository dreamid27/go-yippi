package entities

import (
	"time"

	"github.com/google/uuid"
)

// Address represents a user's shipping/billing address
type Address struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Label          string
	RecipientName  string
	Phone          string
	AddressLine1   string
	AddressLine2   string
	ProvinceID     string
	ProvinceName   string
	CityID         string
	CityName       string
	District       string
	PostalCode     string
	IsDefault      bool
	IsDeleted      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName returns the database table name (optional, for documentation)
func (Address) TableName() string {
	return "addresses"
}