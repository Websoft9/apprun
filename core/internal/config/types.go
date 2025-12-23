package config

// Config 结构体定义配置项，包含校验、默认值和DB存储标记
type Config struct {
	App struct {
		Name    string `validate:"required,min=1" default:"apprun" db:"true"` // 校验必需，默认值，可存储DB
		Version string `validate:"required" default:"1.0.0" db:"false"`       // 校验必需，默认值，不存储DB
	} `validate:"required"`

	Database struct {
		Driver   string `validate:"required,oneof=postgres mysql" default:"postgres" db:"false"` // 校验必需，默认值，不存储DB
		Host     string `validate:"required" default:"localhost" db:"false"`                     // 校验必需，默认值，不存储DB
		Port     int    `validate:"required,min=1,max=65535" default:"5432" db:"false"`          // 校验必需，默认值，不存储DB
		User     string `validate:"required" default:"postgres" db:"false"`                      // 校验必需，默认值，不存储DB
		Password string `validate:"required,min=8" db:"false"`                                   // 校验必需，不存储DB（敏感）
		DBName   string `validate:"required" default:"apprun" db:"false"`                        // 校验必需，默认值，不存储DB
	} `validate:"required"`

	POC struct {
		Enabled  bool   `default:"true" db:"true"`                                                                   // 默认值，可存储DB
		Database string `validate:"required,url" default:"postgres://user:pass@localhost:5432/apprun_poc" db:"true"` // 校验必需，默认值，可存储DB
		APIKey   string `validate:"required,min=10" db:"true"`                                                       // 校验必需，可存储DB
	} `validate:"required"`
}
