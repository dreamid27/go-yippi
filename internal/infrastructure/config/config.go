package config

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
	// TODO: Implement configuration loading
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Driver: "sqlite3",
			DSN:    "file:ent?mode=memory&cache=shared&_fk=1",
		},
	}
}
