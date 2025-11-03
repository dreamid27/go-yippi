# Go-Hex-Yippi Documentation

Welcome to the Go-Hex-Yippi documentation! This project is a REST API built with Go using Hexagonal Architecture.

## Table of Contents

### 🏗️ Architecture
- [Architecture Overview](./architecture/overview.md) - High-level system architecture
- [Hexagonal Pattern](./architecture/hexagonal-pattern.md) - Understanding the hexagonal architecture
- [Dependency Flow](./architecture/dependency-flow.md) - Layer dependencies and rules

### 📚 Guides
- [Getting Started](./guides/getting-started.md) - Quick start guide for developers
- [Adding Features](./guides/adding-features.md) - Step-by-step guide to add new features
- [File Upload Guide](./guides/file-upload.md) - Working with file uploads

### 🔌 API Documentation
- [Product API](./api/products.md) - Product management endpoints
- [Storage Files API](./api/storage-files.md) - File storage and management
- **OpenAPI Docs**: Available at `http://localhost:8080/docs` when running

### 🛠️ Infrastructure
- [Database](./infrastructure/database.md) - Database setup, migrations, and Ent ORM
- [MinIO Integration](./infrastructure/minio.md) - Object storage configuration

### 👨‍💻 Development
- [Testing](./development/testing.md) - Testing guidelines and examples
- [Contributing](./development/contributing.md) - How to contribute to the project

## Quick Links

- [Main README](../README.md) - Project overview
- [CLAUDE.md](../CLAUDE.md) - Claude Code assistant instructions
- [Makefile](../Makefile) - Available make commands

## Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: GoFiber
- **API Framework**: Huma v2
- **ORM**: Ent
- **Database**: PostgreSQL
- **Object Storage**: MinIO (S3-compatible)

## Project Structure

```
.
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/          # Business entities and interfaces (core)
│   ├── application/     # Use cases and business logic
│   ├── adapters/        # External interfaces (HTTP, DB)
│   └── infrastructure/  # Cross-cutting concerns
├── docs/                # This documentation
└── Makefile            # Build and run commands
```

## Getting Help

- Check the relevant documentation section above
- Review the [Getting Started](./guides/getting-started.md) guide
- See [CLAUDE.md](../CLAUDE.md) for architecture rules and patterns
- Open an issue on the project repository
