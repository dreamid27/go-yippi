.PHONY: run generate dev air build clean test

# Run the application with automatic generation
run: generate
	go run cmd/api/main.go

# Generate Ent code
generate:
	go run -mod=mod entgo.io/ent/cmd/ent generate --target ./internal/adapters/persistence/db/ent ./internal/adapters/persistence/db/schema

# Development mode with live reload using Air
dev: air

# Run with Air for live reloading
air:
	air

# Build binary
build: generate
	mkdir -p bin
	go build -o bin/api cmd/api/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...
