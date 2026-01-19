// Package database provides database migration utilities.
// This module uses Atlas toolchain for schema versioning and migrations.
//
// Dependencies:
//   - Go packages: ariga.io/atlas/sql/{migrate,postgres,schema}
//   - External binary: atlas CLI (required, installed at /usr/local/bin/atlas)
//
// The atlas CLI is a required dependency for:
//   - Schema diff generation (ent:// URL parsing)
//   - Complex schema comparison algorithms
//   - Dev database management (docker:// URLs)
//
// Note: The Go SDK (ariga.io/atlas) provides database operations,
// but schema diff generation requires the atlas CLI binary.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	db                  *sql.DB
	migrationsDir       string
	workingDir          string  // Working directory for Atlas CLI execution
	config              *Config // Database configuration for URL construction
	atlasConfigProvider func(dbURL string) (configPath string, cleanup func(), err error)
}

// NewMigrator creates a new Migrator instance
func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	// Determine working directory (for Atlas CLI execution)
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		workingDir:    workingDir,
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
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Failed to close database after ping failure: %v", closeErr)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseConnectFailed, "failed to ping database")
	}

	// Determine migrations directory
	// If APPRUN_CONFIG_DIR is set, use it as base for relative path
	migrationsDir := "migrations"
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	if configDir := os.Getenv("APPRUN_CONFIG_DIR"); configDir != "" {
		migrationsDir = filepath.Join(configDir, "..", "migrations")
		// Use config directory's parent as working directory
		workingDir = filepath.Join(configDir, "..")
	}

	migrator := NewMigrator(db, migrationsDir)
	migrator.workingDir = workingDir
	migrator.config = cfg // Store config for URL construction

	return migrator, nil
}

// SetAtlasConfigProvider sets the config provider for Atlas CLI operations
func (m *Migrator) SetAtlasConfigProvider(provider func(dbURL string) (configPath string, cleanup func(), err error)) {
	m.atlasConfigProvider = provider
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

	// Create a simple revision read/writer that uses the atlas_schema_revisions table
	// We need to implement a simple version since Atlas doesn't export a ready-to-use implementation
	rrw := &simpleRevisionReadWriter{db: m.db}

	executor, err := migrate.NewExecutor(drv, dir, rrw, migrate.WithAllowDirty(true))
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

// simpleRevisionReadWriter implements migrate.RevisionReadWriter using atlas_schema_revisions table
type simpleRevisionReadWriter struct {
	db *sql.DB
}

func (r *simpleRevisionReadWriter) Ident() *migrate.TableIdent {
	return &migrate.TableIdent{
		Name:   "atlas_schema_revisions",
		Schema: "public",
	}
}

func (r *simpleRevisionReadWriter) ReadRevisions(ctx context.Context) ([]*migrate.Revision, error) {
	// Check if table exists first
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'atlas_schema_revisions'
		)
	`).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT version, description, type, applied, total, executed_at, execution_time, error, error_stmt, hash
		FROM atlas_schema_revisions
		ORDER BY executed_at
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("failed to close rows", logger.Field{Key: "error", Value: err})
		}
	}()

	var revisions []*migrate.Revision
	for rows.Next() {
		rev := &migrate.Revision{}
		var errorStr, errorStmt sql.NullString
		if err := rows.Scan(
			&rev.Version, &rev.Description, &rev.Type, &rev.Applied, &rev.Total,
			&rev.ExecutedAt, &rev.ExecutionTime, &errorStr, &errorStmt, &rev.Hash,
		); err != nil {
			return nil, err
		}
		if errorStr.Valid {
			rev.Error = errorStr.String
		}
		if errorStmt.Valid {
			rev.ErrorStmt = errorStmt.String
		}
		revisions = append(revisions, rev)
	}
	return revisions, rows.Err()
}

func (r *simpleRevisionReadWriter) ReadRevision(ctx context.Context, version string) (*migrate.Revision, error) {
	// Check if table exists first
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'atlas_schema_revisions'
		)
	`).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, migrate.ErrRevisionNotExist
	}

	rev := &migrate.Revision{}
	var errorStr, errorStmt sql.NullString
	err = r.db.QueryRowContext(ctx, `
		SELECT version, description, type, applied, total, executed_at, execution_time, error, error_stmt, hash
		FROM atlas_schema_revisions
		WHERE version = $1
	`, version).Scan(
		&rev.Version, &rev.Description, &rev.Type, &rev.Applied, &rev.Total,
		&rev.ExecutedAt, &rev.ExecutionTime, &errorStr, &errorStmt, &rev.Hash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, migrate.ErrRevisionNotExist
		}
		return nil, err
	}
	if errorStr.Valid {
		rev.Error = errorStr.String
	}
	if errorStmt.Valid {
		rev.ErrorStmt = errorStmt.String
	}
	return rev, nil
}

