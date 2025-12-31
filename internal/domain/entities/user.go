package entities

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAdmin    UserRole = "admin"
)

// User represents a domain entity
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Role         UserRole
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsCustomer checks if the user has customer role
func (u *User) IsCustomer() bool {
	return u.Role == UserRoleCustomer
}
