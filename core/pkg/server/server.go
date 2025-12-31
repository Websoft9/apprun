package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apprun/pkg/env"
)

// Config holds HTTP/HTTPS server configuration
// This is infrastructure configuration, NOT managed by config center
// Values should be provided via environment variables at startup
type Config struct {
	HTTPPort    string `yaml:"http_port" validate:"required,min=1,max=5" default:"8080" db:"false"`
	HTTPSPort   string `yaml:"https_port" validate:"required,min=1,max=5" default:"8443" db:"false"`
	SSLCertFile string `yaml:"ssl_cert_file" validate:"omitempty,file" default:"" db:"false"`
	SSLKeyFile  string `yaml:"ssl_key_file" validate:"omitempty,file" default:"" db:"false"`

	// Graceful shutdown timeout (e.g., "30s", "1m")
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" validate:"required,min=1s" default:"30s" db:"false"`

	// Enable HTTP server even when HTTPS is enabled (for health checks)
	EnableHTTPWithHTTPS bool `yaml:"enable_http_with_https" default:"true" db:"false"`
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

// Start starts the HTTP/HTTPS server with graceful shutdown support
func Start(router http.Handler, cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Check if TLS is enabled
	enableTLS := cfg.SSLCertFile != "" && cfg.SSLKeyFile != ""

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Channel to listen for errors
	serverErrors := make(chan error, 1)

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	if enableTLS {
		// Start HTTPS server
		httpsServer := &http.Server{
			Addr:    ":" + cfg.HTTPSPort,
			Handler: router,
		}

		log.Printf("🔒 Starting HTTPS server on :%s", cfg.HTTPSPort)
		log.Printf("📄 Using certificate: %s", cfg.SSLCertFile)

		// Start HTTPS in goroutine
		go func() {
			if err := httpsServer.ListenAndServeTLS(cfg.SSLCertFile, cfg.SSLKeyFile); err != nil && err != http.ErrServerClosed {
				serverErrors <- fmt.Errorf("HTTPS server error: %w", err)
			}
		}()

		// Optionally start HTTP server for health checks
		if cfg.EnableHTTPWithHTTPS {
			log.Printf("🌐 Starting HTTP server on :%s (for health checks)", cfg.HTTPPort)
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					serverErrors <- fmt.Errorf("HTTP server error: %w", err)
				}
			}()
		}

		// Wait for shutdown signal or error
		select {
		case err := <-serverErrors:
			return err
		case sig := <-shutdown:
			log.Printf("📊 Received signal: %v, starting graceful shutdown...", sig)

			// Create context for graceful shutdown
			ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()

			// Shutdown both servers
			if err := httpsServer.Shutdown(ctx); err != nil {
				log.Printf("⚠️  HTTPS server shutdown error: %v", err)
			}
			if cfg.EnableHTTPWithHTTPS {
				if err := httpServer.Shutdown(ctx); err != nil {
					log.Printf("⚠️  HTTP server shutdown error: %v", err)
				}
			}

			log.Println("✅ Server gracefully stopped")
			return nil
		}
	} else {
		// Start HTTP only
		log.Printf("🌐 Starting HTTP server on :%s", cfg.HTTPPort)
		log.Printf("💡 Tip: Set SSL_CERT_FILE and SSL_KEY_FILE to enable HTTPS")

		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrors <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()

		// Wait for shutdown signal or error
		select {
		case err := <-serverErrors:
			return err
		case sig := <-shutdown:
			log.Printf("📊 Received signal: %v, starting graceful shutdown...", sig)

			ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("server shutdown error: %w", err)
			}

			log.Println("✅ Server gracefully stopped")
			return nil
		}
	}
}

// StartWithDefaults starts the server with default configuration
func StartWithDefaults(router http.Handler) error {
	return Start(router, DefaultConfig())
}
