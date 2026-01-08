// Package server provides HTTP server initialization and lifecycle management.
package server

import (
	"net/http"
	"time"

	"apprun/pkg/env"
)

// Config holds HTTP/HTTPS server configuration
// This is infrastructure configuration, NOT managed by config center
// Values should be provided via environment variables at startup
type Config struct {
	HTTPPort            string        `yaml:"http_port" validate:"required,min=1,max=5" default:"8080" db:"false"`
	HTTPSPort           string        `yaml:"https_port" validate:"required,min=1,max=5" default:"8443" db:"false"`
	SSLCertFile         string        `yaml:"ssl_cert_file" validate:"omitempty,file" default:"" db:"false"`
	SSLKeyFile          string        `yaml:"ssl_key_file" validate:"omitempty,file" default:"" db:"false"`
	ShutdownTimeout     time.Duration `yaml:"shutdown_timeout" validate:"required,min=1s" default:"30s" db:"false"`
	EnableHTTPWithHTTPS bool          `yaml:"enable_http_with_https" default:"true" db:"false"`
}

// DefaultConfig returns default server configuration
// Configuration is loaded from environment variables with SERVER_ prefix
// Environment variable naming: SERVER_HTTP_PORT, SERVER_HTTPS_PORT, etc.
// These env vars are set by env.LoadConfigToEnv() from default.yaml
func DefaultConfig() *Config {
	return &Config{
		HTTPPort:            env.Get("SERVER_HTTP_PORT", "8080"),
		HTTPSPort:           env.Get("SERVER_HTTPS_PORT", "8443"),
		SSLCertFile:         env.Get("SERVER_SSL_CERT_FILE", ""),
		SSLKeyFile:          env.Get("SERVER_SSL_KEY_FILE", ""),
		ShutdownTimeout:     env.GetDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		EnableHTTPWithHTTPS: env.GetBool("SERVER_ENABLE_HTTP_WITH_HTTPS", true),
	}
}

// NewServerFromConfig creates a server starter function from config
// This is the factory connector pattern required by Story 10a
// Returns a function that accepts a router and starts the server
func NewServerFromConfig(cfg *Config) func(router http.Handler) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return func(router http.Handler) error {
		return Start(router, cfg)
	}
}