func (r *simpleRevisionReadWriter) WriteRevision(ctx context.Context, rev *migrate.Revision) error {
	// Create table if not exists
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS atlas_schema_revisions (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT,
			type INTEGER,
			applied INTEGER,
			total INTEGER,
			executed_at TIMESTAMP,
			execution_time BIGINT,
			error TEXT,
			error_stmt TEXT,
			hash TEXT,
			operator_version TEXT
		)
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO atlas_schema_revisions 
		(version, description, type, applied, total, executed_at, execution_time, error, error_stmt, hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (version) DO UPDATE SET
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			applied = EXCLUDED.applied,
			total = EXCLUDED.total,
			executed_at = EXCLUDED.executed_at,
			execution_time = EXCLUDED.execution_time,
			error = EXCLUDED.error,
			error_stmt = EXCLUDED.error_stmt,
			hash = EXCLUDED.hash
	`, rev.Version, rev.Description, rev.Type, rev.Applied, rev.Total,
		rev.ExecutedAt, rev.ExecutionTime, rev.Error, rev.ErrorStmt, rev.Hash)
	return err
}

func (r *simpleRevisionReadWriter) DeleteRevision(ctx context.Context, version string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM atlas_schema_revisions WHERE version = $1", version)
	return err
}

// RollbackMigration rolls back the last applied migration
func (m *Migrator) RollbackMigration(ctx context.Context) error {
	// Get current status to find last applied migration
	status, err := m.Status(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to get migration status")
	}

	if len(status.Applied) == 0 {
		return errors.New(errors.ErrCodeDatabaseMigrateFailed, "no migrations to rollback")
	}

	lastApplied := status.Applied[len(status.Applied)-1]
	logger.Info("rolling back migration", logger.Field{Key: "version", Value: lastApplied})

	// Atlas versioned migrations support down migrations if they exist
	// Look for down migration file
	downFile := filepath.Join(m.migrationsDir, lastApplied+".down.sql")
	if _, err := os.Stat(downFile); os.IsNotExist(err) {
		return errors.New(errors.ErrCodeDatabaseMigrateFailed,
			fmt.Sprintf("no down migration found for version %s. Create %s to enable rollback", lastApplied, downFile))
	}

	// Execute down migration
	downSQL, err := os.ReadFile(downFile)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to read down migration file")
	}

	if _, err := m.db.ExecContext(ctx, string(downSQL)); err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to execute down migration")
	}

	// Remove version from atlas_schema_revisions
	if _, err := m.db.ExecContext(ctx, "DELETE FROM atlas_schema_revisions WHERE version = $1", lastApplied); err != nil {
		logger.Error("failed to update revision table", logger.Field{Key: "error", Value: err.Error()})
		// Continue anyway as the migration was executed
	}

	logger.Info("migration rolled back successfully", logger.Field{Key: "version", Value: lastApplied})
	return nil
}

// ResetDatabase drops all tables and resets migration history
func (m *Migrator) ResetDatabase(ctx context.Context) error {
	logger.Info("resetting database - dropping all tables")

	// Get all tables in public schema
	rows, err := m.db.QueryContext(ctx, `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to list tables")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Error("failed to close rows", logger.Field{Key: "error", Value: closeErr})
		}
	}()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to scan table name")
		}
		tables = append(tables, table)
	}

	// Drop all tables
	for _, table := range tables {
		logger.Info("dropping table", logger.Field{Key: "table", Value: table})
		if _, err := m.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, fmt.Sprintf("failed to drop table %s", table))
		}
	}

	logger.Info("database reset completed successfully")
	return nil
}

