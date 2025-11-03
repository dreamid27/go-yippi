# Hexagonal Architecture Pattern

## What is Hexagonal Architecture?

Hexagonal Architecture (also called Ports and Adapters) is an architectural pattern that isolates the core business logic from external concerns like databases, APIs, or third-party services.

## Why Hexagonal Architecture?

### Benefits

1. **Business Logic Independence**: Core domain logic is isolated from infrastructure
2. **Testability**: Easy to test business logic with mocks and stubs
3. **Flexibility**: Swap implementations without changing business logic
4. **Maintainability**: Clear separation of concerns
5. **Technology Agnostic**: Core doesn't depend on specific frameworks

### Trade-offs

- More boilerplate code (interfaces, DTOs)
- Steeper learning curve
- May be over-engineering for simple CRUD apps

## Core Concepts

### 1. The Hexagon (Core Domain)

The center of the architecture contains:
- **Entities**: Business objects
- **Use Cases**: Business operations
- **Ports**: Interfaces to the outside world

### 2. Ports

Ports are interfaces that define how the core communicates with the outside world.

**Two types of ports**:

#### Primary/Driving Ports (Inbound)
Interfaces that define what the application does (use cases).

Example:
```go
// internal/domain/ports/services.go
type UserService interface {
    CreateUser(ctx context.Context, input CreateUserInput) (*entities.User, error)
    GetUser(ctx context.Context, id int) (*entities.User, error)
}
```

#### Secondary/Driven Ports (Outbound)
Interfaces that define what the application needs from external systems.

Example:
```go
// internal/domain/ports/repository.go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, id int) (*entities.User, error)
}
```

### 3. Adapters

Adapters implement the ports and connect the core to external systems.

**Two types of adapters**:

#### Primary/Driving Adapters (Inbound)
Drive the application (e.g., HTTP handlers, CLI commands).

Example:
```go
// internal/adapters/api/handlers/user_handler.go
type UserHandler struct {
    service services.UserService // Uses primary port
}

func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
    user, err := h.service.CreateUser(ctx, req.Body)
    // ...
}
```

#### Secondary/Driven Adapters (Outbound)
Provide implementations for external systems (e.g., database repositories).

Example:
```go
// internal/adapters/persistence/user_repository.go
type UserRepository struct {
    client *ent.Client
}

// Implements domain port
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
    // Ent ORM implementation
}
```

## Project Structure Mapping

```
internal/
├── domain/                    # THE HEXAGON (CORE)
│   ├── entities/             # Business entities
│   ├── ports/                # Port definitions (interfaces)
│   └── errors/               # Domain errors
│
├── application/              # USE CASES (CORE)
│   └── services/            # Business services
│
├── adapters/                 # ADAPTERS (EXTERNAL)
│   ├── api/                 # Primary adapters (HTTP)
│   │   ├── handlers/        # HTTP handlers
│   │   └── dto/             # Data transfer objects
│   │
│   └── persistence/         # Secondary adapters (Database)
│       ├── user_repository.go
│       └── db/              # Ent schemas and generated code
│
└── infrastructure/          # INFRASTRUCTURE
    └── config/             # Configuration
```

## Dependency Flow

**Critical Rule**: Dependencies point inward toward the core.

```
┌─────────────────────────────────────┐
│   External Systems (DB, HTTP, etc)  │
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│          Adapters Layer             │ ◄─── Implements Ports
│  (Handlers, Repositories)           │
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│       Application Layer             │ ◄─── Uses Ports
│       (Services/Use Cases)          │
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│         Domain Layer                │ ◄─── Defines Ports
│    (Entities, Ports, Errors)        │      Pure Business Logic
└─────────────────────────────────────┘
```

## Example: Creating a User

### 1. Define Domain Entity
```go
// internal/domain/entities/user.go
type User struct {
    ID    int
    Name  string
    Email string
}
```

### 2. Define Repository Port
```go
// internal/domain/ports/repository.go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
}
```

### 3. Implement Service
```go
// internal/application/services/user_service.go
type UserService struct {
    repo ports.UserRepository // Depends on interface, not implementation
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*entities.User, error) {
    user := &entities.User{Name: input.Name, Email: input.Email}
    err := s.repo.Create(ctx, user)
    return user, err
}
```

### 4. Implement Repository Adapter
```go
// internal/adapters/persistence/user_repository.go
type UserRepository struct {
    client *ent.Client
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
    _, err := r.client.User.Create().
        SetName(user.Name).
        SetEmail(user.Email).
        Save(ctx)
    return err
}
```

### 5. Implement HTTP Handler Adapter
```go
// internal/adapters/api/handlers/user_handler.go
type UserHandler struct {
    service *services.UserService
}

func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
    user, err := h.service.CreateUser(ctx, req.Body)
    // Convert to DTO and return
}
```

### 6. Wire Everything Together
```go
// cmd/api/main.go
func main() {
    // Infrastructure
    db := ent.Open(...)

    // Adapters (implements ports)
    userRepo := persistence.NewUserRepository(db)

    // Services (uses ports)
    userService := services.NewUserService(userRepo)

    // Handlers (uses services)
    userHandler := handlers.NewUserHandler(userService)
}
```

## Testing Benefits

### Unit Test Domain Logic
```go
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{} // Mock implementation
    service := services.NewUserService(mockRepo)

    user, err := service.CreateUser(ctx, input)
    // Test business logic without database
}
```

### Integration Test Adapters
```go
func TestUserRepository_Create(t *testing.T) {
    // Use real database or test container
    repo := persistence.NewUserRepository(testDB)

    err := repo.Create(ctx, user)
    // Test actual database operations
}
```

## Key Principles

1. **The Dependency Rule**: Source code dependencies point inward
2. **Domain is King**: Core business logic has no dependencies
3. **Ports Before Adapters**: Define interfaces before implementations
4. **Separation of Concerns**: Each layer has a single responsibility
5. **Framework Independence**: Core doesn't know about frameworks

## Anti-Patterns to Avoid

❌ **Don't import infrastructure in domain**:
```go
// BAD: Domain entity importing Ent
package entities
import "myapp/internal/adapters/persistence/db/ent"
```

❌ **Don't import adapters in services**:
```go
// BAD: Service importing repository implementation
package services
import "myapp/internal/adapters/persistence"
```

❌ **Don't bypass the service layer**:
```go
// BAD: Handler calling repository directly
func (h *UserHandler) CreateUser() {
    h.repository.Create(...) // Should call h.service.CreateUser()
}
```

✅ **Always use ports (interfaces)**:
```go
// GOOD: Service depends on interface
type UserService struct {
    repo ports.UserRepository // Interface from domain
}
```

## Further Reading

- [Architecture Overview](./overview.md)
- [Dependency Flow Details](./dependency-flow.md)
- [Adding New Features Guide](../guides/adding-features.md)
