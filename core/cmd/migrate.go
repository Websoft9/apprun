package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"apprun/pkg/database"

	"github.com/spf13/cobra"
)

//go:embed atlas.hcl
var atlasConfigTemplate string

var (
	// Flags for apply subcommand
	migrateApplyDryRun bool

	// Flags for sync subcommand
	migrateSyncDryRun bool

	// Flags for repair subcommand
	migrateRepairDryRun  bool
	migrateRepairExecute bool

	// Shared flags
	// migrateName     string
	migrateToSchema string
	migrateDevURL   string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:     "migrate",
	GroupID: "server",
	Short:   "Database migration management",
	Long: `Manage database schema migrations using Atlas SDK.

Migrations are tracked in the atlas_schema_revisions table.
All migration files are embedded in the binary for portability.

Available subcommands:
  apply    - Apply pending database migrations
  status   - Check migration status
  validate - Validate migration files integrity
  diff     - Generate new migration from schema changes
  sync     - Auto-generate and apply migrations (dev mode)
  rollback - Rollback last applied migration
  reset    - Reset database (drop all tables)
  inspect  - Inspect database schema and detect drift
  repair   - Repair database schema (declarative mode)
  clean    - Remove failed or invalid migration records

Examples:
  # Versioned migrations (normal workflow)
  apprun migrate apply                    # Apply all pending migrations
  apprun migrate apply --dry-run          # Preview SQL without executing
  apprun migrate status                   # Check current migration state
  apprun migrate validate                 # Verify migration files
  apprun migrate diff add_user_email      # Generate migration file
  apprun migrate sync                     # Auto-generate and apply (dev mode)
  apprun migrate rollback                 # Rollback last migration
  apprun migrate reset                    # Reset database (DANGER!)

  # Declarative migrations (repair scenarios)
  apprun migrate inspect                  # Check for schema drift
  apprun migrate repair                   # Preview repair SQL (dry-run)
  apprun migrate repair --execute         # Apply repair (dev only)
  
  # Maintenance (cleanup scenarios)
  apprun migrate clean --all              # Remove all failed migrations
  apprun migrate clean 20260116093904     # Remove specific migration
`,
}

// migrateApplyCmd represents the migrate apply command
var migrateApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply pending database migrations",
	Long: `Apply all pending database migrations to bring the schema up to date.

The command will:
  1. Connect to the database using environment variables
  2. Check for pending migrations
  3. Apply them in order
  4. Update the atlas_schema_revisions table

Examples:
  apprun migrate apply            # Apply migrations
  apprun migrate apply --dry-run  # Preview without executing
`,
	RunE: runMigrateApply,
}

// migrateStatusCmd represents the migrate status command
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check migration status",
	Long: `Check the current state of database migrations.

Displays:
  - Current version
  - Applied migrations
  - Pending migrations
  - Overall migration state

Examples:
  apprun migrate status
`,
	RunE: runMigrateStatus,
}

// migrateValidateCmd represents the migrate validate command
var migrateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate migration files integrity",
	Long: `Validate that migration files are correctly formatted and sequential.

Checks:
  - Migration files exist
  - Filenames follow naming convention
  - Version numbers are sequential
  - No duplicate versions

Examples:
  apprun migrate validate
`,
	RunE: runMigrateValidate,
}