// GenerateMigration generates a new migration file by calling atlas CLI.
// This function requires the atlas CLI binary to be installed in PATH.
//
// The atlas CLI is required because:
//   - It natively supports ent:// schema URLs (e.g., ent://ent/schema)
//   - It handles complex schema diff algorithms
//   - It manages dev databases (docker:// URLs) for comparison
//
// The Go SDK (ariga.io/atlas) does not provide schema diff functionality,
// so we must use os/exec to call the atlas CLI binary directly.
func (m *Migrator) GenerateMigration(ctx context.Context, name string, toSchema string, devURL string) (string, error) {
	if name == "" {
		return "", errors.New(errors.ErrCodeInvalidParam, "migration name is required")
	}

	logger.Info("generating migration using Atlas CLI", logger.Field{Key: "name", Value: name})

	// Check if Atlas CLI is available
	if _, err := exec.LookPath("atlas"); err != nil {
		return "", errors.New(errors.ErrCodeDatabaseMigrateFailed,
			"atlas CLI not found. Install it with: curl -sSf https://atlasgo.sh | sh")
	}

	// Get database URL for config generation
	dbURL, err := m.getDatabaseURL()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to get database URL")
	}

	// Get atlas config (embedded or external)
	var configPath string
	var configCleanup func()
	if m.atlasConfigProvider != nil {
		var err error
		configPath, configCleanup, err = m.atlasConfigProvider(dbURL)
		if err != nil {
			return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to get atlas config")
		}
		if configCleanup != nil {
			defer configCleanup()
		}
	}

	// Build atlas migrate diff command
	args := []string{
		"migrate", "diff", name,
		"--dir", fmt.Sprintf("file://%s", m.migrationsDir),
		"--to", toSchema,
		"--dev-url", devURL,
	}

	// Add config flag if available
	if configPath != "" {
		args = append(args, "--config", "file://"+configPath, "--env", "local")
	}

	cmd := exec.CommandContext(ctx, "atlas", args...)
	// Set working directory to project root (where ent/schema is located)
	cmd.Dir = m.workingDir

	logger.Info("executing atlas CLI",
		logger.Field{Key: "workdir", Value: m.workingDir},
		logger.Field{Key: "migrationsDir", Value: m.migrationsDir})

	// Record current file list before running atlas
	filesBefore := make(map[string]bool)
	if entries, err := os.ReadDir(m.migrationsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				filesBefore[entry.Name()] = true
			}
		}
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Check if "no changes" message (regardless of error status)
	if strings.Contains(outputStr, "no changes") ||
		strings.Contains(outputStr, "synced with the desired state") ||
		strings.Contains(outputStr, "migration directory is synced") {
		logger.Info("no schema changes detected")
		return "", errors.New(errors.ErrCodeDatabaseMigrateFailed, "no schema changes detected")
	}

	if err != nil {
		logger.Error("atlas migrate diff failed",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "output", Value: outputStr})
		return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed,
			fmt.Sprintf("atlas migrate diff failed: %s", outputStr))
	}

	// Find the newly generated file (not in filesBefore)
	filesAfter, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrCodeDatabaseMigrateFailed, "failed to read migrations directory")
	}

	var newFile string
	for _, f := range filesAfter {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			if !filesBefore[f.Name()] {
				// This is a new file
				newFile = f.Name()
				break
			}
		}
	}

	if newFile == "" {
		return "", errors.New(errors.ErrCodeDatabaseMigrateFailed, "no migration file was generated")
	}

	fullPath := filepath.Join(m.migrationsDir, newFile)
	logger.Info("migration generated successfully", logger.Field{Key: "file", Value: fullPath})

	return fullPath, nil
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
			defer func() {
				if err := rows.Close(); err != nil {
					log.Printf("Failed to close rows: %v", err)
				}
			}()
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
		// Extract version from filename (e.g., "20260116032122_initial_schema.sql" -> "20260116032122")
		// Version is everything before the first underscore
		parts := strings.SplitN(name, "_", 2)
		if len(parts) > 0 {
			version := strings.TrimSuffix(parts[0], ".sql")
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
	defer func() {
		if err := migrator.Close(); err != nil {
			log.Printf("Failed to close migrator: %v", err)
		}
	}()

	return migrator.ApplyMigrations(ctx)
}

// SchemaDiff represents differences between database and Ent schema
type SchemaDiff struct {
	HasDifferences bool
	Description    string
	SQL            string // Generated SQL to fix differences
}

