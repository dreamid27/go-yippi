package main

import (
	"context"
	"fmt"
	"log"

	"example.com/go-yippi/internal/adapters/api/handlers"
	"example.com/go-yippi/internal/adapters/persistence"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/infrastructure/config"
	"example.com/go-yippi/internal/infrastructure/logging"
	"example.com/go-yippi/internal/infrastructure/telemetry"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logging
	if err := logging.Initialize(&cfg.Logging); err != nil {
		logging.GetGlobalLogger().Fatal("failed to initialize logging",
			zap.Error(err))
	}
	defer logging.Sync()

	// Initialize OpenTelemetry Tracing
	tracerCleanup := telemetry.InitTracer(&cfg.OpenTelemetry)
	defer tracerCleanup(context.Background())

	// Initialize OpenTelemetry Metrics
	metricsCleanup := telemetry.InitMetrics(&cfg.OpenTelemetry)
	defer metricsCleanup(context.Background())

	// Initialize Ent client
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		logging.GetGlobalLogger().Fatal("failed opening connection to database",
			zap.Error(err))
	}
	defer client.Close()

	// Wrap with tracing if enabled
	var dbClient *ent.Client
	if telemetry.IsEnabled(&cfg.OpenTelemetry) {
		tracedClient := telemetry.NewTracedEntClient(client)
		dbClient = tracedClient.Intercept().Client
		log.Println("Database tracing enabled")
	} else {
		dbClient = client
	}

	// Run auto migration
	if err := dbClient.Schema.Create(context.Background()); err != nil {
		logging.GetGlobalLogger().Fatal("failed creating schema resources",
			zap.Error(err))
	}

	// Initialize Fiber app
	app := fiber.New()

	// Add logging middleware
	app.Use(logging.RequestIDMiddleware())
	app.Use(logging.HTTPLoggerMiddleware())

	// Add security logging middleware
	app.Use(logging.SecurityLoggerMiddleware())

	// Add audit logging middleware
	app.Use(logging.AuditLoggerMiddleware())

	// Add enhanced OpenTelemetry middleware if enabled
	if telemetry.IsEnabled(&cfg.OpenTelemetry) {
		app.Use(telemetry.HTTPMiddleware(cfg.OpenTelemetry.ServiceName))
	}

	// Initialize Huma API with custom config for Scalar docs
	humaConfig := huma.DefaultConfig("Go Hexagonal API", "1.0.0")
	humaConfig.DocsPath = "" // Disable default docs to use Scalar instead
	humaAPI := humafiber.New(app, humaConfig)

	// Add custom /docs route for Scalar API documentation
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		html := `<!doctype html>
<html>
  <head>
    <title>API Reference - Go Hexagonal API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
		return c.SendString(html)
	})

	// Dependency injection
	userRepo := persistence.NewUserRepository(dbClient)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	categoryRepo := persistence.NewCategoryRepository(dbClient)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	productRepo := persistence.NewProductRepository(dbClient)
	productService := services.NewProductService(productRepo, categoryRepo)
	productHandler := handlers.NewProductHandler(productService)

	brandRepo := persistence.NewBrandRepository(dbClient)
	brandService := services.NewBrandService(brandRepo)
	brandHandler := handlers.NewBrandHandler(brandService)

	// Initialize MinIO client (infrastructure)
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})

	if err != nil {
		logging.GetGlobalLogger().Fatal("failed to initialize MinIO client",
			zap.Error(err))
	}

	// Ensure default bucket exists
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.MinIO.BucketName)
	if err != nil {
		logging.GetGlobalLogger().Fatal("failed to check if bucket exists",
			zap.Error(err))
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.MinIO.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			logging.GetGlobalLogger().Fatal("failed to create bucket",
				zap.Error(err))
		}
		logging.GetGlobalLogger().Info("Created bucket",
			zap.String("bucket_name", cfg.MinIO.BucketName))
	}

	// Initialize storage repository (adapter)
	storageRepo := persistence.NewMinIOStorageRepository(minioClient, cfg.MinIO.Endpoint, cfg.MinIO.UseSSL)
	// Initialize storage service (application layer)
	storageService := services.NewStorageService(storageRepo, cfg.MinIO.BucketName)
	// Initialize file handler (API adapter)
	fileHandler := handlers.NewFileHandler(storageService)

	// Register Huma routes
	userHandler.RegisterRoutes(humaAPI)
	categoryHandler.RegisterRoutes(humaAPI)
	productHandler.RegisterRoutes(humaAPI)
	brandHandler.RegisterRoutes(humaAPI)
	fileHandler.RegisterRoutes(humaAPI)
	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logging.GetGlobalLogger().Info("Starting server",
		zap.String("address", addr))
	if err := app.Listen(addr); err != nil {
		logging.GetGlobalLogger().Fatal("failed to start server",
			zap.Error(err))
	}
}
