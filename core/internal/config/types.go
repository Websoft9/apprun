package config

// This struct is the root define source of all configuration items

type Config struct {
	App struct {
		Name    string `validate:"required,min=1" default:"apprun" db:"true"`
		Version string `validate:"required" default:"1.0.0" db:"false"`
	} `validate:"required"`

	Database struct {
		Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"`
		Host     string `validate:"required" default:"localhost" db:"false"`
		Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`
		User     string `validate:"required" default:"postgres" db:"false"`
		Password string `validate:"required,min=8" db:"false"`
		DBName   string `validate:"required" default:"apprun" db:"false"`
	} `validate:"required"`

	POC struct {
		Enabled  bool   `default:"true" db:"true"`
		Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"`
		APIKey   string `validate:"required,min=10" db:"true"`
	} `validate:"required"`
}