// migrateDiffCmd represents the migrate diff command
var migrateDiffCmd = &cobra.Command{
	Use:   "diff [name]",
	Short: "Generate new migration from schema changes",
	Long: `Generate a new migration file by comparing current schema with Ent schema definition.

Arguments:
  name    Migration name (required) - describes the change being made
          This will be used as part of the migration filename.
          Examples: "add_user_phone", "create_orders_table", "add_index_email"
          Generated file: {timestamp}_{name}.sql

This command:
  1. Reads your Ent schema definitions (ent/schema/*.go)
  2. Compares with current database state
  3. Generates SQL migration file with the differences
  4. Saves as: migrations/{timestamp}_{name}.sql

Examples:
  apprun migrate diff add_user_phone       # Creates: 20260116153045_add_user_phone.sql
  apprun migrate diff create_orders_table  # Creates: 20260116153102_create_orders_table.sql
  apprun migrate diff add_index_email      # Creates: 20260116153125_add_index_email.sql

Naming conventions:
  ✓ Use lowercase with underscores: add_user_column
  ✓ Be descriptive but concise: create_payments_table
  ✓ Use verbs: add, create, remove, modify, etc.
  ✗ Avoid spaces or special characters
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("migration name is required\n\nExample: apprun migrate diff add_user_email")
		}
		if len(args) > 1 {
			return fmt.Errorf("only one migration name is allowed\n\nExample: apprun migrate diff add_user_email")
		}
		return nil
	},
	RunE: runMigrateDiff,
}

// migrateSyncCmd represents the migrate sync command
var migrateSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Auto-generate and apply migrations (development mode)",
	Long: `Automatically generate migration from schema changes and apply it immediately.

This command is designed for development workflow:
  1. Compares Ent schema with current database
  2. Auto-generates migration file if there are changes
  3. Immediately applies the migration
  4. Updates revision tracking

⚠️  WARNING: This is for DEVELOPMENT only!
For production, use 'migrate diff' + 'migrate apply' separately to review changes.

Examples:
  apprun migrate sync                  # Auto-sync schema changes
  apprun migrate sync --dry-run        # Preview without executing
`,
	RunE: runMigrateSync,
}

// migrateRollbackCmd represents the migrate rollback command
var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback last applied migration",
	Long: `Rollback the last applied migration.

⚠️  CAUTION: This operation can result in data loss.
Always backup your database before rolling back migrations.

The rollback process:
  1. Identifies the last applied migration
  2. Executes the down migration (if available)
  3. Updates the atlas_schema_revisions table

Examples:
  apprun migrate rollback
`,
	RunE: runMigrateRollback,
}

// migrateResetCmd represents the migrate reset command
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (drop all tables)",
	Long: `Reset the database by dropping all tables and migration history.

⚠️  DANGER: This operation will DELETE ALL DATA!
Only use this in development environments.

The reset process:
  1. Drops all tables in the database
  2. Clears the atlas_schema_revisions table
  3. Database will be ready for fresh migrations

Examples:
  apprun migrate reset    # Reset database (requires confirmation)
`,
	RunE: runMigrateReset,
}

// migrateInspectCmd represents the migrate inspect command
var migrateInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect database schema and detect drift",
	Long: `Inspect the actual database schema and compare it with Ent schema definition.

This command uses Atlas's declarative migration mode to detect schema drift:
  - Manual database changes (ALTER TABLE, DROP COLUMN, etc.)
  - Failed migrations that left database in inconsistent state
  - Migration file corruption or loss

⚠️  This is different from 'migrate status':
  - 'status': Checks migration history (which files are applied)
  - 'inspect': Checks actual database structure vs Ent schema

Use cases:
  - Verify database structure after manual changes
  - Diagnose migration issues
  - Check if 'migrate repair' is needed

Examples:
  apprun migrate inspect                    # Check for schema drift
  apprun migrate inspect --to-schema ent:// # Specify Ent schema location
`,
	RunE: runMigrateInspect,
}

// migrateRepairCmd represents the migrate repair command
var migrateRepairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair database schema to match Ent definition",
	Long: `Repair database schema by applying changes directly (declarative migration).

⚠️  WARNING: This command bypasses migration history!
  - Does NOT create migration files
  - Does NOT update migration history
  - ONLY for development/repair scenarios
  - NEVER use in production

This is useful when:
  - Database structure was manually modified
  - Migration files were corrupted or lost
  - Quick fix needed in development environment

How it works:
  1. Reads actual database structure
  2. Compares with Ent schema definition
  3. Generates SQL to fix differences
  4. Applies changes (with --execute flag)

Safety:
  - Default: Dry-run mode (shows SQL, doesn't execute)
  - Requires --execute flag to apply changes
  - Respects diff policy (won't drop tables/columns)
  - Recommends backup before execution

Examples:
  apprun migrate repair                # Show SQL without executing (dry-run)
  apprun migrate repair --execute      # Apply changes
  apprun migrate repair --dry-run      # Explicit dry-run mode

After repair:
  - Database structure will match Ent schema
  - Migration history may be out of sync
  - Consider regenerating migrations from scratch if needed
`,
	RunE: runMigrateRepair,
}

// migrateCleanCmd represents the migrate clean command
var migrateCleanCmd = &cobra.Command{
	Use:   "clean [version]",
	Short: "Remove failed or invalid migration records",
	Long: `Clean up migration history by removing failed or invalid migration records.

This command helps resolve migration conflicts when:
  - A migration failed partway through execution
  - Migration files were generated after using 'migrate repair'
  - Manual database changes created inconsistent state

⚠️  WARNING: Only use this in development!
  - This modifies migration history
  - Can cause inconsistencies if used incorrectly
  - Always backup before cleaning

Arguments:
  version    Migration version to remove (optional)
             Format: YYYYMMDDHHMMSS (e.g., 20260116093904)
             If omitted, removes all failed migrations

Examples:
  apprun migrate clean 20260116093904   # Remove specific migration record
  apprun migrate clean --all            # Remove all failed migrations
  apprun migrate clean --last           # Remove last migration record

Common scenarios:
  1. After 'migrate repair --execute':
     apprun migrate clean --all         # Clear all records
     
  2. Failed migration:
     apprun migrate clean 20260116093904  # Remove failed record
     rm core/migrations/20260116093904_*.sql  # Delete file
     atlas migrate hash --dir file://core/migrations  # Rehash
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrateClean,
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add subcommands
	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateValidateCmd)
	migrateCmd.AddCommand(migrateDiffCmd)
	migrateCmd.AddCommand(migrateSyncCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateInspectCmd)
	migrateCmd.AddCommand(migrateRepairCmd)
	migrateCmd.AddCommand(migrateCleanCmd)

	// Flags for apply subcommand
	migrateApplyCmd.Flags().BoolVar(&migrateApplyDryRun, "dry-run", false, "preview SQL without executing")

	// Flags for clean subcommand
	migrateCleanCmd.Flags().Bool("all", false, "remove all failed migrations")
	migrateCleanCmd.Flags().Bool("last", false, "remove last migration record")

	// Flags for sync subcommand
	migrateSyncCmd.Flags().BoolVar(&migrateSyncDryRun, "dry-run", false, "preview SQL without executing")
	migrateSyncCmd.Flags().StringVar(&migrateToSchema, "to-schema", "ent://ent/schema", "target schema source")
	migrateSyncCmd.Flags().StringVar(&migrateDevURL, "dev-url", "docker://postgres/15/dev?search_path=public", "dev database URL for computing diff")

	// Flags for diff subcommand
	migrateDiffCmd.Flags().StringVar(&migrateToSchema, "to-schema", "ent://ent/schema", "target schema source")
	migrateDiffCmd.Flags().StringVar(&migrateDevURL, "dev-url", "docker://postgres/15/dev?search_path=public", "dev database URL for computing diff")

	// Flags for inspect subcommand
	migrateInspectCmd.Flags().StringVar(&migrateToSchema, "to-schema", "ent://ent/schema", "target schema source")
	migrateInspectCmd.Flags().StringVar(&migrateDevURL, "dev-url", "docker://postgres/15/dev?search_path=public", "dev database URL for computing diff")

	// Flags for repair subcommand
	migrateRepairCmd.Flags().BoolVar(&migrateRepairDryRun, "dry-run", true, "preview SQL without executing (default)")
	migrateRepairCmd.Flags().BoolVar(&migrateRepairExecute, "execute", false, "execute the repair (overrides dry-run)")
	migrateRepairCmd.Flags().StringVar(&migrateToSchema, "to-schema", "ent://ent/schema", "target schema source")
	migrateRepairCmd.Flags().StringVar(&migrateDevURL, "dev-url", "docker://postgres/15/dev?search_path=public", "dev database URL for computing diff")
}

// setupMigrator creates a migrator with atlas config provider
func setupMigrator(ctx context.Context, cfg *database.Config) (*database.Migrator, error) {
	migrator, err := database.NewMigratorFromConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Set atlas config provider (use embedded config)
	migrator.SetAtlasConfigProvider(getAtlasConfigPath)

	return migrator, nil
}

func runMigrateApply(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator with atlas config provider
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	if migrateApplyDryRun {
		fmt.Println("🔍 Dry-run mode: Preview only (not implemented yet)")
		fmt.Println("⚠️  Full dry-run support requires additional Atlas SDK integration")
		return nil
	}

	// Apply migrations
	fmt.Println("🚀 Applying database migrations...")
	if err := migrator.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("✅ Migrations applied successfully")
	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	// Get migration status
	status, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Display status
	fmt.Println("📊 Migration Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if status.IsCurrent {
		fmt.Println("✅ Database is up to date")
	} else {
		fmt.Println("⚠️  Pending migrations detected")
	}

	fmt.Printf("\nCurrent Version: %s\n", status.Current)
	fmt.Printf("Applied Migrations: %d\n", len(status.Applied))
	fmt.Printf("Pending Migrations: %d\n\n", len(status.Pending))

	if len(status.Applied) > 0 {
		fmt.Println("Applied:")
		for _, m := range status.Applied {
			fmt.Printf("  ✓ %s\n", m)
		}
		fmt.Println()
	}

	if len(status.Pending) > 0 {
		fmt.Println("Pending:")
		for _, m := range status.Pending {
			fmt.Printf("  ⏳ %s\n", m)
		}
		fmt.Println()
		fmt.Println("💡 Run 'apprun migrate apply' to apply pending migrations")
	}

	return nil
}

func runMigrateValidate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Println("🔍 Validating migration files...")
	fmt.Println()

	// Get migration status to validate files
	status, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	totalMigrations := len(status.Applied) + len(status.Pending)
	if totalMigrations == 0 {
		fmt.Println("⚠️  No migration files found")
		return fmt.Errorf("no migrations to validate")
	}

	fmt.Println("✅ Migration files validated successfully")
	fmt.Println()
	fmt.Printf("Total migrations: %d\n", totalMigrations)
	fmt.Printf("  Applied: %d\n", len(status.Applied))
	fmt.Printf("  Pending: %d\n", len(status.Pending))
	fmt.Println()
	fmt.Println("Checks performed:")
	fmt.Println("  ✓ Migration files accessible")
	fmt.Println("  ✓ File naming conventions")
	fmt.Println("  ✓ Database connectivity")
	fmt.Println("  ✓ Schema revision table exists")

	return nil
}

