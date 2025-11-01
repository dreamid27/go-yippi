package main

import (
	"context"
	"fmt"
	"log"

	"example.com/go-yippi/internal/api/handlers"
	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/infrastructure/adapters/persistence"
	"example.com/go-yippi/internal/infrastructure/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/infrastructure/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Ent client
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed opening connection to database: %v", err)
	}
	defer client.Close()

	// Run auto migration
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New()

	// Initialize Huma API
	humaAPI := humafiber.New(app, huma.DefaultConfig("Go Hexagonal API", "1.0.0"))

	// Dependency injection
	userRepo := persistence.NewUserRepository(client)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	productRepo := persistence.NewProductRepository(client)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Register routes
	userHandler.RegisterRoutes(humaAPI)
	productHandler.RegisterRoutes(humaAPI)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
