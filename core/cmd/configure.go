package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Validation functions

// isValidURL checks if the string is a valid HTTP/HTTPS URL
func isValidURL(s string) bool {
	if s == "" {
		return true // Empty is allowed (optional field)
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	// Must have http or https scheme and a valid host
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// isValidPath checks if the string is a valid local filesystem path
func isValidPath(s string) bool {
	if s == "" {
		return true // Empty is allowed (optional field)
	}
	// Check if path is absolute or relative
	// Convert to absolute path for validation
	absPath := s
	if !filepath.IsAbs(s) {
		var err error
		absPath, err = filepath.Abs(s)
		if err != nil {
			return false
		}
	}
	// Path should not contain invalid characters
	// On Unix: basically any character is valid except null
	// On Windows: we rely on filepath.Abs to reject invalid paths
	return absPath != ""
}

// UserConfig represents the CLI user configuration stored in ~/.apprun/config.yaml
type UserConfig struct {
	// Client CLI configuration (for remote operations)
	Endpoint string `yaml:"endpoint,omitempty"` // API endpoint (e.g., https://api.apprun.com)
	APIKey   string `yaml:"api_key,omitempty"`  // Authentication key

	// Server CLI configuration (for local operations)
	ConfigPath string `yaml:"config_path,omitempty"` // Application config file path
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:     "configure",
	GroupID: "server",
	Short:   "Configure AppRun CLI settings",
	Long: `Interactive configuration wizard for AppRun CLI.

This command sets up user-level configuration stored in ~/.apprun/config.yaml.
Configuration includes:
  - endpoint: API endpoint for remote operations (e.g., deploy, logs, backup)
  - api_key: Authentication key for client commands
  - config_path: Application config file path for server commands (default: ./config/default.yaml)

The configuration file is used by CLI commands to avoid repeated parameter entry.

Examples:
  apprun configure           # Interactive configuration wizard
  apprun configure show      # Display current configuration
`,
	RunE: runConfigure,
}

// configureShowCmd shows current configuration
var configureShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current CLI configuration from ~/.apprun/config.yaml.

Shows:
  - API endpoint (if configured)
  - API key (masked for security)
  - Application config path
  - Configuration file location

Examples:
  apprun configure show
`,
	RunE: runConfigureShow,
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.AddCommand(configureShowCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	fmt.Println("🔧 AppRun CLI Configuration Wizard")
	fmt.Println("-----------------------------------")
	fmt.Println()

	// Load existing config if it exists
	existingConfig, configPath, err := loadUserConfig()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	// Initialize with empty config if it doesn't exist
	if existingConfig == nil {
		existingConfig = &UserConfig{}
	}

	// If config exists, show current values
	if existingConfig.Endpoint != "" || existingConfig.APIKey != "" || existingConfig.ConfigPath != "" {
		fmt.Println("📋 Current configuration found:")
		displayConfig(existingConfig, false)
		fmt.Println()
	}

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	// Endpoint configuration
	fmt.Println("🌐 Client CLI Configuration (for remote operations)")
	fmt.Println("   Used by: deploy, logs, backup commands")
	fmt.Println()

	endpoint := promptWithValidation(reader, "API Endpoint", existingConfig.Endpoint, "https://api.apprun.com", isValidURL, "Invalid URL. Must be a valid http:// or https:// URL")
	apiKey := promptWithDefault(reader, "API Key", existingConfig.APIKey, "")

	fmt.Println()
	fmt.Println("💻 Server CLI Configuration (for local operations)")
	fmt.Println("   Used by: serve, migrate commands")
	fmt.Println()

	appConfigPath := promptWithValidation(reader, "Application Config Path", existingConfig.ConfigPath, "./config/default.yaml", isValidPath, "Invalid path. Must be a valid filesystem path")

	// Create new config
	newConfig := &UserConfig{
		Endpoint:   endpoint,
		APIKey:     apiKey,
		ConfigPath: appConfigPath,
	}

	// Save configuration
	if err := saveUserConfig(newConfig, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Configuration saved successfully!")
	fmt.Println()
	fmt.Println("📍 Configuration file:", configPath)
	fmt.Println()
	fmt.Println("You can now use CLI commands without specifying these parameters.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  apprun serve              # Uses configured config_path")
	fmt.Println("  apprun migrate status     # Uses configured config_path")
	fmt.Println("  apprun deploy             # Uses configured endpoint and api_key (future)")
	fmt.Println()
	fmt.Println("To view your configuration:")
	fmt.Println("  apprun configure show")

	return nil
}

func runConfigureShow(cmd *cobra.Command, args []string) error {
	config, configPath, err := loadUserConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("❌ No configuration found")
			fmt.Println()
			fmt.Println("Run 'apprun configure' to set up your CLI configuration.")
			return nil
		}
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println("📋 Current AppRun CLI Configuration")
	fmt.Println("====================================")
	fmt.Println()
	displayConfig(config, true)
	fmt.Println()
	fmt.Println("📍 Configuration file:", configPath)
	fmt.Println()
	fmt.Println("To update configuration:")
	fmt.Println("  apprun configure")

	return nil
}

// promptWithDefault prompts the user for input with a default value
func promptWithDefault(reader *bufio.Reader, prompt, currentValue, defaultValue string) string {
	var displayValue string
	if currentValue != "" {
		displayValue = currentValue
	} else {
		displayValue = defaultValue
	}

	if displayValue != "" {
		fmt.Printf("%s [%s]: ", prompt, displayValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return displayValue
	}

	return input
}

// promptWithValidation prompts the user for input with validation
func promptWithValidation(reader *bufio.Reader, prompt, currentValue, defaultValue string, validator func(string) bool, errorMsg string) string {
	for {
		var displayValue string
		if currentValue != "" {
			displayValue = currentValue
		} else {
			displayValue = defaultValue
		}

		if displayValue != "" {
			fmt.Printf("%s [%s]: ", prompt, displayValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Use default if no input
		if input == "" {
			input = displayValue
		}

		// Validate input
		if validator(input) {
			return input
		}

		// Show error and retry
		fmt.Printf("❌ %s\n", errorMsg)
		fmt.Println("   Please try again.")
		fmt.Println()
	}
}

// displayConfig displays the configuration with optional masking of sensitive data
func displayConfig(config *UserConfig, maskSecrets bool) {
	fmt.Println("Client Configuration:")
	fmt.Printf("  Endpoint:    %s\n", getValueOrEmpty(config.Endpoint))

	apiKey := config.APIKey
	if maskSecrets && apiKey != "" {
		if len(apiKey) > 8 {
			apiKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		} else {
			apiKey = "****"
		}
	}
	fmt.Printf("  API Key:     %s\n", getValueOrEmpty(apiKey))

	fmt.Println()
	fmt.Println("Server Configuration:")
	fmt.Printf("  Config Path: %s\n", getValueOrEmpty(config.ConfigPath))
}

func getValueOrEmpty(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}

// getUserConfigPath is a variable to allow testing with dependency injection
var getUserConfigPath = defaultGetUserConfigPath

// defaultGetUserConfigPath returns the path to the user configuration file
func defaultGetUserConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".apprun")
	configPath := filepath.Join(configDir, "config.yaml")

	return configPath, nil
}

// loadUserConfig loads the user configuration from ~/.apprun/config.yaml
func loadUserConfig() (*UserConfig, string, error) {
	configPath, err := getUserConfigPath()
	if err != nil {
		return nil, "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, configPath, err
	}

	var config UserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, configPath, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, configPath, nil
}

// saveUserConfig saves the user configuration to ~/.apprun/config.yaml
func saveUserConfig(config *UserConfig, configPath string) error {
	configPath, err := getUserConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with appropriate permissions (user read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetUserConfig loads the user configuration or returns defaults
func GetUserConfig() (*UserConfig, error) {
	config, _, err := loadUserConfig()
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &UserConfig{
				ConfigPath: "./config/default.yaml",
			}, nil
		}
		return nil, err
	}

	return config, nil
}
