// Package i18n provides internationalization support for the application,
// including message translation, language detection, and locale management.
package i18n

// Config defines the configuration for the i18n system
// Follows internal/config/types.go standards for consistency with config center
type Config struct {
	// DefaultLanguage is the fallback language when user preference is not available
	DefaultLanguage string `yaml:"default_language" json:"default_language" default:"en-US" db:"true" validate:"required"`

	// SupportedLanguages is the list of languages supported by the application
	SupportedLanguages []string `yaml:"supported_languages" json:"supported_languages" default:"en-US,zh-CN" db:"true" validate:"required,min=1"`

	// TranslationsPath is the directory path containing translation files
	TranslationsPath string `yaml:"translations_path" json:"translations_path" default:"./locales" db:"true" validate:"required"`
}

// DefaultConfig returns the default i18n configuration
func DefaultConfig() Config {
	return Config{
		DefaultLanguage:    "en-US",
		SupportedLanguages: []string{"en-US", "zh-CN"},
		TranslationsPath:   "./locales",
	}
}
