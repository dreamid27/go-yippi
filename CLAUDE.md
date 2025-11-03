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
- **[API Documentation](docs/api/)**: Endpoint documentation for products and storage files
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
4. Run `make generate` to generate Ent code
5. Implement repository in `internal/adapters/persistence/`
6. Create service in `internal/application/services/`
7. Create DTOs in `internal/adapters/api/dto/`
8. Create handler in `internal/adapters/api/handlers/`
9. Register routes in the handler's `RegisterRoutes()` method
10. Wire dependencies in `cmd/api/main.go`

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
