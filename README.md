# Go Hexagonal Architecture Project

This project implements hexagonal architecture (ports and adapters) with:
- **GoFiber**: Web framework
- **Ent**: ORM for database operations
- **Huma**: OpenAPI-based API framework

## Project Structure

```
.
├── cmd/
│   └── api/                    # Application entry point
│       └── main.go
├── internal/
│   ├── domain/                 # Domain layer (core business logic)
│   │   ├── entities/          # Domain entities
│   │   └── ports/             # Interfaces (ports)
│   ├── application/           # Application layer
│   │   └── services/          # Use cases / business services
│   ├── infrastructure/        # Infrastructure layer
│   │   ├── adapters/
│   │   │   └── persistence/   # Database adapters (Ent)
│   │   └── config/            # Configuration
│   └── api/                   # API layer
│       └── handlers/          # HTTP handlers (GoFiber + Huma)
└── ent/
    └── schema/                # Ent schema definitions
```

## Hexagonal Architecture Layers

1. **Domain Layer** (`internal/domain/`): Core business logic, entities, and port interfaces
2. **Application Layer** (`internal/application/`): Use cases and services
3. **Infrastructure Layer** (`internal/infrastructure/`): External adapters (database, etc.)
4. **API Layer** (`internal/api/`): HTTP handlers and routing

## Getting Started

### Generate Ent Code

```bash
go generate ./ent
```

### Run the Application

```bash
go run cmd/api/main.go
```

### API Documentation

Once running, visit: http://localhost:8080/docs

## Next Steps

1. Generate Ent code: `go generate ./ent`
2. Add more entities in `internal/domain/entities/`
3. Define ports in `internal/domain/ports/`
4. Implement services in `internal/application/services/`
5. Create Ent schemas in `ent/schema/`
6. Implement repository adapters in `internal/infrastructure/adapters/persistence/`
7. Add HTTP handlers in `internal/api/handlers/`



go run -mod=mod entgo.io/ent/cmd/ent new --target internal/scheme User