// InspectSchema inspects the database and compares it with Ent schema
// This uses Atlas's declarative migration mode to detect schema drift
func (m *Migrator) InspectSchema(ctx context.Context, schemaURL, devURL string) (*SchemaDiff, error) {
	if schemaURL == "" {
		schemaURL = "ent://ent/schema"
	}
	if devURL == "" {
		devURL = "docker://postgres/15/dev?search_path=public"
	}

	// Build database URL from connection
	dbURL, err := m.getDatabaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get database URL: %w", err)
	}

	// Add search_path parameter if not already present
	fromURL := dbURL
	if !strings.Contains(fromURL, "search_path=") {
		fromURL = dbURL + "&search_path=public"
	}

	// Get atlas config (embedded or external)
	var configPath string
	var configCleanup func()
	if m.atlasConfigProvider != nil {
		var err error
		configPath, configCleanup, err = m.atlasConfigProvider(dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get atlas config: %w", err)
		}
		if configCleanup != nil {
			defer configCleanup()
		}
	}

	// Use atlas schema diff to compare actual database with desired schema
	args := []string{
		"schema", "diff",
		"--from", fromURL,
		"--to", schemaURL,
		"--dev-url", devURL,
	}

	// Add config flag if available
	if configPath != "" {
		args = append(args, "--config", "file://"+configPath, "--env", "local")
	}

	cmd := exec.CommandContext(ctx, "atlas", args...)
	cmd.Dir = m.workingDir

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Atlas schema diff behavior with ent://
	// - May return exit code 0 even when differences exist
	// - Must check output content first before exit code

	// Filter out Atlas metadata tables from diff output
	// These tables are managed by Atlas SDK and should not be considered schema drift
	outputStr = filterAtlasMetadataTables(outputStr)

	// Check if output contains SQL DDL statements (indicates differences)
	hasDDL := strings.Contains(outputStr, "CREATE TABLE") ||
		strings.Contains(outputStr, "ALTER TABLE") ||
		strings.Contains(outputStr, "DROP TABLE") ||
		strings.Contains(outputStr, "CREATE INDEX") ||
		strings.Contains(outputStr, "DROP INDEX")

	if hasDDL {
		// Schema has differences
		return &SchemaDiff{
			HasDifferences: true,
			Description:    outputStr,
		}, nil
	}

	// Check for sync message
	if strings.Contains(outputStr, "Schemas are synced") {
		return &SchemaDiff{
			HasDifferences: false,
			Description:    "Database schema is in sync with Ent schema",
		}, nil
	}

	// If there's an error and no DDL, it's a real error
	if err != nil {
		return nil, fmt.Errorf("atlas schema diff failed: %s", outputStr)
	}

	// No differences found
	return &SchemaDiff{
		HasDifferences: false,
		Description:    "Database schema is in sync with Ent schema",
	}, nil
}

// RepairSchema generates and optionally applies SQL to fix schema drift
// This is a declarative migration that doesn't create migration files
func (m *Migrator) RepairSchema(ctx context.Context, schemaURL, devURL string, dryRun bool) error {
	if schemaURL == "" {
		schemaURL = "ent://ent/schema"
	}
	if devURL == "" {
		devURL = "docker://postgres/15/dev?search_path=public"
	}

	// First, check if there are any differences and get the SQL
	diff, err := m.InspectSchema(ctx, schemaURL, devURL)
	if err != nil {
		return fmt.Errorf("failed to inspect schema: %w", err)
	}

	if !diff.HasDifferences {
		return nil // Nothing to repair
	}

	// In dry-run mode, the SQL has already been displayed by the caller
	if dryRun {
		return nil
	}

	// Execute the SQL directly
	sqlStatements := diff.Description

	// Execute the SQL
	if _, err := m.db.ExecContext(ctx, sqlStatements); err != nil {
		return fmt.Errorf("failed to execute repair SQL: %w", err)
	}

	return nil
}

