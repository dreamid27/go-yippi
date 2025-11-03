# Architecture Overview

## System Architecture

Go-Hex-Yippi is built using **Hexagonal Architecture** (also known as Ports and Adapters pattern), which promotes separation of concerns and makes the application highly testable and maintainable.

## High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     External World                           │
│  (HTTP Clients, Databases, File Storage, etc.)              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Adapters Layer                            │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ API Handler │  │ Persistence  │  │ External Services│   │
│  │  (HTTP/REST)│  │ (PostgreSQL) │  │    (MinIO)       │   │
│  └─────────────┘  └──────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           Business Services                          │   │
│  │  (UserService, ProductService, StorageService)      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer (Core)                      │
│  ┌──────────┐  ┌───────┐  ┌────────────────────────────┐   │
│  │ Entities │  │ Ports │  │      Domain Errors         │   │
│  │          │  │(Intf) │  │                            │   │
│  └──────────┘  └───────┘  └────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (Core)
**Location**: `internal/domain/`

The heart of the application containing:
- **Entities**: Pure business objects (User, Product, StorageFile)
- **Ports**: Interface definitions (repository contracts, service contracts)
- **Domain Errors**: Business-specific error types

**Rules**:
- No external dependencies
- No framework imports
- Pure Go code
- Contains only business logic

### 2. Application Layer
**Location**: `internal/application/services/`

Contains business use cases and orchestration:
- Implements business workflows
- Coordinates between different domain entities
- Depends ONLY on domain layer (entities and ports)
- Framework-agnostic

**Examples**: UserService, ProductService, StorageFileService

### 3. Adapters Layer
**Location**: `internal/adapters/`

Implements the ports defined in the domain layer:

#### API Adapter
- **Handlers** (`api/handlers/`): HTTP request handlers using Huma
- **DTOs** (`api/dto/`): Request/response data structures
- Translates HTTP requests to service calls

#### Persistence Adapter
- **Repositories** (`persistence/`): Database operations using Ent ORM
- Implements repository ports from domain layer
- Handles data persistence and retrieval

### 4. Infrastructure Layer
**Location**: `internal/infrastructure/`

Cross-cutting concerns:
- Configuration management
- Logging
- Monitoring
- Shared utilities

## Key Design Principles

### Dependency Rule
**Dependencies flow inward**: Outer layers can depend on inner layers, but never the reverse.

```
Infrastructure/API → Application → Domain
      (outer)                      (inner)
```

### Interface Segregation
Ports (interfaces) are defined in the domain layer, implemented in adapters.

### Testability
Each layer can be tested independently:
- Domain: Pure unit tests
- Application: Service tests with mock repositories
- Adapters: Integration tests

## Technology Stack

| Layer | Technologies |
|-------|-------------|
| API Framework | GoFiber + Huma v2 |
| ORM | Ent |
| Database | PostgreSQL |
| Object Storage | MinIO (S3-compatible) |
| Configuration | Environment variables |
| Documentation | OpenAPI (auto-generated) |

## Request Flow Example

Here's how a typical request flows through the system:

1. **HTTP Request** arrives at the API handler
2. **Handler** validates request (DTOs) and calls service
3. **Service** executes business logic, calls repository through port
4. **Repository** performs database operations using Ent
5. **Response** flows back through the layers
6. **Handler** serializes response to JSON

```
HTTP Request
    ↓
[Handler] → validates DTO
    ↓
[Service] → business logic
    ↓
[Repository Port]
    ↓
[Repository Implementation] → Ent ORM
    ↓
Database
```

## Error Handling Strategy

Errors flow from inner layers to outer layers:

1. **Repository Layer**: Converts infrastructure errors (e.g., `ent.IsNotFound`) to domain errors
2. **Service Layer**: Uses domain errors for business validation
3. **Handler Layer**: Maps domain errors to HTTP status codes

See [Dependency Flow](./dependency-flow.md) for detailed error handling patterns.

## Related Documentation

- [Hexagonal Pattern Details](./hexagonal-pattern.md)
- [Dependency Flow Rules](./dependency-flow.md)
- [Adding New Features](../guides/adding-features.md)
