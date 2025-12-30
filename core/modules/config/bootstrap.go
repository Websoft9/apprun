package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"apprun/ent"
	internalConfig "apprun/internal/config"
)

// Bootstrap 配置引导器，统一管理配置初始化流程
type Bootstrap struct {
	configDir string
}

// NewBootstrap 创建配置引导器
func NewBootstrap(configDir string) *Bootstrap {
	// 如果配置目录不存在，使用默认路径
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		configDir = filepath.Join(".", "config")
	}
	return &Bootstrap{configDir: configDir}
}

// LoadInitialConfig 加载初始配置（不依赖数据库）
// 这是启动流程的第一步，用于获取数据库连接信息
func (b *Bootstrap) LoadInitialConfig(ctx context.Context) (*internalConfig.Config, error) {
	// 创建没有数据库支持的加载器
	loader, err := NewLoader(b.configDir, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	// 加载配置（从文件、环境变量等）
	cfg, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// InitDatabase 使用配置初始化数据库连接
// 环境变量优先级高于配置文件
func (b *Bootstrap) InitDatabase(ctx context.Context, cfg *internalConfig.Config) (*ent.Client, error) {
	// 环境变量优先，如果没有则使用配置文件
	dbHost := getEnvOrDefault("DB_HOST", cfg.Database.Host)
	dbPort := getEnvOrDefault("DB_PORT", fmt.Sprintf("%d", cfg.Database.Port))
	dbUser := getEnvOrDefault("DB_USER", cfg.Database.User)
	dbPass := getEnvOrDefault("DB_PASSWORD", cfg.Database.Password)
	dbName := getEnvOrDefault("DB_NAME", cfg.Database.DBName)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移 schema
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return client, nil
}

// CreateService 创建配置服务（包含数据库支持）
// 这是启动流程的第三步，创建完整的配置服务以支持动态配置
func (b *Bootstrap) CreateService(ctx context.Context, dbClient *ent.Client) (*Service, error) {
	// 创建配置仓储
	repo := NewRepository(dbClient)

	// 创建配置加载器（带数据库支持）
	loader, err := NewLoader(b.configDir, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader with DB support: %w", err)
	}

	// 创建配置服务
	service := NewService(loader, repo)

	// 重新加载配置（现在包含数据库层）
	_, err = service.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to reload config with DB: %w", err)
	}

	return service, nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
