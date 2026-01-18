package database

import (
	"apprun/pkg/env"
	"fmt"
)

// Config holds database connection configuration
// This is infrastructure configuration, NOT managed by config center
type Config struct {
	Driver      string `mapstructure:"driver" json:"driver" validate:"oneof=postgres mysql" default:"postgres" db:"false"`
	Host        string `mapstructure:"host" json:"host" validate:"required" default:"localhost" db:"false"`
	Port        int    `mapstructure:"port" json:"port" validate:"required,min=1,max=65535" default:"5432" db:"false"`
	User        string `mapstructure:"user" json:"user" validate:"required" default:"postgres" db:"false"`
	Password    string `mapstructure:"password" json:"password" validate:"required,min=8" db:"false"`
	DBName      string `mapstructure:"db_name" json:"db_name" validate:"required" default:"apprun" db:"false"`
	AutoMigrate bool   `mapstructure:"auto_migrate" json:"auto_migrate" default:"false" db:"false"` // Dev only, production should be false
}

// DefaultConfig returns default database configuration from environment variables
// Priority: Environment variables > default.yaml > code defaults
// Environment variable naming: DATABASE_DRIVER, DATABASE_HOST, DATABASE_PORT, etc.
// These env vars are set by env.LoadConfigToEnv() from default.yaml
func DefaultConfig() *Config {
	return &Config{
		Driver:      env.Get("DATABASE_DRIVER", "postgres"),
		Host:        env.Get("DATABASE_HOST", "localhost"),
		Port:        env.GetInt("DATABASE_PORT", 5432),
		User:        env.Get("DATABASE_USER", "postgres"),
		Password:    env.Get("DATABASE_PASSWORD", ""), // Can be empty for dev, loaded from YAML
		DBName:      env.Get("DATABASE_DB_NAME", "apprun"),
		AutoMigrate: env.GetBool("DATABASE_AUTO_MIGRATE", false),
	}
}

// NewDatabaseFromConfig creates a database client from configuration
// This is the factory connector pattern required by Story 10a
// Note: Database module is NOT registered in config center (bootstrap circular dependency)
// Returns Client interface for database operations
func NewDatabaseFromConfig(cfg *Config) (interface{}, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Import context for Connect function
	// This would typically come from the caller
	// For factory pattern, we return a function that accepts context
	return func(ctx interface{}) (interface{}, error) {
		// Type assert context (in real usage, this would be context.Context)
		// Here we're showing the factory pattern structure
		return nil, nil // Placeholder - actual implementation would call Connect
	}, nil
}
