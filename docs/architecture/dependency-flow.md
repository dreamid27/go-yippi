# Dependency Flow and Rules

## The Dependency Rule

The overriding rule that makes this architecture work:

> **Source code dependencies must point inward, toward higher-level policies.**

Inner circles (domain) know nothing about outer circles (infrastructure). Outer circles depend on inner circles, never the reverse.

## Layer Dependency Diagram

```
┌────────────────────────────────────────────────────────┐
│                 cmd/api/main.go                        │
│                 (Composition Root)                     │
│         Wires everything together                      │
└───────────────────┬────────────────────────────────────┘
                    │ depends on
                    ▼
┌────────────────────────────────────────────────────────┐
│              Infrastructure Layer                       │
│         internal/infrastructure/config                 │
│                                                        │
│  NO business logic, only configuration                │
└────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────┐
│                Adapters Layer                          │
│       internal/adapters/{api,persistence}              │
│                                                        │
│  Depends on: Application + Domain                     │
│  Imports: Services, Ports, Entities                   │
└───────────────────┬────────────────────────────────────┘
                    │ depends on
                    ▼
┌────────────────────────────────────────────────────────┐
│             Application Layer                          │
│         internal/application/services                  │
│                                                        │
│  Depends on: Domain ONLY                              │
│  Imports: Entities, Ports (interfaces)                │
└───────────────────┬────────────────────────────────────┘
                    │ depends on
                    ▼
┌────────────────────────────────────────────────────────┐
│               Domain Layer (CORE)                      │
│              internal/domain/                          │
│                                                        │
│  Depends on: NOTHING (pure Go)                        │
│  Exports: Entities, Ports, Errors                     │
└────────────────────────────────────────────────────────┘
```

## Import Rules by Layer

### Domain Layer (`internal/domain/`)

**Allowed imports**:
- Standard library only (`context`, `time`, etc.)
- NO application layer
- NO adapters layer
- NO infrastructure layer
- NO external frameworks

```go
// ✅ GOOD
package entities

import (
    "time"
)

type User struct {
    ID        int
    Name      string
    CreatedAt time.Time
}
```

```go
// ❌ BAD
package entities

import (
    "myapp/internal/adapters/persistence/db/ent" // NO!
)
```

### Application Layer (`internal/application/services/`)

**Allowed imports**:
- Domain layer (`internal/domain/entities`, `internal/domain/ports`, `internal/domain/errors`)
- Standard library

**NOT allowed**:
- Adapters layer (handlers, repositories)
- Infrastructure layer implementations
- External frameworks (Fiber, Ent, etc.)

```go
// ✅ GOOD
package services

import (
    "context"
    "myapp/internal/domain/entities"
    "myapp/internal/domain/ports"
    "myapp/internal/domain/errors"
)

type UserService struct {
    repo ports.UserRepository // Interface, not implementation
}
```

```go
// ❌ BAD
package services

import (
    "myapp/internal/adapters/persistence" // NO!
    "entgo.io/ent" // NO!
)
```

### Adapters Layer (`internal/adapters/`)

**Allowed imports**:
- Domain layer
- Application layer
- External frameworks (for implementation)
- Standard library

#### API Handlers (`internal/adapters/api/handlers/`)

```go
// ✅ GOOD
package handlers

import (
    "context"
    "myapp/internal/application/services"  // Uses services
    "myapp/internal/domain/entities"       // Uses entities
    "myapp/internal/domain/errors"         // Checks domain errors
    "myapp/internal/adapters/api/dto"      // Own DTOs
    "github.com/danielgtaylor/huma/v2"     // Framework
)

type UserHandler struct {
    service *services.UserService
}
```

```go
// ❌ BAD
package handlers

import (
    "myapp/internal/adapters/persistence/db/ent" // NO! Don't import other adapters
)
```

#### Repositories (`internal/adapters/persistence/`)

```go
// ✅ GOOD
package persistence

import (
    "context"
    "myapp/internal/domain/entities"
    "myapp/internal/domain/ports"
    "myapp/internal/domain/errors"
    "myapp/internal/adapters/persistence/db/ent"
    "entgo.io/ent"
)

type UserRepository struct {
    client *ent.Client
}

// Implements ports.UserRepository
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
    // Implementation using Ent
}
```

### Infrastructure Layer (`internal/infrastructure/`)

**Allowed imports**:
- Standard library
- External packages for configuration/utilities
- NO domain or application layers (should be independent)

```go
// ✅ GOOD
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    ServerPort string
    DBDSN      string
}
```

## Error Handling Flow

Errors flow from infrastructure to domain and back out.

### 1. Repository Layer (Adapter)

Convert infrastructure errors to domain errors:

```go
// internal/adapters/persistence/user_repository.go
func (r *UserRepository) FindByID(ctx context.Context, id int) (*entities.User, error) {
    entUser, err := r.client.User.Get(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            // Convert to domain error
            return nil, domainErrors.NewNotFoundError("User", id)
        }
        return nil, err
    }
    // Map ent.User to entities.User
    return mapEntUserToEntity(entUser), nil
}
```

