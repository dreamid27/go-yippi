# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based REST API implementing **Hexagonal Architecture** (Ports and Adapters pattern) with:
- **GoFiber** for HTTP routing
- **Huma v2** for OpenAPI specification and validation
- **Ent** as the ORM for database operations
- **PostgreSQL** as the primary database

## Documentation

Comprehensive documentation is available in the `docs/` folder:
- **[Architecture](docs/architecture/)**: System architecture, hexagonal pattern, and dependency flow
- **[Guides](docs/guides/)**: Getting started, adding features, and file uploads
- **[API Documentation](docs/api/)**: Endpoint documentation for products 
- **[Infrastructure](docs/infrastructure/)**: Database and MinIO setup
- **Development**: Testing and contribution guidelines

For a complete overview, see [docs/README.md](docs/README.md).

## Build & Run Commands

### Generate Ent Code
Always run after modifying schemas in `internal/adapters/persistence/db/schema/`:
```bash
make generate
```

Or directly:
```bash
go run -mod=mod entgo.io/ent/cmd/ent generate --target ./internal/adapters/persistence/db/ent ./internal/adapters/persistence/db/schema
```

### Run Application
```bash
make run          # Generates code and runs
go run cmd/api/main.go  # Direct run
```

### Development with Live Reload
```bash
make dev          # Uses Air for hot reloading
```

### Build
```bash
make build        # Generates code and builds binary to bin/api
go build ./...    # Build all packages
```

### Test
```bash
make test         # Run all tests
go test ./...     # Run tests directly
```

### Clean
```bash
make clean        # Remove build artifacts
```

## Architecture & Layer Boundaries

This project strictly follows hexagonal architecture. **Critical**: dependencies flow inward only.

For comprehensive architecture documentation, see:
- [Architecture Overview](docs/architecture/overview.md)
- [Hexagonal Pattern Details](docs/architecture/hexagonal-pattern.md)
- [Dependency Flow Rules](docs/architecture/dependency-flow.md)

### Layer Structure

```
Domain Layer (core) → Application Layer → Infrastructure/API (adapters)
```

1. **Domain Layer** (`internal/domain/`)
   - **Entities** (`entities/`): Pure business objects (User, Product, etc.)
   - **Ports** (`ports/`): Interfaces defining contracts (repositories, services)
   - **Errors** (`errors/`): Domain-specific error types
   - **Rules**: NO external dependencies, NO framework imports

2. **Application Layer** (`internal/application/services/`)
   - Business logic and use cases
   - Depends ONLY on domain layer (entities, ports)
   - Services implement business workflows

3. **Adapters Layer** (`internal/adapters/`)
   - **API Adapter** (`api/`): HTTP/REST interface
     - **Handlers** (`api/handlers/`): HTTP request handlers
     - **DTOs** (`api/dto/`): Request/response data transfer objects
   - **Persistence Adapter** (`persistence/`): Database interface
     - **Repositories** (`persistence/*.go`): Implement domain ports using Ent
     - **Ent schemas** (`persistence/db/schema/`): Database schema definitions
     - **Generated Ent code** (`persistence/db/ent/`): Auto-generated ORM code
   - Depends on application services, NOT on infrastructure directly

4. **Infrastructure Layer** (`internal/infrastructure/`)
   - **Config** (`config/`): Application configuration
   - Cross-cutting concerns (logging, monitoring, etc.)

### Critical Architecture Rules

**NEVER import infrastructure packages in handlers or services:**
- ❌ `internal/adapters/api/handlers` importing `internal/adapters/persistence/db/ent`
- ❌ `internal/application/services` importing Ent or database libraries
- ✅ Use domain errors (`internal/domain/errors`) for error handling across layers

**Error Handling Pattern:**
- Repository layer converts infrastructure errors (e.g., `ent.IsNotFound`) to domain errors
- Service layer uses domain errors for business logic validation
- Handler layer checks domain errors with `errors.Is()` and returns appropriate HTTP codes

