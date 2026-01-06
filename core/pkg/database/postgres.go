package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"apprun/ent"
	"apprun/pkg/errors"

	_ "github.com/lib/pq"
)

// Connect establishes a database connection using the provided configuration
// Note: This function only establishes the connection. For schema migrations,
// use the Migrator from migrate.go or run migrations via CLI/CI/CD.
func Connect(ctx context.Context, cfg *Config) (Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	// First, verify connection is reachable using database/sql directly
	// This is necessary because ent.Open() does lazy connection
	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "Failed to open database connection")
	}

	// Ping to verify connection is alive
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Failed to close database after ping failure: %v", closeErr)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "Failed to connect to database")
	}
	if err := db.Close(); err != nil {
		log.Printf("Warning: Failed to close test connection: %v", err)
	} // Close the test connection

	// Now open with Ent client
	client, err := ent.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "Failed to create ent client")
	}

	return &entClient{client: client}, nil
}