// getDatabaseURL constructs a database URL from the connection
func (m *Migrator) getDatabaseURL() (string, error) {
	// Use config if available (preferred method)
	if m.config != nil {
		// Get password from config (already loaded from env or file)
		password := m.config.Password

		// Construct URL using config values
		url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			m.config.User, password, m.config.Host, m.config.Port, m.config.DBName)

		return url, nil
	}

	// Fallback: Query database connection info (legacy method)
	// Note: This method has limitations with IPv6 addresses
	var host, dbname, user string
	var port int

	// Query PostgreSQL system catalogs for connection info
	err := m.db.QueryRow(`
		SELECT 
			COALESCE(inet_server_addr()::text, 'localhost') as host,
			inet_server_port() as port,
			current_database() as dbname,
			current_user as user
	`).Scan(&host, &port, &dbname, &user)

	if err != nil {
		return "", fmt.Errorf("failed to query connection info: %w", err)
	}

	// Get password from environment (Atlas CLI will use it)
	password := os.Getenv("DATABASE_PASSWORD")
	if password == "" {
		password = os.Getenv("POSTGRES_PASSWORD")
	}
	if password == "" {
		password = "dev_password_123" // Development default
	}

	// Handle IPv6 addresses (must be wrapped in brackets)
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	// Construct URL
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbname)

	return url, nil
}

// CleanFailedMigrations removes all failed migration records from history
func (m *Migrator) CleanFailedMigrations(ctx context.Context) (int, error) {
	// Delete failed migrations (applied = 0)
	result, err := m.db.ExecContext(ctx, `
		DELETE FROM atlas_schema_revisions 
		WHERE applied = 0
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete failed migrations: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return int(count), nil
}

// CleanLastMigration removes the last migration record
func (m *Migrator) CleanLastMigration(ctx context.Context) (string, error) {
	// Get last migration version
	var version string
	err := m.db.QueryRowContext(ctx, `
		SELECT version 
		FROM atlas_schema_revisions 
		ORDER BY executed_at DESC 
		LIMIT 1
	`).Scan(&version)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no migrations to clean")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last migration: %w", err)
	}

	// Delete the migration
	if err := m.CleanMigration(ctx, version); err != nil {
		return "", err
	}

	return version, nil
}

// CleanMigration removes a specific migration record by version
func (m *Migrator) CleanMigration(ctx context.Context, version string) error {
	result, err := m.db.ExecContext(ctx, `
		DELETE FROM atlas_schema_revisions 
		WHERE version = $1
	`, version)

	if err != nil {
		return fmt.Errorf("failed to delete migration: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("migration version not found: %s", version)
	}

	return nil
}

// filterAtlasMetadataTables removes Atlas metadata tables from schema diff output
// This prevents 'migrate diff' and 'migrate inspect' from reporting drift on framework-managed tables
func filterAtlasMetadataTables(output string) string {
	// List of Atlas-managed tables that should be ignored in schema comparison
	metadataTables := []string{
		"atlas_schema_revisions", // Migration history tracking table
	}

	lines := strings.Split(output, "\n")
	var filtered []string
	skipBlock := false
	blockDepth := 0

	for i, line := range lines {
		// Check if this line starts a block for a metadata table
		isMetadataTable := false
		for _, table := range metadataTables {
			// Match patterns like: -- Drop "atlas_schema_revisions" table
			if strings.Contains(line, fmt.Sprintf(`"%s"`, table)) &&
				(strings.Contains(line, "DROP TABLE") ||
					strings.Contains(line, "Create") ||
					strings.Contains(line, "Modify")) {
				isMetadataTable = true
				skipBlock = true
				break
			}
			// Match patterns like: DROP TABLE "atlas_schema_revisions";
			if strings.Contains(line, fmt.Sprintf(`DROP TABLE "%s"`, table)) {
				isMetadataTable = true
				// This is a single line statement, skip it
				continue
			}
		}

		if isMetadataTable && !strings.HasPrefix(strings.TrimSpace(line), "--") {
			// Found the actual DDL statement for metadata table, skip it
			continue
		}

		if skipBlock {
			// Track SQL statement blocks (they end with semicolon)
			if strings.Contains(line, "(") {
				blockDepth++
			}
			if strings.Contains(line, ")") {
				blockDepth--
			}
			if strings.Contains(line, ";") && blockDepth <= 0 {
				skipBlock = false
				blockDepth = 0
			}
			continue
		}

		// Keep empty lines between statements for readability
		if i > 0 && strings.TrimSpace(line) == "" && len(filtered) > 0 && strings.TrimSpace(filtered[len(filtered)-1]) == "" {
			continue // Skip duplicate empty lines
		}

		filtered = append(filtered, line)
	}

	// Clean up trailing empty lines
	result := strings.Join(filtered, "\n")
	result = strings.TrimSpace(result)

	return result
}
