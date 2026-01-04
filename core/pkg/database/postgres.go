package database

import (
	"context"
	"fmt"

	"apprun/ent"
	"apprun/pkg/errors"

	_ "github.com/lib/pq"
)

// Connect establishes a database connection using the provided configuration
// It also runs schema migration automatically
func Connect(ctx context.Context, cfg *Config) (Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	// Open connection
	client, err := ent.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "Failed to open database connection")
	}

	// Run schema migration
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "Failed to create schema")
	}

	return &entClient{client: client}, nil
}