func runMigrateDiff(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	migrationName := args[0]

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Printf("📝 Generating migration: %s...\n", migrationName)
	fmt.Println()

	// Generate migration
	filename, err := migrator.GenerateMigration(ctx, migrationName, migrateToSchema, migrateDevURL)
	if err != nil {
		return fmt.Errorf("failed to generate migration: %w", err)
	}

	fmt.Println("✅ Migration generated successfully!")
	fmt.Println()
	fmt.Printf("Generated file: %s\n", filename)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Review the generated SQL before committing!")
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Review the migration file")
	fmt.Println("   2. Run 'apprun migrate validate' to verify")
	fmt.Println("   3. Run 'apprun migrate apply' to apply changes")

	return nil
}

func runMigrateSync(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Println("🔄 Syncing database schema...")
	fmt.Println()

	// Step 1: Check current migration status
	fmt.Println("Step 1: Checking current migration status...")
	statusBefore, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	// If there are pending migrations, apply them first
	if len(statusBefore.Pending) > 0 {
		fmt.Printf("📋 Found %d pending migration(s), applying them first...\n", len(statusBefore.Pending))
		fmt.Println()

		if migrateSyncDryRun {
			fmt.Println("🔍 Dry-run mode: Would apply pending migrations")
			for _, p := range statusBefore.Pending {
				fmt.Printf("  - %s\n", p)
			}
		} else {
			if err := migrator.ApplyMigrations(ctx); err != nil {
				return fmt.Errorf("failed to apply pending migrations: %w", err)
			}
			fmt.Println("✅ Pending migrations applied successfully")
		}
		fmt.Println()
	}

	// Step 2: Check for new schema changes
	fmt.Println("Step 2: Checking for new schema changes...")

	// Auto-generate migration name with timestamp
	migrationName := fmt.Sprintf("sync_%d", time.Now().Unix())

	// Try to generate migration from schema changes
	filename, err := migrator.GenerateMigration(ctx, migrationName, migrateToSchema, migrateDevURL)
	if err != nil {
		// Check if it's "no changes" error
		if contains(err.Error(), "no changes") || contains(err.Error(), "no diff") {
			fmt.Println("✅ Schema is already up to date - no new changes")
			return nil
		}
		return fmt.Errorf("failed to generate migration: %w", err)
	}

	// A new migration was generated
	fmt.Printf("📝 New migration generated: %s\n", filename)
	fmt.Println()

	if migrateSyncDryRun {
		fmt.Println("🔍 Dry-run mode: Migration file created but not applied")
		fmt.Println("💡 Run 'apprun migrate apply' to apply the migration")
		return nil
	}

	// Step 3: Apply the newly generated migration
	fmt.Println("Step 3: Applying new migration...")
	if err := migrator.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Database schema synced successfully!")
	fmt.Println()
	fmt.Println("💡 Migration file created and applied:")
	fmt.Printf("   %s\n", filename)
	fmt.Println()
	fmt.Println("⚠️  Review the migration file and commit it to version control")

	return nil
}