Example:
```go
// Repository (infrastructure)
if ent.IsNotFound(err) {
    return nil, domainErrors.NewNotFoundError("User", id)
}

// Handler (API)
if errors.Is(err, domainErrors.ErrNotFound) {
    return nil, huma.Error404NotFound("User not found")
}
```

## Adding New Features

For a detailed step-by-step guide on adding new entities and features, see [docs/guides/adding-features.md](docs/guides/adding-features.md).

### Quick Reference: Adding a New Entity

1. Create domain entity in `internal/domain/entities/`
2. Create Ent schema in `internal/adapters/persistence/db/schema/`
3. Define repository port in `internal/domain/ports/repository.go`
4. Define service port in `internal/domain/ports/service.go` (for testability)
5. Run `make generate` to generate Ent code
6. Implement repository in `internal/adapters/persistence/`
7. Create service in `internal/application/services/`
8. **Write service unit tests** in `internal/application/services/*_test.go`
9. Create DTOs in `internal/adapters/api/dto/`
10. Create handler in `internal/adapters/api/handlers/`
11. **Write handler unit tests** in `internal/adapters/api/handlers/*_test.go`
12. Register routes in the handler's `RegisterRoutes()` method
13. Wire dependencies in `cmd/api/main.go`

**Note**: Steps 8 and 11 (testing) are **critical** and should not be skipped.

### Ent Schema Location

Ent schemas are stored in: `internal/adapters/persistence/db/schema/`

Generated code goes to: `internal/adapters/persistence/db/ent/`

### Creating New Ent Schema

```bash
go run -mod=mod entgo.io/ent/cmd/ent new --target internal/adapters/persistence/db/schema EntityName
```

## Configuration

Environment variables (see `internal/infrastructure/config/config.go`):
- `SERVER_PORT` (default: 8080)
- `SERVER_HOST` (default: 0.0.0.0)
- `DB_DRIVER` (default: postgres)
- `DB_DSN` (default: host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable)

## API Documentation

When the application is running, OpenAPI docs are available at:
```
http://localhost:8080/docs
```

This is auto-generated by Huma based on handler registrations and DTO struct tags.

## Dependency Injection Flow

All wiring happens in `cmd/api/main.go`:

```go
// Infrastructure
client := ent.Open(...)

// Repositories (implement ports)
userRepo := persistence.NewUserRepository(client)

// Services (depend on ports)
userService := services.NewUserService(userRepo)

// Handlers (depend on services)
userHandler := handlers.NewUserHandler(userService)

// Register routes
userHandler.RegisterRoutes(humaAPI)
```

## Huma Request/Response Pattern

Handlers use Huma's type-safe request/response pattern:

- Request types embed path params, query params, and body
- Response types embed the response body
- Validation is automatic based on struct tags
- OpenAPI schema is auto-generated

Example:
```go
type GetUserRequest struct {
    ID int `path:"id" doc:"User ID"`
}

type UserResponse struct {
    Body struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }
}
```

### IMPORTANT: DTO Best Practices

**NEVER use inline anonymous structs in DTOs:**

❌ **BAD - Inline anonymous struct:**
```go
type ListUsersResponse struct {
    Body struct {
        Users []struct {  // Anonymous inline struct
            ID   int    `json:"id"`
            Name string `json:"name"`
        } `json:"users"`
    }
}
```

This causes:
- OpenAPI schema generation issues
- Type conflicts in Huma's schema registry
- Incorrect field definitions in generated documentation

✅ **GOOD - Named struct type:**
```go
// Create a named type for list items
type UserListItem struct {
    ID   int    `json:"id" doc:"User ID"`
    Name string `json:"name" doc:"User name"`
}

type ListUsersResponse struct {
    Body struct {
        Users []UserListItem `json:"users" doc:"List of users"`
    }
}
```

