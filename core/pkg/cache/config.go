package cache

import (
	"time"

	"apprun/pkg/env"
	"apprun/pkg/errors"
)

// ============================================================================
// Module Constants (Validation Rules & Defaults)
// ============================================================================

const (
	// Default configuration values
	DefaultHost       = "localhost"
	DefaultPort       = "6379"
	DefaultDB         = 0
	DefaultPoolSize   = 10
	DefaultTimeout    = 2 * time.Second
	DefaultMaxRetries = 3
	DefaultFailOpen   = true // Fail-open by default for stateless business logic

	// Environment variable keys
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
	EnvRedisPoolSize = "REDIS_POOL_SIZE"
	EnvRedisTimeout  = "REDIS_TIMEOUT"
	EnvRedisFailOpen = "REDIS_FAIL_STRATEGY" // "open" or "fast"
)

// Config holds Redis cache configuration.
// This is infrastructure configuration, NOT managed by config center (bootstrap dependency)
type Config struct {
	Host       string        `mapstructure:"host" json:"host" validate:"required" default:"localhost" db:"false"`    // Redis server host
	Port       string        `mapstructure:"port" json:"port" validate:"required" default:"6379" db:"false"`         // Redis server port
	Password   string        `mapstructure:"password" json:"password" default:"" db:"false"`                         // Optional password
	DB         int           `mapstructure:"db" json:"db" validate:"min=0,max=15" default:"0" db:"false"`            // Database number (0-15)
	PoolSize   int           `mapstructure:"pool_size" json:"pool_size" validate:"min=1" default:"10" db:"false"`    // Connection pool size
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout" validate:"min=1s" default:"2s" db:"false"`       // Operation timeout
	MaxRetries int           `mapstructure:"max_retries" json:"max_retries" validate:"min=0" default:"3" db:"false"` // Max retry attempts
	FailOpen   bool          `mapstructure:"fail_open" json:"fail_open" default:"true" db:"false"`                   // True = fail-open, False = fail-fast
	TLSEnabled bool          `mapstructure:"tls_enabled" json:"tls_enabled" default:"false" db:"false"`              // Enable TLS connection
}

// DefaultConfig returns the default configuration loaded from environment variables.
func DefaultConfig() *Config {
	timeout := env.GetDuration(EnvRedisTimeout, DefaultTimeout)
	failStrategy := env.Get(EnvRedisFailOpen, "open")

	return &Config{
		Host:       env.Get(EnvRedisHost, DefaultHost),
		Port:       env.Get(EnvRedisPort, DefaultPort),
		Password:   env.Get(EnvRedisPassword, ""),
		DB:         env.GetInt(EnvRedisDB, DefaultDB),
		PoolSize:   env.GetInt(EnvRedisPoolSize, DefaultPoolSize),
		Timeout:    timeout,
		MaxRetries: DefaultMaxRetries,
		FailOpen:   failStrategy == "open",
		TLSEnabled: false,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New(CodeInvalidConfig, "host cannot be empty")
	}
	if c.Port == "" {
		return errors.New(CodeInvalidConfig, "port cannot be empty")
	}
	if c.DB < 0 || c.DB > 15 {
		return errors.New(CodeInvalidConfig, "db must be between 0 and 15")
	}
	if c.PoolSize < 1 {
		return errors.New(CodeInvalidConfig, "pool_size must be at least 1")
	}
	if c.Timeout <= 0 {
		return errors.New(CodeInvalidConfig, "timeout must be positive")
	}
	return nil
}

// Address returns the Redis server address in "host:port" format.
func (c *Config) Address() string {
	return c.Host + ":" + c.Port
}
