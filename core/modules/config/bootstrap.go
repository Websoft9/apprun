package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	internalConfig "apprun/internal/config"
	"apprun/pkg/database"
)

// Bootstrap 配置引导器，统一管理配置初始化流程
type Bootstrap struct {
	configDir string
	registry  *ConfigRegistry // 模块配置注册表
}

// NewBootstrap 创建配置引导器
func NewBootstrap(configDir string) *Bootstrap {
	return NewBootstrapWithRegistry(configDir, nil)
}

// NewBootstrapWithRegistry 创建支持模块注册的配置引导器
func NewBootstrapWithRegistry(configDir string, registry *ConfigRegistry) *Bootstrap {
	// 如果配置目录不存在，使用默认路径
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		configDir = filepath.Join(".", "config")
	}
	return &Bootstrap{
		configDir: configDir,
		registry:  registry,
	}
}

// LoadInitialConfig 加载初始配置（不依赖数据库）
// 这是启动流程的第一步，用于获取数据库连接信息
func (b *Bootstrap) LoadInitialConfig(ctx context.Context) (*internalConfig.Config, error) {
	// 创建没有数据库支持的加载器（但支持模块注册）
	loader, err := NewLoaderWithRegistry(b.configDir, nil, b.registry)
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

// CreateService 创建配置服务（接收外部数据库客户端）
// 数据库连接由调用方负责（通过 pkg/database）
func (b *Bootstrap) CreateService(ctx context.Context, dbClient database.Client) (*Service, error) {
	// 获取底层 Ent client
	entClient := dbClient.GetEntClient()

	// 创建配置仓储
	repo := NewRepository(entClient)

	// 创建配置加载器（带数据库支持和注册表）
	loader, err := NewLoaderWithRegistry(b.configDir, repo, b.registry)
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
