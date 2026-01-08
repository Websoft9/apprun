// Package server provides HTTP server initialization and lifecycle management.
package server

import (
	"apprun/pkg/errors"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start starts the HTTP/HTTPS server with graceful shutdown support
func Start(router http.Handler, cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Check if TLS is enabled
	enableTLS := cfg.SSLCertFile != "" && cfg.SSLKeyFile != ""

	// Create HTTP server
	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	// Channel to listen for errors
	serverErrors := make(chan error, 1)

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	if enableTLS {
		// Start HTTPS server
		httpsServer := &http.Server{
			Addr:              ":" + cfg.HTTPSPort,
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
		}

		log.Printf("🔒 Starting HTTPS server on :%s", cfg.HTTPSPort)
		log.Printf("📄 Using certificate: %s", cfg.SSLCertFile)

		// Start HTTPS in goroutine
		go func() {
			if err := httpsServer.ListenAndServeTLS(cfg.SSLCertFile, cfg.SSLKeyFile); err != nil && err != http.ErrServerClosed {
				serverErrors <- errors.Wrap(err, errors.ErrCodeServerStartFailed, "HTTPS server error")
			}
		}()

		// Optionally start HTTP server for health checks
		if cfg.EnableHTTPWithHTTPS {
			log.Printf("🌐 Starting HTTP server on :%s (for health checks)", cfg.HTTPPort)
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					serverErrors <- errors.Wrap(err, errors.ErrCodeServerStartFailed, "HTTP server error")
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
				serverErrors <- errors.Wrap(err, errors.ErrCodeServerStartFailed, "HTTP server error")
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
				return errors.Wrap(err, errors.ErrCodeServerShutdownFailed, "Server shutdown error")
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