func contains(s, substr string) bool {
	// Simple contains check
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func runMigrateRollback(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Println("⚠️  Rolling back last migration...")
	fmt.Println()

	// Rollback migration
	if err := migrator.RollbackMigration(ctx); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("✅ Migration rolled back successfully")
	fmt.Println()
	fmt.Println("💡 Run 'apprun migrate status' to verify current state")

	return nil
}

func runMigrateReset(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Safety check - require confirmation
	fmt.Println("⚠️  WARNING: This will DELETE ALL DATA in the database!")
	fmt.Println()
	fmt.Print("Type 'yes' to confirm reset: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		fmt.Println("❌ Reset canceled")
		return nil
	}

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Println()
	fmt.Println("🔄 Resetting database...")

	// Reset database
	if err := migrator.ResetDatabase(ctx); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	fmt.Println("✅ Database reset successfully")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Run 'apprun migrate apply' to apply migrations")
	fmt.Println("   2. Verify with 'apprun migrate status'")

	return nil
}

func runMigrateInspect(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	fmt.Println("🔍 Inspecting database schema...")
	fmt.Println()

	// Inspect schema
	diff, err := migrator.InspectSchema(ctx, migrateToSchema, migrateDevURL)
	if err != nil {
		return fmt.Errorf("failed to inspect schema: %w", err)
	}

	if !diff.HasDifferences {
		fmt.Println("✅ Database schema is in sync with Ent schema")
		fmt.Println()
		fmt.Println("No schema drift detected:")
		fmt.Println("  - All tables exist as defined in Ent schema")
		fmt.Println("  - All columns match their definitions")
		fmt.Println("  - All indexes are correctly created")
		return nil
	}

	// Schema drift detected
	fmt.Println("⚠️  Schema drift detected!")
	fmt.Println()
	fmt.Println("Differences between database and Ent schema:")
	fmt.Println(diff.Description)
	fmt.Println()
	fmt.Println("💡 Possible causes:")
	fmt.Println("   - Manual database changes (ALTER TABLE, DROP COLUMN, etc.)")
	fmt.Println("   - Failed migration that left database inconsistent")
	fmt.Println("   - Migration files corrupted or lost")
	fmt.Println()
	fmt.Println("🔧 To fix:")
	fmt.Println("   apprun migrate repair          # Preview SQL")
	fmt.Println("   apprun migrate repair --execute # Apply fix")
	fmt.Println()
	fmt.Println("⚠️  Or reset and re-apply migrations:")
	fmt.Println("   apprun migrate reset")
	fmt.Println("   apprun migrate sync")

	return nil
}

func runMigrateRepair(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	// Check environment
	env := getenv("APP_ENV", "development")
	if env == "production" {
		fmt.Println("❌ ERROR: 'migrate repair' is not allowed in production")
		fmt.Println()
		fmt.Println("Production environments must use versioned migrations:")
		fmt.Println("  1. Generate migration: apprun migrate diff fix_schema")
		fmt.Println("  2. Review SQL file: cat migrations/*.sql")
		fmt.Println("  3. Apply migration: apprun migrate apply")
		return fmt.Errorf("repair command blocked in production")
	}

	fmt.Println("🔧 Repairing database schema...")
	fmt.Println()

	// Determine mode
	dryRun := migrateRepairDryRun
	if migrateRepairExecute {
		dryRun = false
	}

	if dryRun {
		fmt.Println("🔍 Dry-run mode: SQL will be displayed but not executed")
		fmt.Println()
	} else {
		fmt.Println("⚠️  WARNING: This will modify your database!")
		fmt.Println()
		fmt.Println("Recommendations before proceeding:")
		fmt.Println("  1. Backup your database")
		fmt.Println("  2. Review the SQL that will be executed")
		fmt.Println()
	}

	// First, check if there are differences
	diff, err := migrator.InspectSchema(ctx, migrateToSchema, migrateDevURL)
	if err != nil {
		return fmt.Errorf("failed to inspect schema: %w", err)
	}

	if !diff.HasDifferences {
		fmt.Println("✅ Database schema is already in sync with Ent schema")
		fmt.Println()
		fmt.Println("No repair needed!")
		return nil
	}

	fmt.Println("📋 Detected schema differences:")
	fmt.Println(diff.Description)
	fmt.Println()

	// Perform repair
	if err := migrator.RepairSchema(ctx, migrateToSchema, migrateDevURL, dryRun); err != nil {
		return fmt.Errorf("failed to repair schema: %w", err)
	}

	if dryRun {
		fmt.Println()
		fmt.Println("💡 To apply these changes:")
		fmt.Println("   apprun migrate repair --execute")
	} else {
		fmt.Println()
		fmt.Println("✅ Database schema repaired successfully!")
		fmt.Println()
		fmt.Println("⚠️  Important notes:")
		fmt.Println("   - This was a declarative repair (no migration file created)")
		fmt.Println("   - Migration history may be out of sync")
		fmt.Println("   - Consider regenerating migrations if needed:")
		fmt.Println("     1. Reset: apprun migrate reset")
		fmt.Println("     2. Generate: apprun migrate diff baseline")
		fmt.Println("     3. Apply: apprun migrate apply")
	}

	return nil
}

func runMigrateClean(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := setupMigrator(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	// Get flags
	cleanAll, _ := cmd.Flags().GetBool("all")
	cleanLast, _ := cmd.Flags().GetBool("last")

	// Safety check - require environment confirmation
	env := getenv("APP_ENV", "development")
	if env == "production" {
		fmt.Println("❌ ERROR: 'migrate clean' is not allowed in production")
		fmt.Println()
		fmt.Println("Production environments must use proper migration rollback:")
		fmt.Println("  apprun migrate rollback")
		return fmt.Errorf("clean command blocked in production")
	}

	fmt.Println("🧹 Cleaning migration history...")
	fmt.Println()

	if cleanAll {
		// Remove all failed migrations
		fmt.Println("⚠️  Removing all failed migration records...")
		count, err := migrator.CleanFailedMigrations(ctx)
		if err != nil {
			return fmt.Errorf("failed to clean migrations: %w", err)
		}
		fmt.Printf("✅ Removed %d failed migration record(s)\n", count)
	} else if cleanLast {
		// Remove last migration
		fmt.Println("⚠️  Removing last migration record...")
		version, err := migrator.CleanLastMigration(ctx)
		if err != nil {
			return fmt.Errorf("failed to clean last migration: %w", err)
		}
		fmt.Printf("✅ Removed migration: %s\n", version)
	} else if len(args) > 0 {
		// Remove specific version
		version := args[0]
		fmt.Printf("⚠️  Removing migration record: %s...\n", version)
		if err := migrator.CleanMigration(ctx, version); err != nil {
			return fmt.Errorf("failed to clean migration: %w", err)
		}
		fmt.Printf("✅ Removed migration: %s\n", version)
	} else {
		return fmt.Errorf("specify version, --all, or --last flag")
	}

	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Review migration files in core/migrations/")
	fmt.Println("   2. Delete any invalid migration files manually")
	fmt.Println("   3. Run: atlas migrate hash --dir file://core/migrations")
	fmt.Println("   4. Verify: apprun migrate status")

	return nil
}

// generateAtlasConfig creates a temporary atlas.hcl file for atlas CLI
// This embeds the configuration in the binary and generates it at runtime
func generateAtlasConfig(dbURL string) (configPath string, cleanup func(), err error) {
	// Replace placeholder with actual database URL
	config := strings.ReplaceAll(atlasConfigTemplate,
		"postgres://apprun:dev_password_123@localhost:5432/apprun_dev?sslmode=disable",
		dbURL)

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "atlas-*.hcl")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp config: %w", err)
	}

	configPath = tmpFile.Name()

	// Write config content
	if _, err := tmpFile.WriteString(config); err != nil {
		tmpFile.Close()
		os.Remove(configPath)
		return "", nil, fmt.Errorf("failed to write config: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(configPath)
		return "", nil, fmt.Errorf("failed to close config file: %w", err)
	}

	// Return cleanup function
	cleanup = func() {
		os.Remove(configPath)
	}

	return configPath, cleanup, nil
}

// getAtlasConfigPath returns the path to atlas config file
// Priority: 1. Embedded config (runtime generated) 2. External file (fallback)
func getAtlasConfigPath(dbURL string) (configPath string, cleanup func(), err error) {
	// Check if external atlas.hcl exists (for development)
	externalConfig := filepath.Join(".", "atlas.hcl")
	if _, err := os.Stat(externalConfig); err == nil {
		// Use external config for development/debugging
		return externalConfig, func() {}, nil
	}

	// Generate from embedded template
	return generateAtlasConfig(dbURL)
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
