package cmd

import (
	"apprun/internal/bootstrap"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command (renamed from start)
var serveCmd = &cobra.Command{
	Use:     "serve",
	GroupID: "server",
	Short:   "Start AppRun HTTP server",
	Long: `Start the AppRun BaaS platform HTTP server with REST API and GraphQL endpoints.

The server will initialize all required components including:
  - Database connections
  - Authentication services
  - Configuration management
  - Cache clients (Redis)
  - RBAC enforcement
  - HTTP routes and middleware

Examples:
  apprun serve                                    # Start with default config
  apprun serve --config ./config/production.yaml  # Use custom config
`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Start server with full bootstrap
	return bootstrap.StartServer()
}
