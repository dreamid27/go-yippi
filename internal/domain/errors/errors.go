package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates that the provided input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateEntry indicates that the resource already exists
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrUnauthorized indicates that the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the user does not have permission
	ErrForbidden = errors.New("forbidden")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal error")
)

// NotFoundError represents a resource not found error with additional context
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %v not found", e.Resource, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource string, id interface{}) error {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// DuplicateError represents a duplicate entry error
type DuplicateError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s with %s '%v' already exists", e.Resource, e.Field, e.Value)
}

func (e *DuplicateError) Is(target error) bool {
	return target == ErrDuplicateEntry
}

// NewDuplicateError creates a new DuplicateError
func NewDuplicateError(resource, field string, value interface{}) error {
	return &DuplicateError{
		Resource: resource,
		Field:    field,
		Value:    value,
	}
}
