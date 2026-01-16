package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// backupCmd represents the backup command (client command - requires auth)
var backupCmd = &cobra.Command{
	Use:     "backup",
	GroupID: "client",
	Short:   "Manage remote backups (future feature)",
	Long: `Trigger and manage backups for remote AppRun application.

This is a client command that connects to a remote AppRun API.
Authentication is required via endpoint and API key configured in ~/.apprun/config.yaml.

Examples:
  apprun backup create             # Create a new backup
  apprun backup list               # List available backups
  apprun backup restore <backup-id> # Restore from backup

Status: Not implemented yet (future expansion)
`,
	RunE: runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
	// Future subcommands: create, list, restore, delete
}

func runBackup(cmd *cobra.Command, args []string) error {
	fmt.Println("⚠️  Backup command not yet implemented")
	fmt.Println()
	fmt.Println("This is a placeholder for future client command functionality.")
	fmt.Println()
	fmt.Println("Planned features:")
	fmt.Println("  - Create backups of remote AppRun instance")
	fmt.Println("  - List available backups")
	fmt.Println("  - Restore from backup")
	fmt.Println("  - Authentication via API key from ~/.apprun/config.yaml")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  apprun configure                              # Set default endpoint")
	fmt.Println("  apprun backup --endpoint https://api.example.com  # Override endpoint")
	fmt.Println()
	fmt.Println("Note: Full implementation provided in Story 1.6.2 (CLI-API Adapter)")

	return nil
}
