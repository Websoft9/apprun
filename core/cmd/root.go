// Package cmd implements the command-line interface for AppRun platform.
// It uses the Cobra framework for command structure and flag management.
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"apprun/pkg/env"
	"apprun/pkg/version"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	configFile string
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apprun",
	Short: "AppRun BaaS Platform",
	Long: `AppRun - Enterprise-grade Backend-as-a-Service Platform

A comprehensive BaaS solution with built-in authentication, database management,
API scaffolding, and runtime configuration support.

Find more information at: https://github.com/websoft9/apprun`,
	Version: version.BuildVersion(),
	// PersistentPreRun loads configuration before any command runs
	// This ensures all commands have access to config file and environment variables
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip config loading for commands that don't need it
		// version command should work without any configuration
		// configure command manages user config, not application config
		// client commands (deploy, logs, backup) are placeholders in Story 1.6, full implementation in Story 1.6.2
		skipConfigCommands := []string{"version", "help", "completion", "configure", "show", "deploy", "logs", "backup"}
		for _, skipCmd := range skipConfigCommands {
			if cmd.Name() == skipCmd {
				return
			}
		}

		// Determine config directory
		// Priority: 1. --config flag, 2. CONFIG_DIR env var, 3. user config, 4. default
		configDir := "./config"

		// Check if user has configured a config path via 'apprun configure'
		userConfig, err := GetUserConfig()
		if err == nil && userConfig.ConfigPath != "" {
			// Use configured path as default
			configDir = filepath.Dir(userConfig.ConfigPath)
		}

		// Override with flag if provided
		if configFile != "" {
			configDir = configFile
		} else if envConfigDir := os.Getenv("CONFIG_DIR"); envConfigDir != "" {
			configDir = envConfigDir
		}

		// Check if config file exists
		configPath := filepath.Join(configDir, "default.yaml")
		configFileExists := true
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configFileExists = false
		}

		// If config file doesn't exist, check if essential environment variables are set
		if !configFileExists {
			// Different commands require different environment variables
			var essentialEnvVars []string

			switch cmd.Name() {
			case "migrate":
				// Migration commands only need database configuration
				essentialEnvVars = []string{
					"DATABASE_HOST",
					"DATABASE_USER",
					"DATABASE_PASSWORD",
					"DATABASE_DB_NAME",
				}
			case "serve":
				// Server needs database + JWT secret for authentication
				essentialEnvVars = []string{
					"DATABASE_HOST",
					"DATABASE_USER",
					"DATABASE_PASSWORD",
					"DATABASE_DB_NAME",
					"AUTH_JWT_SECRET",
				}
			default:
				// Other commands might need database access
				essentialEnvVars = []string{
					"DATABASE_HOST",
					"DATABASE_USER",
					"DATABASE_PASSWORD",
					"DATABASE_DB_NAME",
				}
			}

			missingVars := []string{}
			for _, envVar := range essentialEnvVars {
				if os.Getenv(envVar) == "" {
					missingVars = append(missingVars, envVar)
				}
			}

			// If essential environment variables are missing, show error
			if len(missingVars) > 0 {
				cwd, _ := os.Getwd()
				cmdPath := cmd.CommandPath()

				fmt.Fprintf(os.Stderr, "\n❌ Configuration Error: Config file not found and essential environment variables missing\n\n")
				fmt.Fprintf(os.Stderr, "Expected config file: %s\n", configPath)
				fmt.Fprintf(os.Stderr, "Current directory: %s\n", cwd)
				fmt.Fprintf(os.Stderr, "Command: %s\n", cmd.Name())
				fmt.Fprintf(os.Stderr, "Missing environment variables: %s\n\n", strings.Join(missingVars, ", "))
				fmt.Fprintf(os.Stderr, "Solutions:\n")
				fmt.Fprintf(os.Stderr, "  1. Configure CLI settings (recommended):\n")
				fmt.Fprintf(os.Stderr, "     apprun configure\n\n")
				fmt.Fprintf(os.Stderr, "  2. Run from project root:\n")
				fmt.Fprintf(os.Stderr, "     cd /path/to/apprun/core\n")
				fmt.Fprintf(os.Stderr, "     ./bin/%s\n\n", cmdPath)
				fmt.Fprintf(os.Stderr, "  3. Specify config directory:\n")
				fmt.Fprintf(os.Stderr, "     ./%s --config /path/to/config\n\n", cmdPath)
				fmt.Fprintf(os.Stderr, "  4. Set essential environment variables:\n")

				// Show command-specific environment variable examples
				switch cmd.Name() {
				case "migrate":
					fmt.Fprintf(os.Stderr, "     export DATABASE_HOST=localhost\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_USER=postgres\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_PASSWORD=your_password\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_DB_NAME=apprun\n")
				case "serve":
					fmt.Fprintf(os.Stderr, "     export DATABASE_HOST=localhost\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_USER=postgres\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_PASSWORD=your_password\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_DB_NAME=apprun\n")
					fmt.Fprintf(os.Stderr, "     export AUTH_JWT_SECRET=your-32-char-jwt-secret-key\n")
				default:
					fmt.Fprintf(os.Stderr, "     export DATABASE_HOST=localhost\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_USER=postgres\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_PASSWORD=your_password\n")
					fmt.Fprintf(os.Stderr, "     export DATABASE_DB_NAME=apprun\n")
				}
				fmt.Fprintf(os.Stderr, "     ./%s\n\n", cmdPath)
				os.Exit(1)
			}

			// Config file doesn't exist but essential env vars are set - continue with env-only mode
			if verbose {
				log.Printf("✅ Using environment variables only (no config file)")
			}
			return
		}

		// Config file exists - load it to environment variables
		// Priority: runtime env > config file > code defaults
		if err := env.LoadConfigToEnv(configDir); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Configuration Error: Failed to load config file\n\n")
			fmt.Fprintf(os.Stderr, "File: %s\n", configPath)
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "Please check:\n")
			fmt.Fprintf(os.Stderr, "  - YAML syntax is valid\n")
			fmt.Fprintf(os.Stderr, "  - File is readable\n")
			fmt.Fprintf(os.Stderr, "  - Required fields are present\n\n")
			os.Exit(1)
		}

		// Set the absolute config directory path for use by subcommands
		// This allows migration commands to find migrations directory relative to config
		absConfigDir, err := filepath.Abs(configDir)
		if err == nil {
			os.Setenv("APPRUN_CONFIG_DIR", absConfigDir)
		}

		if verbose {
			log.Printf("✅ Config loaded from: %s", configPath)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path (default: ./config/default.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Register command groups
	rootCmd.AddGroup(&cobra.Group{
		ID:    "server",
		Title: "Server Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "client",
		Title: "Client Commands:",
	})

	// Customize help template to group commands
	rootCmd.SetUsageTemplate(customUsageTemplate())

	// Register subcommands
	// serve, migrate, version, configure commands will be registered via their init() functions
}

// customUsageTemplate returns a custom usage template with command grouping
func customUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

// GetConfigFile returns the config file path from flags
func GetConfigFile() string {
	return configFile
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// Exit is a wrapper around os.Exit for testability
func Exit(code int) {
	os.Exit(code)
}

// Printf prints to stdout (wrapper for testability)
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
