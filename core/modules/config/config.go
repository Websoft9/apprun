// Package config defines the configuration center and its global configuration structure.
package config

import (
	authmod "apprun/modules/auth"
	"apprun/pkg/cache"
	"apprun/pkg/database"
	"apprun/pkg/i18n"
	"apprun/pkg/logger"
)

// Config is the root configuration structure for the entire application.
// It embeds all module configurations to serve as the single source of truth.
//
// The `register:"auto"` tag indicates which modules should be automatically
// registered with the config center at startup. This enables:
// - Single definition: Module configs are defined once in this struct
// - Multiple uses: Used for both YAML loading and runtime registration
//
// YAML keys are implicitly mapped from lowercase field names (e.g., Name -> name).
// Only nested structs require explicit mapstructure tags to define root keys.
type Config struct {
	App struct {
		Name     string `mapstructure:"name" json:"name" validate:"required,min=1" default:"apprun" db:"true"`
		Version  string `mapstructure:"version" json:"version" validate:"required" default:"1.0.0" db:"false"`
		Timezone string `mapstructure:"timezone" json:"timezone" validate:"required,timezone" default:"Asia/Shanghai" db:"true"`
	} `mapstructure:"app" json:"app" validate:"required"`

	// Module configurations with auto-registration support
	// The register:"auto" tag marks modules that should be automatically registered
	// The description:"..." tag provides human-readable descriptions for logging

	Database database.Config `mapstructure:"database" json:"database" validate:"required" register:"skip" description:"Database connection (startup-only, not runtime configurable)"`
	Cache    cache.Config    `mapstructure:"cache" json:"cache" validate:"required" register:"skip" description:"Cache connection (startup-only, not runtime configurable)"`

	Logger logger.Config  `mapstructure:"logger" json:"logger" validate:"required" register:"auto" description:"Logger module (runtime logging configuration)"`
	I18n   i18n.Config    `mapstructure:"i18n" json:"i18n" validate:"required" register:"auto" description:"Internationalization module (language and translations)"`
	Auth   authmod.Config `mapstructure:"auth" json:"auth" validate:"required" register:"auto" description:"Authentication module (includes JWT and security settings)"`
}
