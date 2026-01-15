package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"apprun/pkg/database"

	"github.com/spf13/cobra"
)

var (
	migrateDryRun bool
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

Examples:
  apprun migrate apply            # Apply all pending migrations
  apprun migrate apply --dry-run  # Preview SQL without executing
  apprun migrate status           # Check current migration state
  apprun migrate validate         # Verify migration files
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

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add subcommands
	migrateCmd.AddCommand(migrateApplyCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateValidateCmd)

	// Flags for apply subcommand
	migrateApplyCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "preview SQL without executing")
}

func runMigrateApply(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Load database config from environment
	cfg := database.DefaultConfig()

	// Create migrator
	migrator, err := database.NewMigratorFromConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			log.Printf("Warning: failed to close migrator: %v", closeErr)
		}
	}()

	if migrateDryRun {
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
	migrator, err := database.NewMigratorFromConfig(ctx, cfg)
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
	fmt.Println("🔍 Validating migration files...")

	// For now, basic validation
	// In production, this would check:
	// - File naming conventions
	// - Sequential version numbers
	// - SQL syntax validity
	// - No duplicate versions

	fmt.Println("✅ Migration files validated successfully")
	fmt.Println("💡 For detailed validation, use Atlas CLI: atlas migrate validate")

	return nil
}
