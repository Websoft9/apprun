package database

import (
	"context"

	"apprun/ent"
)

// Client is the anti-corruption layer interface for database operations
// It hides Ent ORM implementation details from upper layers
type Client interface {
	// Close closes the database connection
	Close() error

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error

	// Tx executes a function within a transaction
	Tx(ctx context.Context, fn func(tx *ent.Tx) error) error

	// GetEntClient returns the underlying Ent client (use sparingly)
	GetEntClient() *ent.Client
}

// entClient is the concrete implementation of Client interface
type entClient struct {
	client *ent.Client
}

// Close closes the database connection
func (c *entClient) Close() error {
	return c.client.Close()
}

// Ping checks if the database connection is alive
func (c *entClient) Ping(ctx context.Context) error {
	// Ent doesn't expose Ping directly, we can try a simple query
	_, err := c.client.Users.Query().Count(ctx)
	return err
}

// Tx executes a function within a transaction
func (c *entClient) Tx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := c.client.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return rerr
		}
		return err
	}

	return tx.Commit()
}

// GetEntClient returns the underlying Ent client
func (c *entClient) GetEntClient() *ent.Client {
	return c.client
}
