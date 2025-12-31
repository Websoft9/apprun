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

	POC struct {
		Enabled  bool   `yaml:"enabled" default:"true" db:"true"`
		Database string `yaml:"database" validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
		APIKey   string `yaml:"api_key" validate:"required,min=10" db:"true"`
	} `yaml:"poc" validate:"required"`
}
