package database

import (
	"context"
	"fmt"

	"apprun/ent"

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
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Run schema migration
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &entClient{client: client}, nil
}
