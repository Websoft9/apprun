package main

import (
	"context"
	"log"
	"time"

	_ "apprun/docs" // Swagger docs (自动生成)
	"apprun/modules/config"
	"apprun/pkg/database"
	"apprun/pkg/env"
	"apprun/pkg/logger"
	"apprun/pkg/server"
	"apprun/routes"

	_ "github.com/lib/pq"
)

// @title           AppRun API
// @version         1.0
// @description     AppRun Platform REST API Documentation
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.email   support@websoft9.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api

// @schemes         http https
func main() {
	// Recover from panics during startup (e.g., missing required environment variables)
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("❌ Startup failed: %v", r)
		}
	}()

	// Phase 0: Load infrastructure config from file to environment variables
	// This allows default.yaml to provide defaults while respecting existing env vars
	// Priority: runtime env > config file > code defaults
	configDir := env.Get("CONFIG_DIR", "./config")
	if err := env.LoadConfigToEnv(configDir); err != nil {
		log.Printf("⚠️  Warning: Failed to load config file: %v", err)
		log.Println("⚠️  Using environment variables and code defaults only")
	}

	// Create context with timeout for startup phase
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Phase 1: Initialize Configuration Module Registry
	// Register business module configurations for centralized management
	// Note: Infrastructure configs (server, database) are NOT registered here
	// They are managed via environment variables loaded in Phase 0
	registry := config.NewRegistry()
	if err := registry.Register("logger", &logger.Config{}); err != nil {
		log.Fatalf("❌ Failed to register logger config: %v", err)
	}
	log.Println("✅ Logger module registered with config center")

	// Create configuration bootstrapper with registry
	bootstrap := config.NewBootstrapWithRegistry(env.Get("CONFIG_DIR", "./config"), registry)

	// Phase 2: Connect to Database (Layer 1 infrastructure)
	// Database connection is required for config service and business logic
	// Note: DB_PASSWORD environment variable is required (set via Phase 0 or export)
	dbCfg := database.DefaultConfig()
	dbClient, err := database.Connect(ctx, dbCfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer dbClient.Close()
	log.Println("✅ Database connected")

	// Phase 3: Initialize Config Service (Layer 2 - Configuration Center)
	// Config service manages runtime configurations stored in database
	configService, err := bootstrap.CreateService(ctx, dbClient)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create config service: %v", err)
		log.Println("⚠️  Config API routes will not be registered")
	} else {
		log.Println("✅ Config service initialized with DB support")
	}

	// Phase 4: Initialize Business Logger (Layer 2 - Runtime Logger)
	// Business logger is used for application runtime logging (request handling, business logic)
	// Startup logs continue using standard log package (this is still bootstrap phase)
	loggerCfg := logger.Config{
		Level: logger.LevelInfo, // Default level, can be overridden by config service
		Output: logger.OutputConfig{
			Targets: []string{"stdout"},
		},
	}
	// TODO: Load logger config from config service in future iterations
	// For now, use default config for business logger initialization
	businessLogger, err := logger.NewZapLogger(loggerCfg)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize business logger: %v", err)
		log.Println("⚠️  Using NopLogger (no-op) for runtime logging")
		// Fallback to NopLogger if initialization fails
	} else {
		logger.SetLogger(businessLogger)
		defer businessLogger.Close()
		log.Println("✅ Business logger initialized (runtime logging ready)")
	}

	// Phase 5: Setup HTTP Routes
	// Register all HTTP handlers and middleware
	router := routes.SetupRoutes(configService)
	log.Println("✅ HTTP routes configured")

	// Phase 6: Configure HTTP/HTTPS Server
	// Server configuration is automatically loaded from environment variables by DefaultConfig()
	// Environment variables are set in Phase 0 by LoadConfigToEnv() from default.yaml
	// Naming convention: SERVER_HTTP_PORT, SERVER_HTTPS_PORT, SERVER_SSL_CERT_FILE, etc.
	serverCfg := server.DefaultConfig()

	// Print startup summary (still using standard log - bootstrap phase)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 AppRun Server Starting...")
	log.Printf("   Database: %s@%s:%d/%s", dbCfg.User, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
	log.Printf("   HTTP Port: %s", serverCfg.HTTPPort)
	if serverCfg.SSLCertFile != "" {
		log.Printf("   HTTPS Port: %s (TLS Enabled)", serverCfg.HTTPSPort)
	}
	log.Printf("   Config Dir: %s", env.Get("CONFIG_DIR", "./config"))
	log.Printf("   Logger Level: %s", loggerCfg.Level)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📝 Note: Using standard log for startup, business logger for runtime")

	// Phase 7: Start HTTP/HTTPS Server (enters runtime phase)
	// From this point, handlers will use logger.L() for business logging
	if err := server.Start(router, serverCfg); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
