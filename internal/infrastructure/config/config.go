package config

import (
	"log"
	"os"
	"strconv"

	"example.com/go-yippi/internal/infrastructure/logging"
	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	MinIO          MinIOConfig
	Storage        StorageConfig
	JWT            JWTConfig
	OpenTelemetry OpenTelemetryConfig
	Logging        logging.LogConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type StorageConfig struct {
	Backend string // "database" or "minio"
}

type JWTConfig struct {
	Secret string
}

type OpenTelemetryConfig struct {
	ServiceName        string
	CollectorURL       string
	InsecureMode       bool
	Enabled            bool
}

// Load loads configuration from environment or files
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "postgres"),
			DSN:    getEnv("DB_DSN", "host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable"),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "go-yippi"),
		},
		Storage: StorageConfig{
			Backend: getEnv("STORAGE_BACKEND", "database"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		},
		OpenTelemetry: OpenTelemetryConfig{
			ServiceName:  getEnv("OTEL_SERVICE_NAME", "go-yippi-api"),
			CollectorURL: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			InsecureMode: getEnvBool("OTEL_INSECURE_MODE", true),
			Enabled:      getEnvBool("OTEL_ENABLED", false),
		},
		Logging: logging.LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return fallback
}
