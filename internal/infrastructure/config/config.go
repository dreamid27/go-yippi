package config

import "os"

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

// Load loads configuration from environment or files
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "postgres"),
			DSN:    getEnv("DB_DSN", "host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
