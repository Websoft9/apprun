// Package config defines the configuration types and structures for the application.
package config

// Config is the root define source of all configuration items
// YAML keys are implicitly mapped from lowercase field names (e.g., Name -> name)
// Only nested structs require explicit yaml tags to define root keys
type Config struct {
	App struct {
		Name     string `validate:"required,min=1" default:"apprun" db:"true"`
		Version  string `validate:"required" default:"1.0.0" db:"false"`
		Timezone string `validate:"required,timezone" default:"Asia/Shanghai" db:"true"`
	} `yaml:"app" validate:"required"`

	Database struct {
		Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
		Host     string `validate:"required" default:"localhost" db:"false"`
		Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
		User     string `validate:"required" default:"postgres" db:"false"`
		Password string `yaml:"password" validate:"required,min=8" db:"false"`
		DBName   string `yaml:"dbname" validate:"required" default:"apprun" db:"false"`
	} `yaml:"database" validate:"required"`

	Cache struct {
		Host       string `yaml:"host" mapstructure:"host" validate:"required" default:"localhost" db:"false"`
		Port       string `yaml:"port" mapstructure:"port" validate:"required" default:"6379" db:"false"`
		Password   string `yaml:"password" mapstructure:"password" default:"" db:"false"`
		DB         int    `yaml:"db" mapstructure:"db" validate:"min=0,max=15" default:"0" db:"false"`
		PoolSize   int    `yaml:"pool_size" mapstructure:"pool_size" validate:"min=1" default:"10" db:"false"`
		Timeout    string `yaml:"timeout" mapstructure:"timeout" validate:"required" default:"2s" db:"false"`
		MaxRetries int    `yaml:"max_retries" mapstructure:"max_retries" validate:"min=0" default:"3" db:"false"`
		FailOpen   bool   `yaml:"fail_open" mapstructure:"fail_open" default:"true" db:"false"`
		TLSEnabled bool   `yaml:"tls_enabled" mapstructure:"tls_enabled" default:"false" db:"false"`
	} `yaml:"cache" validate:"required"`
}
