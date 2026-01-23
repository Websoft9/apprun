package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"apprun/internal/bootstrap"
	"apprun/modules/config"
	"apprun/pkg/cache"
	"apprun/pkg/database"
	"apprun/pkg/server"

	"github.com/spf13/cobra"
)

var (
	// Generate command flags
	generateOutputDir  string
	generateConfigOnly bool
	generateEnvOnly    bool

	// OpenAPI generation flags
	openapiOutputDir string
	openapiMainFile  string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate configuration and code artifacts",
	Long: `Generate various configuration files and code artifacts.

Available subcommands:
  config  - Generate config.example and .env.example files
  model   - Generate Ent ORM code from schema definitions
  openapi - Generate OpenAPI/Swagger documentation

Examples:
  apprun generate config                  # Generate configuration examples
  apprun generate model                   # Generate Ent models
  apprun generate openapi                 # Generate API documentation
  apprun generate config --output /path   # Custom output directory`,
}

// generateConfigCmd generates configuration example files
var generateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config.example and .env.example",
	Long: `Generate configuration example files from registered modules.

This command introspects the module registry and generates:
- config.example: YAML configuration template with all available settings
- .env.example: Environment variables template for operational settings

The generated files include:
- Default values from struct tags
- Validation rules
- Database persistence flags (db:"true/false")
- Environment-only operational switches (envonly:"true")

Output location:
- config.example: {output-dir}/config/config.example (default: ./config/)
- .env.example: {output-dir}/.env.example (default: ./)`,
	RunE: runGenerateConfig,
}

func init() {
	// Add generate command to root
	rootCmd.AddCommand(generateCmd)

	// Add config subcommand to generate
	generateCmd.AddCommand(generateConfigCmd)

	// Add model subcommand to generate
	generateCmd.AddCommand(generateModelCmd)

	// Add openapi subcommand to generate
	generateCmd.AddCommand(generateOpenAPICmd)

	// Add flags to generate config command
	generateConfigCmd.Flags().StringVarP(&generateOutputDir, "output", "o", ".", "Output directory for generated files")
	generateConfigCmd.Flags().BoolVar(&generateConfigOnly, "config-only", false, "Generate config.example only")
	generateConfigCmd.Flags().BoolVar(&generateEnvOnly, "env-only", false, "Generate .env.example only")

	// Add flags to generate openapi command
	generateOpenAPICmd.Flags().StringVarP(&openapiOutputDir, "output", "o", "apidocs", "Output directory for OpenAPI documentation")
	generateOpenAPICmd.Flags().StringVarP(&openapiMainFile, "main", "m", "internal/bootstrap/server.go", "Main file for API annotations")
}

// runGenerateConfig executes the config generation
func runGenerateConfig(cmd *cobra.Command, args []string) error {
	// Create and populate registry
	registry := createGeneratorRegistry()

	// Create generator with populated registry
	generator := config.NewGenerator(registry)

	// Determine what to generate
	generateBoth := !generateConfigOnly && !generateEnvOnly

	// Generate config.example
	if generateBoth || generateConfigOnly {
		if err := generateConfigFile(generator); err != nil {
			return fmt.Errorf("failed to generate config.example: %w", err)
		}
	}

	// Generate .env.example
	if generateBoth || generateEnvOnly {
		if err := generateEnvFile(generator); err != nil {
			return fmt.Errorf("failed to generate .env.example: %w", err)
		}
	}

	fmt.Printf("   Registered modules: %d\n", registry.Count())
	return nil
}

// createGeneratorRegistry creates and populates a registry with all modules
func createGeneratorRegistry() *config.ConfigRegistry {
	registry := config.NewRegistry()

	// Register business modules using auto-registration
	modules := config.DefaultModules()
	for _, mod := range modules {
		if err := registry.Register(mod.Namespace, mod.ConfigStruct); err != nil {
			fmt.Printf("Warning: Failed to register module '%s': %v\n", mod.Namespace, err)
		}
	}

	// Register infrastructure modules (not in Config struct but needed for config.example)
	// These modules are initialized at startup and don't support runtime configuration
	if err := registry.Register("database", &database.Config{}); err != nil {
		fmt.Printf("Warning: Failed to register database: %v\n", err)
	}
	if err := registry.Register("cache", &cache.Config{}); err != nil {
		fmt.Printf("Warning: Failed to register cache: %v\n", err)
	}
	if err := registry.Register("server", &server.Config{}); err != nil {
		fmt.Printf("Warning: Failed to register server: %v\n", err)
	}

	// Register bootstrap module (operational settings, envonly)
	if err := registry.Register("bootstrap", &bootstrap.Config{}); err != nil {
		fmt.Printf("Warning: Failed to register bootstrap: %v\n", err)
	}

	return registry
}

