package config
package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	// 应用配置
	AppName    string
	AppPort    string
	AppEnv     string

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT 配置
	JWTSecret        string
	JWTRefreshSecret string

	// Temporal 配置
	TemporalHost string

	// NATS 配置
	NATSUrl string

	// Redis 配置（可选）
	RedisHost string
	RedisPort string

	// Ory Kratos 配置
	KratosPublicURL string
	KratosAdminURL  string
}

// Load 加载配置
func Load() *Config {
	return &Config{
		// 应用配置
		AppName: getEnv("APP_NAME", "apprun-core"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		// 数据库配置
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "apprun"),

		// JWT 配置
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),

		// Temporal 配置
		TemporalHost: getEnv("TEMPORAL_HOST", "localhost:7233"),

		// NATS 配置
		NATSUrl: getEnv("NATS_URL", "nats://localhost:4222"),

		// Redis 配置
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		// Ory Kratos 配置
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