**Benefits:**
- Proper OpenAPI schema generation
- Reusable type definitions
- Better documentation through struct tags
- Type safety and clarity

**When creating list/collection responses:**
1. Always create a separate named type for list items (e.g., `UserListItem`, `ProductListItem`)
2. Add `doc` tags to all fields for OpenAPI documentation
3. Use the named type in the response struct

## Database Migrations

Migrations are handled automatically by Ent's auto-migration:
```go
client.Schema.Create(context.Background())
```

For production, use Ent's versioned migrations or Atlas.

## Testing Strategy

This project follows a layered testing approach that respects hexagonal architecture boundaries.

### Testing Stack

- **`testing`** - Go standard library for test runner
- **`github.com/stretchr/testify`** - Assertion and mocking framework
  - `assert` - Friendly assertions for test validation
  - `require` - Critical assertions that stop test execution on failure
  - `mock` - Interface mocking for isolation testing

### Testing Layers

Tests should be written for **three layers** in hexagonal architecture:

#### 1. Service Layer Tests (Priority: HIGH)
**Location**: `internal/application/services/*_test.go`

Service layer tests provide the **highest value** as they test all business logic and validation rules.

**What to test:**
- Business validation logic (required fields, value constraints)
- Default value handling
- Domain entity state transitions
- Error handling and domain error generation
- Repository interaction (via mocks)

**Example:**
```go
func TestCreateProduct_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockProductRepository)
    service := NewProductService(mockRepo)
    ctx := context.Background()

    product := &entities.Product{
        SKU:   "TEST-001",
        Name:  "Test Product",
        Price: 99.99,
    }

    mockRepo.On("Create", ctx, mock.Anything).Return(nil)

    // Act
    err := service.CreateProduct(ctx, product)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, entities.ProductStatusDraft, product.Status)
    mockRepo.AssertExpectations(t)
}
```

**Key principles:**
- Mock the repository interface (from `internal/domain/ports`)
- Test business logic in isolation
- Verify all validation rules
- Test both success and error scenarios
- Use `mock.MatchedBy()` for complex argument validation

#### 2. Handler Layer Tests (Priority: MEDIUM)
**Location**: `internal/adapters/api/handlers/*_test.go`

Handler tests verify HTTP adapter concerns and error translation.

**What to test:**
- DTO → Domain Entity mapping
- Domain Entity → Response DTO mapping
- Domain errors → HTTP status code translation:
  - `ErrInvalidInput` → 400 Bad Request
  - `ErrNotFound` → 404 Not Found
  - `ErrDuplicateEntry` → 409 Conflict
  - Generic errors → 500 Internal Server Error
- Optional field handling (nil pointers)
- Request validation

**Example:**
```go
func TestCreateProduct_ValidationError(t *testing.T) {
    // Arrange
    mockService := new(MockProductService)
    handler := NewProductHandler(mockService)
    ctx := context.Background()

    input := &dto.CreateProductRequest{}
    input.Body.SKU = "" // Invalid

    validationErr := domainErrors.NewValidationError("sku", "SKU is required")
    mockService.On("CreateProduct", ctx, mock.Anything).Return(validationErr)

    // Act
    response, err := handler.CreateProduct(ctx, input)

    // Assert
    require.Error(t, err)
    assert.Nil(t, response)

    var humaErr huma.StatusError
    require.True(t, errors.As(err, &humaErr))
    assert.Equal(t, 400, humaErr.GetStatus())
    mockService.AssertExpectations(t)
}
```

**Key principles:**
- Mock the service interface (from `internal/domain/ports`)
- Test HTTP-specific concerns only
- Verify error status code mapping
- Test DTO mapping completeness

#### 3. Repository Layer Tests (Priority: LOW - Optional)
**Location**: `internal/adapters/persistence/*_test.go`

Repository tests are integration tests requiring a test database.

**What to test:**
- Actual database operations (CRUD)
- Infrastructure error → Domain error conversion
- Database constraint handling
- Query logic and filters

