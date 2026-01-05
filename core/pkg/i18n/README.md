# i18n Internationalization Infrastructure

## Overview

The apprun i18n module provides multilingual support, including text translation, language detection, and localized formatting. Supports zh-CN and en-US languages.

## Architecture

- **pkg/i18n**: Core translation engine
- **internal/middleware**: HTTP language detection middleware
- **locales/**: Translation files storage directory

## Configuration

Configure in `config/default.yaml`:

```yaml
i18n:
  default_language: "en-US"
  supported_languages: ["en-US", "zh-CN"]
  translations_path: "./locales"
```

## Usage

### Initialization
```go
import "core/pkg/i18n"

config := i18n.Config{
    DefaultLanguage:    "en-US",
    SupportedLanguages: []string{"en-US", "zh-CN"},
    TranslationsPath:   "./locales",
}
err := i18n.Init(config)
```

### Translation
```go
// Basic translation
msg := i18n.Translate("zh-CN", "errors.not_found", nil)

// Context translation
msg := i18n.TranslateContext(r.Context(), "success.created", nil)
```

### Middleware
```go
router.Use(middleware.LanguageDetector())
```

## Translation Files

- `active.en-US.toml`: English translations
- `active.zh-CN.toml`: Chinese translations

Example format:
```toml
[errors.not_found]
other = "Resource not found"
```

## Language Detection Priority

1. Query parameter: `?lang=zh-CN`
2. Cookie: `lang=zh-CN`
3. Accept-Language header
4. Default language

## Testing

```bash
go test ./pkg/i18n/...
go test ./internal/middleware/...
```

## Maintenance

- Add new language: Add to `supported_languages` and create corresponding translation file.
- Update translations: Modify `.toml` files, restart service to take effect.
- Performance: Translation calls < 1ms, supports high concurrency.