### 2. Service Layer (Application)

Use domain errors for business logic:

```go
// internal/application/services/user_service.go
func (s *UserService) GetUser(ctx context.Context, id int) (*entities.User, error) {
    if id <= 0 {
        return nil, domainErrors.NewValidationError("User ID must be positive")
    }

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err // Propagate domain error
    }

    return user, nil
}
```

### 3. Handler Layer (Adapter)

Map domain errors to HTTP responses:

```go
// internal/adapters/api/handlers/user_handler.go
func (h *UserHandler) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
    user, err := h.service.GetUser(ctx, req.ID)
    if err != nil {
        // Check domain error types
        if errors.Is(err, domainErrors.ErrNotFound) {
            return nil, huma.Error404NotFound("User not found")
        }
        if errors.Is(err, domainErrors.ErrValidation) {
            return nil, huma.Error400BadRequest(err.Error())
        }
        return nil, huma.Error500InternalServerError("Internal server error")
    }

    // Map to DTO
    return &UserResponse{Body: mapUserToDTO(user)}, nil
}
```

## Data Flow and Mapping

Data transformations happen at adapter boundaries.

### Inbound Flow (HTTP → Domain)

```
HTTP Request (JSON)
    ↓
[DTO] CreateUserRequest
    ↓ (Handler maps to)
[Service Input] CreateUserInput
    ↓ (Service creates)
[Entity] entities.User
    ↓ (Repository maps to)
[Ent Model] ent.User
    ↓
Database
```

### Outbound Flow (Domain → HTTP)

```
Database
    ↓
[Ent Model] ent.User
    ↓ (Repository maps to)
[Entity] entities.User
    ↓ (Handler maps to)
[DTO] UserResponse
    ↓
HTTP Response (JSON)
```

## Dependency Injection

All dependencies are injected through constructors, wired in `cmd/api/main.go`:

```go
// cmd/api/main.go
func main() {
    // 1. Infrastructure: Create external connections
    cfg := config.LoadConfig()
    dbClient := ent.Open("postgres", cfg.DBDSN)
    defer dbClient.Close()

    // 2. Adapters: Create repository implementations
    userRepo := persistence.NewUserRepository(dbClient)
    productRepo := persistence.NewProductRepository(dbClient)

    // 3. Application: Create services with injected dependencies
    userService := services.NewUserService(userRepo)
    productService := services.NewProductService(productRepo)

    // 4. Adapters: Create handlers with injected services
    userHandler := handlers.NewUserHandler(userService)
    productHandler := handlers.NewProductHandler(productService)

    // 5. Setup API and register routes
    api := setupHumaAPI(fiberApp)
    userHandler.RegisterRoutes(api)
    productHandler.RegisterRoutes(api)

    // 6. Start server
    fiberApp.Listen(cfg.ServerPort)
}
```

## Critical Rules Summary

### ✅ DO

1. **Define ports (interfaces) in domain layer**
2. **Implement ports in adapter layer**
3. **Inject dependencies through constructors**
4. **Use domain errors across all layers**
5. **Map data at adapter boundaries**
6. **Keep domain layer pure (no external deps)**

### ❌ DON'T

1. **Import adapters in domain or application layers**
2. **Import infrastructure in handlers or services**
3. **Use framework-specific types in domain entities**
4. **Skip the service layer (handlers calling repos directly)**
5. **Put business logic in handlers or repositories**
6. **Use implementation types where interfaces are expected**

## Example: Full Request Flow

Let's trace a complete request: `GET /users/123`

```go
// 1. HTTP Request arrives
GET /users/123

// 2. Router maps to handler
// internal/adapters/api/handlers/user_handler.go
func (h *UserHandler) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
    // req.ID = 123

    // 3. Handler calls service (application layer)
    user, err := h.service.GetUser(ctx, req.ID)
    // ...
}

// 4. Service executes business logic
// internal/application/services/user_service.go
func (s *UserService) GetUser(ctx context.Context, id int) (*entities.User, error) {
    // Business validation
    if id <= 0 {
        return nil, domainErrors.NewValidationError("invalid ID")
    }

    // 5. Service calls repository through port interface
    user, err := s.repo.FindByID(ctx, id) // s.repo is ports.UserRepository
    // ...
}

// 6. Repository implementation queries database
// internal/adapters/persistence/user_repository.go
func (r *UserRepository) FindByID(ctx context.Context, id int) (*entities.User, error) {
    // Use Ent to query database
    entUser, err := r.client.User.Get(ctx, id)
    if ent.IsNotFound(err) {
        return nil, domainErrors.NewNotFoundError("User", id)
    }

    // 7. Map Ent model to domain entity
    return &entities.User{
        ID:   entUser.ID,
        Name: entUser.Name,
    }, nil
}

// 8. Response flows back through layers
// Handler receives entities.User, maps to DTO, returns HTTP response
```

## Related Documentation

- [Architecture Overview](./overview.md)
- [Hexagonal Pattern](./hexagonal-pattern.md)
- [Adding Features Guide](../guides/adding-features.md)