**Requires:**
- Test database setup (PostgreSQL)
- Use `testcontainers-go` for isolated testing
- Database migrations/schema setup

**Note**: Repository tests are **optional** for most features. Focus on service and handler tests first.

### Test File Naming

- Test files: `*_test.go` (same package as code under test)
- Mock types: Define in test files (e.g., `MockProductService`, `MockProductRepository`)
- Test functions: `Test<FunctionName>_<Scenario>`

Examples:
- `TestCreateProduct_Success`
- `TestCreateProduct_ValidationError`
- `TestCreateProduct_DuplicateError`

### Interface-Based Testing

**CRITICAL**: Handlers and tests must depend on **interfaces**, not concrete types.

**Service interfaces** are defined in `internal/domain/ports/service.go`:

```go
type ProductService interface {
    CreateProduct(ctx context.Context, product *entities.Product) error
    GetProduct(ctx context.Context, id int) (*entities.Product, error)
    // ... other methods
}
```

**Handler dependency injection:**
```go
// ✅ CORRECT - Depends on interface
type ProductHandler struct {
    service ports.ProductService
}

func NewProductHandler(service ports.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

// ❌ WRONG - Depends on concrete type
type ProductHandler struct {
    service *services.ProductService  // Don't do this!
}
```

This allows:
- Easy mocking in tests
- Loose coupling between layers
- Better adherence to hexagonal architecture

### Running Tests

```bash
# Run all tests
make test
go test ./...

# Run tests for specific package
go test ./internal/application/services -v
go test ./internal/adapters/api/handlers -v

# Run specific test function
go test ./internal/application/services -v -run TestCreateProduct_Success

# Run tests with coverage
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Structure (AAA Pattern)

All tests should follow the **Arrange-Act-Assert** pattern:

```go
func TestSomething(t *testing.T) {
    // Arrange - Set up test data and mocks
    mockRepo := new(MockProductRepository)
    service := NewProductService(mockRepo)
    product := &entities.Product{...}
    mockRepo.On("Create", ...).Return(nil)

    // Act - Execute the function under test
    err := service.CreateProduct(ctx, product)

    // Assert - Verify the results
    require.NoError(t, err)
    assert.Equal(t, expected, actual)
    mockRepo.AssertExpectations(t)
}
```

### Mock Best Practices

**1. Use testify/mock for interface mocking:**
```go
type MockProductRepository struct {
    mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *entities.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}
```

**2. Set up expectations:**
```go
// Exact argument matching
mockRepo.On("Create", ctx, product).Return(nil)

// Flexible matching
mockRepo.On("Create", ctx, mock.Anything).Return(nil)

// Custom matching
mockRepo.On("Create", ctx, mock.MatchedBy(func(p *entities.Product) bool {
    return p.SKU == "TEST-001" && p.Price > 0
})).Return(nil)
```

**3. Always verify expectations:**
```go
mockRepo.AssertExpectations(t)  // Fails test if expectations not met
```

### When Adding New Features

When implementing a new feature, **always add tests**:

1. **Write service tests first** (TDD approach recommended)
2. **Then write handler tests**
3. **Repository tests are optional** (add only if complex query logic)

Example workflow:
```bash
# 1. Create service test file
touch internal/application/services/feature_service_test.go

# 2. Write failing tests
# 3. Implement service logic until tests pass
go test ./internal/application/services -v -run TestFeature

# 4. Create handler test file
touch internal/adapters/api/handlers/feature_handler_test.go

# 5. Write failing tests
# 6. Implement handler until tests pass
go test ./internal/adapters/api/handlers -v -run TestFeature
```

### Coverage Goals

- **Service layer**: Aim for 80%+ coverage
- **Handler layer**: Aim for 70%+ coverage
- **Repository layer**: Optional (integration tests)

Check coverage:
```bash
go test ./internal/application/services -cover
go test ./internal/adapters/api/handlers -cover
```
