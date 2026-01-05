package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundle          *i18n.Bundle
	defaultLanguage string
	supportedLangs  []string
	supportedTags   []language.Tag // Cached parsed language tags for performance
	once            sync.Once
	initError       error
)

// Init initializes the i18n system with the provided configuration
func InitWithConfig(cfg Config) error {
	return Init(cfg.DefaultLanguage, cfg.SupportedLanguages, cfg.TranslationsPath)
}

// Init initializes the i18n system by loading translation files from the specified path
func Init(defaultLang string, supportedLanguages []string, translationsPath string) error {
	once.Do(func() {
		defaultLanguage = defaultLang
		supportedLangs = supportedLanguages

		// Parse default language tag
		tag, err := language.Parse(defaultLang)
		if err != nil {
			initError = fmt.Errorf("invalid default language '%s': %w", defaultLang, err)
			return
		}

		// Pre-parse supported language tags for performance (cached for hot path)
		supportedTags = make([]language.Tag, 0, len(supportedLanguages))
		for _, lang := range supportedLanguages {
			if t, err := language.Parse(lang); err == nil {
				supportedTags = append(supportedTags, t)
			}
		}

		// Create bundle
		bundle = i18n.NewBundle(tag)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		// Load translations from path
		if translationsPath != "" {
			if err := loadTranslations(translationsPath); err != nil {
				initError = fmt.Errorf("failed to load translations from '%s': %w", translationsPath, err)
				return
			}
		}
	})

	return initError
}

// loadTranslations loads translation files from the specified directory
func loadTranslations(translationsPath string) error {
	if _, err := os.Stat(translationsPath); os.IsNotExist(err) {
		return fmt.Errorf("translations path does not exist: %s", translationsPath)
	}

	entries, err := os.ReadDir(translationsPath)
	if err != nil {
		return fmt.Errorf("failed to read translations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if filepath.Ext(fileName) != ".toml" {
			continue
		}

		fullPath := filepath.Join(translationsPath, fileName)
		if _, err := bundle.LoadMessageFile(fullPath); err != nil {
			return fmt.Errorf("failed to load translation file '%s': %w", fileName, err)
		}
	}

	return nil
}

// Translate translates a message ID with optional template data
func Translate(lang, messageID string, data map[string]interface{}) string {
	if bundle == nil {
		return messageID // Fallback if not initialized
	}

	localizer := i18n.NewLocalizer(bundle, lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})

	if err != nil {
		// Fallback to message ID if translation not found
		return messageID
	}

	return msg
}

// GetSupportedLanguages returns the list of supported languages
func GetSupportedLanguages() []string {
	return supportedLangs
}

// IsSupported checks if a language is supported
// Also supports language variants (e.g., "zh" matches "zh-CN")
func IsSupported(lang string) bool {
	// Direct match
	for _, supported := range supportedLangs {
		if supported == lang {
			return true
		}
	}

	// Try language variant matching using cached tags (zh -> zh-CN, zh-TW, etc.)
	tag, err := language.Parse(lang)
	if err != nil {
		return false
	}

	base, _ := tag.Base()
	for _, supportedTag := range supportedTags {
		supportedBase, _ := supportedTag.Base()
		if base == supportedBase {
			return true
		}
	}

	return false
}

// GetMatchedLanguage returns the best matching supported language
// Handles language variants (e.g., "zh" -> "zh-CN")
func GetMatchedLanguage(lang string) string {
	// Direct exact match
	for _, supported := range supportedLangs {
		if supported == lang {
			return lang
		}
	}

	// Try to find variant match using cached tags
	tag, err := language.Parse(lang)
	if err != nil {
		return defaultLanguage
	}

	base, _ := tag.Base()
	for i, supportedTag := range supportedTags {
		supportedBase, _ := supportedTag.Base()
		if base == supportedBase {
			return supportedLangs[i] // Return first match (e.g., "zh" -> "zh-CN")
		}
	}

	return defaultLanguage
}

// GetDefaultLanguage returns the default language
func GetDefaultLanguage() string {
	return defaultLanguage
}
