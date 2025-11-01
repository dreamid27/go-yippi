# Domain Errors

This package contains all domain-level error types used throughout the application. By defining errors at the domain level, we maintain proper hexagonal architecture boundaries and prevent infrastructure concerns from leaking into the application and API layers.

## Error Types

### Base Errors

- `ErrNotFound` - Resource not found
- `ErrInvalidInput` - Invalid input provided
- `ErrDuplicateEntry` - Resource already exists
- `ErrUnauthorized` - User not authenticated
- `ErrForbidden` - User lacks permission
- `ErrInternal` - Internal server error

### Concrete Error Types

#### NotFoundError

Used when a specific resource cannot be found.

```go
err := domainErrors.NewNotFoundError("User", 123)
// Error message: "User with id 123 not found"
```

#### ValidationError

Used when input validation fails.

```go
err := domainErrors.NewValidationError("email", "must be a valid email address")
// Error message: "validation error on field 'email': must be a valid email address"
```

#### DuplicateError

Used when attempting to create a resource that already exists.

```go
err := domainErrors.NewDuplicateError("User", "email", "user@example.com")
// Error message: "User with email 'user@example.com' already exists"
```

## Usage in Layers

### Repository Layer (Infrastructure)

Convert infrastructure-specific errors to domain errors:

```go
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int) (*entities.User, error) {
    found, err := r.client.User.Get(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, domainErrors.NewNotFoundError("User", id)
        }
        return nil, err
    }
    return toEntity(found), nil
}
```

### Service Layer (Application)

Add business logic validation:

```go
func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
    if user.Age < 0 {
        return domainErrors.NewValidationError("age", "must be non-negative")
    }
    return s.repo.Create(ctx, user)
}
```

### Handler Layer (API)

Check error types and return appropriate HTTP responses:

```go
func (h *UserHandler) GetUser(ctx context.Context, input *dto.GetUserRequest) (*dto.UserResponse, error) {
    user, err := h.service.GetUser(ctx, input.ID)
    if err != nil {
        if errors.Is(err, domainErrors.ErrNotFound) {
            return nil, huma.Error404NotFound("User not found")
        }
        return nil, huma.Error500InternalServerError("Failed to get user", err)
    }
    return mapToResponse(user), nil
}
```

## Benefits

1. **No Infrastructure Leakage**: Handlers don't import database libraries
2. **Consistent Error Handling**: Same error types across all layers
3. **Type Safety**: Use `errors.Is()` and `errors.As()` for type checking
4. **Clear Semantics**: Errors convey business meaning, not technical details
5. **Testability**: Easy to mock and test error scenarios
