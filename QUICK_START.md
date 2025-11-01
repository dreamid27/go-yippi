# Quick Start Guide

## 🚀 Get Running in 5 Minutes

### 1. Prerequisites Check
```bash
go version    # Should be 1.23+
psql --version    # PostgreSQL 12+
```

### 2. Database Setup
```bash
# Option A: Local PostgreSQL
createdb go-test

# Option B: Docker
docker run --name postgres \
  -e POSTGRES_PASSWORD=adminadmin \
  -e POSTGRES_USER=admin \
  -e POSTGRES_DB=go-test \
  -p 5432:5432 -d postgres:15
```

### 3. Run the API
```bash
# Clone and enter directory
cd go-hex-yippi

# Install dependencies
go mod download

# Run (auto-generates code)
make run
```

### 4. Test the API
Open browser: `http://localhost:8080/docs`

## 📝 First API Call

### Create a User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "age": 30}'
```

### Create a Product
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "PROD-001",
    "slug": "awesome-product",
    "name": "Awesome Product",
    "price": 99.99,
    "description": "An amazing product",
    "weight": 500,
    "length": 20,
    "width": 15,
    "height": 10,
    "status": "draft"
  }'
```

### List Products
```bash
curl http://localhost:8080/products
```

### Publish a Product
```bash
curl -X POST http://localhost:8080/products/1/publish
```

## 🛠️ Common Commands

| Command | Description |
|---------|-------------|
| `make run` | Start the API server |
| `make dev` | Start with hot reload |
| `make build` | Build production binary |
| `make generate` | Generate Ent code |
| `make test` | Run tests |
| `make clean` | Clean build artifacts |

## 🏗️ Project Architecture

```
Domain (Pure Business Logic)
    ↓
Application (Use Cases)
    ↓
Adapters (HTTP, Database)
```

**Key Rule**: Dependencies always point inward!

## 📚 Next Steps

1. **Read the full README**: [README.md](README.md)
2. **Product API Guide**: [PRODUCT_API.md](PRODUCT_API.md)
3. **For AI Assistants**: [CLAUDE.md](CLAUDE.md)
4. **Explore API**: http://localhost:8080/docs

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Change port
export SERVER_PORT=8081
make run
```

### Database Connection Error
```bash
# Verify PostgreSQL is running
psql -U admin -d go-test

# Or check Docker
docker ps | grep postgres
```

### Ent Generation Errors
```bash
# Clean and regenerate
rm -rf internal/adapters/persistence/db/ent
make generate
```

## 💡 Tips

- Use **Swagger UI** at `/docs` for interactive testing
- Check `internal/domain/errors/README.md` for error handling patterns
- Follow the hexagonal architecture principles
- Handlers should never import database packages

## 🎯 Quick Reference

### Environment Variables
```bash
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0
export DB_DSN="host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable"
```

### Project Status
- ✅ User API (CRUD)
- ✅ Product API (with status workflow)
- ✅ Domain errors
- ✅ OpenAPI docs
- 🚧 Authentication (coming soon)

Happy coding! 🎉
