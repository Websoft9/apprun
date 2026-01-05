// Package database provides database migration utilities.
// This module encapsulates Atlas migration logic for schema versioning.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"apprun/pkg/errors"
	"apprun/pkg/logger"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
)

// MigrationStatus represents the current migration state
type MigrationStatus struct {
	Current   string   // Current version
	Pending   []string // Pending migration files
	Applied   []string // Already applied migrations
	IsCurrent bool     // True if no pending migrations
}

// Migrator handles database schema migrations
type Migrator struct {
	db            *sql.DB
	migrationsDir string
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// NewMigratorFromConfig creates a Migrator from database config
func NewMigratorFromConfig(ctx context.Context, cfg *Config) (*Migrator, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "failed to open database for migration")
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "failed to ping database")
	}

	return NewMigrator(db, "migrations"), nil
}

// ApplyMigrations applies all pending migrations
func (m *Migrator) ApplyMigrations(ctx context.Context) error {
	dir, err := migrate.NewLocalDir(m.migrationsDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to open migrations directory")
	}

	drv, err := postgres.Open(m.db)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to create postgres driver")
	}

	// Create executor with NopRevisionReadWriter for now
	// TODO: Implement proper revision tracking with atlas_schema_revisions table
	executor, err := migrate.NewExecutor(drv, dir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to create migration executor")
	}

	// Log migration start
	logger.Info("starting database migration", logger.Field{Key: "dir", Value: m.migrationsDir})

	if err := executor.ExecuteN(ctx, -1); err != nil {
		logger.Error("migration failed", logger.Field{Key: "error", Value: err.Error()})
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to apply migrations")
	}

	logger.Info("database migration completed successfully")
	return nil
}

// RollbackMigration rolls back the last applied migration
func (m *Migrator) RollbackMigration(ctx context.Context) error {
	// Atlas versioned migrations don't have built-in rollback
	// For now, we document this limitation
	return errors.New(errors.ErrCodeDatabaseMigrateFailed, "rollback requires manual intervention or down migration file")
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) (*MigrationStatus, error) {
	dir, err := migrate.NewLocalDir(m.migrationsDir)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to open migrations directory")
	}

	files, err := dir.Files()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to read migration files")
	}

	status := &MigrationStatus{
		Pending:   make([]string, 0),
		Applied:   make([]string, 0),
		IsCurrent: true,
	}

	// Get applied migrations from database (atlas_schema_revisions table)
	appliedVersions := make(map[string]bool)
	if m.db != nil {
		rows, err := m.db.QueryContext(ctx, "SELECT version FROM atlas_schema_revisions ORDER BY version")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var version string
				if err := rows.Scan(&version); err == nil {
					appliedVersions[version] = true
					status.Applied = append(status.Applied, version)
				}
			}
		}
		// If table doesn't exist, all migrations are pending (first run)
	}

	// Determine pending migrations
	for _, f := range files {
		name := f.Name()
		// Extract version from filename (e.g., "001_create_users.sql" -> "001")
		if len(name) >= 3 {
			version := name[:3]
			if !appliedVersions[version] {
				status.Pending = append(status.Pending, name)
			}
		}
	}

	status.IsCurrent = len(status.Pending) == 0
	if len(status.Applied) > 0 {
		status.Current = status.Applied[len(status.Applied)-1]
	}

	return status, nil
}

// Close closes the database connection
func (m *Migrator) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// AutoMigrateIfEnabled checks config.AutoMigrate and runs migrations if enabled
// This is a convenience function for development environments
func AutoMigrateIfEnabled(ctx context.Context, cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if !cfg.AutoMigrate {
		return nil // Auto-migrate disabled (production default)
	}

	migrator, err := NewMigratorFromConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer migrator.Close()

	return migrator.ApplyMigrations(ctx)
}

// GenerateMigration generates a new migration file from Ent schema diff
// This is typically called via CLI: apprun-admin migrate generate <name>
func GenerateMigration(ctx context.Context, name string, migrationsDir string) (string, error) {
	if name == "" {
		return "", errors.New(errors.ErrCodeInvalidParam, "migration name is required")
	}

	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	// Find next version number by counting only .sql files
	files, err := os.ReadDir(migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to read migrations directory")
	}

	sqlCount := 0
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlCount++
		}
	}

	version := sqlCount + 1
	filename := fmt.Sprintf("%03d_%s.sql", version, name)
	filepath := filepath.Join(migrationsDir, filename)

	// Create empty migration file with header
	content := fmt.Sprintf(`-- Migration: %s
-- Generated: AUTO
-- Version: %03d

-- Add your migration SQL here

`, name, version)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to create migration file")
	}

	return filepath, nil
}