// generateConfigFile generates config.example
func generateConfigFile(generator config.Generator) error {
	// Build output path
	configPath := filepath.Join(generateOutputDir, "config", "config.example")

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	// #nosec G301 -- config directory needs to be readable by service users
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate content
	content, err := generator.GenerateConfigExample()
	if err != nil {
		return err
	}

	// Write to file
	// #nosec G306 -- config.example is an example file, 0644 is appropriate for documentation
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config.example: %w", err)
	}

	fmt.Printf("✅ Successfully generated %s\n", configPath)
	return nil
}

// generateEnvFile generates .env.example
func generateEnvFile(generator config.Generator) error {
	// Build output path
	envPath := filepath.Join(generateOutputDir, ".env.example")

	// Generate content
	content := generator.GenerateEnvExample()

	// Write to file
	// #nosec G306 -- .env.example is an example file, 0644 is appropriate
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .env.example: %w", err)
	}

	fmt.Printf("✅ Successfully generated %s\n", envPath)
	return nil
}

// ============================================
// Model Generation (Ent ORM)
// ============================================

// generateModelCmd generates Ent ORM code
var generateModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate Ent ORM code from schema definitions",
	Long: `Generate Ent ORM code from schema definitions.

This command runs the Ent code generator to create:
- Entity structs and interfaces
- Database query builders
- Migration files (when using Atlas)
- Type-safe CRUD operations

The generator reads schema definitions from ent/schema/ directory
and generates code in the ent/ package.

Features enabled:
- VersionedMigration (Atlas integration)
- Privacy (access control)
- Upsert (insert or update)

Examples:
  apprun generate model           # Generate Ent code
  
See also: https://entgo.io/docs/code-gen`,
	RunE: runGenerateModel,
}

// runGenerateModel executes the Ent code generation
func runGenerateModel(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Generating Ent ORM code...")

	// Change to ent directory
	entDir := "ent"
	if err := os.Chdir(entDir); err != nil {
		return fmt.Errorf("failed to change to ent directory: %w", err)
	}
	defer os.Chdir("..") // Return to original directory

	// Run go generate
	genCmd := exec.Command("go", "generate", ".")
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr

	if err := genCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate Ent code: %w", err)
	}

	fmt.Println("✅ Ent code generated successfully")
	fmt.Println("💡 Generated files in ent/ directory")
	return nil
}

// ============================================
// OpenAPI/Swagger Documentation
// ============================================

// generateOpenAPICmd generates OpenAPI/Swagger documentation
var generateOpenAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate OpenAPI/Swagger API documentation",
	Long: `Generate OpenAPI/Swagger API documentation from code annotations.

This command runs swag init to generate OpenAPI 2.0 (Swagger) documentation
from Go annotations in your handler files.

The generator produces:
- apidocs/swagger.json - OpenAPI specification in JSON format
- apidocs/swagger.yaml - OpenAPI specification in YAML format
- apidocs/docs.go - Go code for serving the documentation

After generation, documentation is accessible at:
  http://localhost:8080/api/docs/

Annotation format:
  // @Summary      Get user by ID
  // @Description  Retrieve user details
  // @Tags         users
  // @Accept       json
  // @Produce      json
  // @Param        id   path      int  true  "User ID"
  // @Success      200  {object}  User
  // @Router       /users/{id} [get]

Examples:
  apprun generate openapi                                # Generate docs
  apprun generate openapi --output apidocs               # Custom output dir
  apprun generate openapi --main internal/bootstrap/app.go  # Custom main file
  
See also: https://github.com/swaggo/swag`,
	RunE: runGenerateOpenAPI,
}

// runGenerateOpenAPI executes the OpenAPI documentation generation
func runGenerateOpenAPI(cmd *cobra.Command, args []string) error {
	fmt.Println("📚 Generating OpenAPI/Swagger documentation...")

	// Check if swag is installed
	if _, err := exec.LookPath("swag"); err != nil {
		return fmt.Errorf("swag command not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest")
	}

	// Build swag command
	swagArgs := []string{
		"init",
		"-g", openapiMainFile,
		"-o", openapiOutputDir,
	}

	// #nosec G204 -- swag binary is hardcoded, only args are dynamic (safe file paths)
	swagCmd := exec.Command("swag", swagArgs...)
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr

	if err := swagCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate OpenAPI docs: %w", err)
	}

	fmt.Printf("✅ OpenAPI documentation generated in %s/\n", openapiOutputDir)
	fmt.Println("📄 Files created:")
	fmt.Printf("   - %s/swagger.json\n", openapiOutputDir)
	fmt.Printf("   - %s/swagger.yaml\n", openapiOutputDir)
	fmt.Printf("   - %s/docs.go\n", openapiOutputDir)
	fmt.Println("💡 Access at: http://localhost:8080/api/apidocs/")
	return nil
}
