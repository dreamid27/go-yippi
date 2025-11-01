.PHONY: run generate dev air

# Run the application with automatic generation
run: generate
	go run cmd/api/main.go

# Generate Ent code
generate:
	go run -mod=mod entgo.io/ent/cmd/ent generate --target ./internal/infrastructure/adapters/persistence/db/ent ./internal/infrastructure/adapters/persistence/db/schema

# Development mode with live reload using Air
dev: air

# Run with Air for live reloading
air:
	air
