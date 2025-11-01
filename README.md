# Go Hexagonal Architecture API

A production-ready REST API built with **Hexagonal Architecture** (Ports and Adapters pattern) demonstrating clean architecture principles, dependency inversion, and proper layer separation.

## Tech Stack

- **[GoFiber](https://gofiber.io/)** - Fast HTTP web framework
- **[Huma v2](https://huma.rocks/)** - Modern OpenAPI 3.1 framework with automatic validation
- **[Ent](https://entgo.io/)** - Entity framework for Go (ORM)
- **[PostgreSQL](https://www.postgresql.org/)** - Primary database
- **Go 1.23+** - Programming language

## Features

- ✅ **Hexagonal Architecture** - Clean separation of concerns with dependency inversion
- ✅ **Domain-Driven Design** - Business logic isolated from infrastructure
- ✅ **OpenAPI Documentation** - Auto-generated interactive API docs
- ✅ **Type-Safe Validation** - Request/response validation with Huma
- ✅ **Domain Error Handling** - Custom error types without infrastructure leakage
- ✅ **Repository Pattern** - Abstract data access through ports
- ✅ **Automatic Migrations** - Database schema managed by Ent
- ✅ **Hot Reload** - Development mode with Air

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                           # Application entry point & DI
├── internal/
│   ├── domain/                               # Core business logic (no dependencies)
│   │   ├── entities/                         # Domain entities (User, Product)
│   │   ├── errors/                           # Domain-specific errors
│   │   └── ports/                            # Repository & service interfaces
│   ├── application/                          # Use cases & business workflows
│   │   └── services/                         # Business logic implementation
│   ├── adapters/                             # External adapters
│   │   ├── api/                              # HTTP/REST adapter
│   │   │   ├── dto/                          # Request/response DTOs
│   │   │   └── handlers/                     # HTTP handlers
│   │   └── persistence/                      # Database adapter
│   │       ├── db/
│   │       │   ├── schema/                   # Ent schema definitions
│   │       │   └── ent/                      # Generated Ent code
│   │       ├── user_repository.go            # Repository implementations
│   │       └── product_repository.go
│   └── infrastructure/
│       └── config/                           # Configuration management
├── Makefile                                  # Build & development commands
├── CLAUDE.md                                 # AI assistant instructions
└── PRODUCT_API.md                            # Product API documentation
```

## Hexagonal Architecture Overview

This project strictly follows the **Ports and Adapters** pattern with clear layer boundaries:

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│                   (Adapters - Driving)                       │
│                  handlers/ + dto/                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                          │
│                  (Use Cases/Services)                        │
│                     services/                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                             │
│            (Core Business Logic - No Dependencies)           │
│          entities/ + ports/ + errors/                        │
└─────────────────────┬───────────────────────┬───────────────┘
                      │                       │
                      ▼                       ▼
         ┌────────────────────┐  ┌────────────────────┐
         │  Persistence        │  │   Other Adapters   │
         │  (Ent/PostgreSQL)   │  │   (Future: Cache,  │
         │  Adapters - Driven  │  │    Message Queue)  │
         └────────────────────┘  └────────────────────┘
```

### Layer Rules

1. **Domain Layer** (`internal/domain/`)
   - ❌ NO external dependencies
   - ❌ NO framework imports
   - ✅ Pure business logic only
   - ✅ Defines interfaces (ports)

2. **Application Layer** (`internal/application/`)
   - ✅ Depends ONLY on domain layer
   - ✅ Orchestrates business workflows
   - ❌ NO infrastructure knowledge

3. **Adapters** (`internal/adapters/`)
   - ✅ Implements domain ports
   - ✅ Handles external concerns
   - ✅ Converts between layers

4. **Infrastructure** (`internal/infrastructure/`)
   - ✅ Cross-cutting concerns (config, logging)
   - ✅ Can be used by adapters

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 12 or higher
- Make (optional, for convenience commands)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-hex-yippi
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL**
   ```bash
   # Create database
   createdb go-test

   # Or use Docker
   docker run --name postgres -e POSTGRES_PASSWORD=adminadmin \
     -e POSTGRES_USER=admin -e POSTGRES_DB=go-test \
     -p 5432:5432 -d postgres:15
   ```

4. **Configure environment** (optional)
   ```bash
   export SERVER_PORT=8080
   export SERVER_HOST=0.0.0.0
   export DB_DSN="host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable"
   ```

5. **Generate Ent code**
   ```bash
   make generate
   ```

6. **Run the application**
   ```bash
   make run
   ```

The API will be available at `http://localhost:8080`

### Development Mode (Hot Reload)

```bash
make dev
```

This uses [Air](https://github.com/cosmtrek/air) for automatic reloading on file changes.

## Available Commands

```bash
make generate    # Generate Ent code from schemas
make run         # Generate code and run application
make dev         # Run with hot reload (Air)
make build       # Generate code and build binary to bin/api
make test        # Run all tests
make clean       # Remove build artifacts
go build ./...   # Build all packages directly
```

## API Documentation

### Interactive Documentation

Once the server is running, visit:
```
http://localhost:8080/docs
```

You'll see a **Swagger UI** with all available endpoints, request/response schemas, and the ability to test API calls directly.

### Available APIs

#### User API
- `POST /users` - Create user
- `GET /users` - List all users
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

#### Product API
- `POST /products` - Create product
- `GET /products` - List all products
- `GET /products/{id}` - Get product by ID
- `GET /products/sku/{sku}` - Get product by SKU
- `GET /products/slug/{slug}` - Get product by slug
- `GET /products/status/{status}` - List by status
- `PUT /products/{id}` - Update product
- `POST /products/{id}/publish` - Publish product
- `POST /products/{id}/archive` - Archive product
- `DELETE /products/{id}` - Delete product

See [PRODUCT_API.md](PRODUCT_API.md) for detailed Product API documentation.

## Configuration

Configuration is loaded from environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `SERVER_HOST` | `0.0.0.0` | HTTP server host |
| `DB_DRIVER` | `postgres` | Database driver |
| `DB_DSN` | See below | Database connection string |

**Default DB_DSN:**
```
host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable
```

## Development Workflow

### Adding a New Entity

Follow these steps to add a new entity (e.g., `Order`):

1. **Create domain entity**
   ```bash
   # Create internal/domain/entities/order.go
   ```

2. **Create Ent schema**
   ```bash
   go run -mod=mod entgo.io/ent/cmd/ent new \
     --target internal/adapters/persistence/db/schema Order
   ```

3. **Define repository port**
   ```go
   // Add to internal/domain/ports/repository.go
   type OrderRepository interface {
       Create(ctx context.Context, order *entities.Order) error
       // ... other methods
   }
   ```

4. **Generate Ent code**
   ```bash
   make generate
   ```

5. **Implement repository**
   ```bash
   # Create internal/adapters/persistence/order_repository.go
   ```

6. **Create service**
   ```bash
   # Create internal/application/services/order_service.go
   ```

7. **Create DTOs**
   ```bash
   # Create internal/adapters/api/dto/order_dto.go
   ```

8. **Create handlers**
   ```bash
   # Create internal/adapters/api/handlers/order_handler.go
   ```

9. **Wire dependencies** in `cmd/api/main.go`
   ```go
   orderRepo := persistence.NewOrderRepository(client)
   orderService := services.NewOrderService(orderRepo)
   orderHandler := handlers.NewOrderHandler(orderService)
   orderHandler.RegisterRoutes(humaAPI)
   ```

### Error Handling Best Practices

**Repository Layer** (convert infrastructure errors → domain errors):
```go
if ent.IsNotFound(err) {
    return nil, domainErrors.NewNotFoundError("User", id)
}
```

**Handler Layer** (convert domain errors → HTTP responses):
```go
if errors.Is(err, domainErrors.ErrNotFound) {
    return nil, huma.Error404NotFound("User not found")
}
```

See [internal/domain/errors/README.md](internal/domain/errors/README.md) for complete error handling guide.

## Architecture Principles

### Dependency Rule
Dependencies flow **inward only**:
```
Adapters → Application → Domain
```

### What NOT to do

❌ **Never import infrastructure in handlers**
```go
// BAD - Handler importing Ent
import "yourapp/internal/adapters/persistence/db/ent"
```

❌ **Never import infrastructure in services**
```go
// BAD - Service importing database library
import "entgo.io/ent"
```

✅ **Always use domain errors**
```go
// GOOD - Using domain errors
import domainErrors "yourapp/internal/domain/errors"
```

## Database Migrations

Migrations are handled automatically by Ent:

```go
// Auto-migration on startup
client.Schema.Create(context.Background())
```

For production, use Ent's [versioned migrations](https://entgo.io/docs/versioned-migrations) or [Atlas](https://atlasgo.io/).

## Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/application/services

# Test with coverage
go test -cover ./...
```

## Project Status

- ✅ User CRUD API
- ✅ Product API with status workflow
- ✅ Domain error handling
- ✅ OpenAPI documentation
- 🚧 Authentication (planned)
- 🚧 Authorization (planned)
- 🚧 Order management (planned)

## Resources

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Ent Documentation](https://entgo.io/docs/getting-started)
- [Huma Documentation](https://huma.rocks/)
- [GoFiber Documentation](https://docs.gofiber.io/)

## License

MIT

## Contributing

1. Follow the hexagonal architecture principles
2. Maintain layer boundaries (no infrastructure in handlers/services)
3. Use domain errors for error handling
4. Add tests for new features
5. Update documentation

---

For AI assistants working on this project, see [CLAUDE.md](CLAUDE.md) for detailed architecture guidelines and development instructions.
