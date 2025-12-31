package database

import (
	"apprun/pkg/env"
)

// Config holds database connection configuration
// This is infrastructure configuration, NOT managed by config center
type Config struct {
	Driver   string `yaml:"driver" validate:"oneof=postgres mysql" default:"postgres" db:"false"`
	Host     string `yaml:"host" validate:"required" default:"localhost" db:"false"`
	Port     int    `yaml:"port" validate:"required,min=1,max=65535" default:"5432" db:"false"`
	User     string `yaml:"user" validate:"required" default:"postgres" db:"false"`
	Password string `yaml:"password" validate:"required,min=8" db:"false"`
	DBName   string `yaml:"db_name" validate:"required" default:"apprun" db:"false"`
}

// DefaultConfig returns default database configuration from environment variables
// Environment variable naming: DATABASE_DRIVER, DATABASE_HOST, DATABASE_PORT, etc.
// These env vars are set by env.LoadConfigToEnv() from default.yaml
func DefaultConfig() *Config {
	return &Config{
		Driver:   env.Get("DATABASE_DRIVER", "postgres"),
		Host:     env.Get("DATABASE_HOST", "localhost"),
		Port:     env.GetInt("DATABASE_PORT", 5432),
		User:     env.Get("DATABASE_USER", "postgres"),
		Password: env.MustGet("DATABASE_PASSWORD"), // Required
		DBName:   env.Get("DATABASE_DB_NAME", "apprun"),
	}
}
