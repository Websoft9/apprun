# Configuration Management

This directory contains the config management features.

## Configuration Structure

The `Config` struct in types.go defines all available configuration options:

- **App**: Application-level settings (name, version)
- **Database**: Database connection parameters
- Any other configuration items

## Configuration Definition Rules

All configuration items must be defined in types.go with appropriate struct tags:

- `validate`: Validation rules (e.g., `required`, `min=1`, `oneof=postgres mysql`)
- `default`: Default values for optional fields
- `db`: Database storage flag (`true` for storable in DB, `false` for runtime-only)

## Database Configuration Processing

Database configuration items are processed through struct reflection on the `Config` struct. Fields marked with `db:"true"` are stored in the database and can be dynamically updated at runtime.

### Reflection Processing

The system uses Go's reflection capabilities to:

- Parse struct tags for validation and defaults
- Identify database-storable fields
- Generate database schemas automatically
- Validate configuration values at runtime

### Example

```go
type Config struct {
    App struct {
        Name string `validate:"required" default:"apprun" db:"true"`
    }
}
```

In this example:
- `Name` is required and has a default value
- It can be stored in the database (`db:"true"`)

## Usage

Configuration is loaded and validated using the defined struct. Database-backed configurations can be updated through the API endpoints, while maintaining type safety and validation rules.