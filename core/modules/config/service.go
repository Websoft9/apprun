package config

import (
	"context"
	"encoding/json"
	"fmt"

	"apprun/internal/config"

	"github.com/go-playground/validator/v10"
)

// Service 配置服务，提供业务逻辑
type Service struct {
	loader    *Loader
	provider  ConfigProvider
	validator *validator.Validate
	cfg       *config.Config // 缓存的配置实例
}

// NewService 创建配置服务
func NewService(loader *Loader, provider ConfigProvider) *Service {
	return &Service{
		loader:    loader,
		provider:  provider,
		validator: validator.New(),
	}
}

// LoadConfig 加载配置（启动时调用）
func (s *Service) LoadConfig(ctx context.Context) (*config.Config, error) {
	cfg, err := s.loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 验证配置
	if err := s.validator.Struct(cfg); err != nil {
		return cfg, fmt.Errorf("config validation failed: %w", err)
	}

	s.cfg = cfg
	return cfg, nil
}

// GetConfig 获取当前配置（用于 API）
func (s *Service) GetConfig() *config.Config {
	return s.cfg
}

// GetConfigValue 根据 key 获取配置值
func (s *Service) GetConfigValue(ctx context.Context, key string) (string, string, error) {
	// 尝试从数据库获取
	value, isDynamic, err := s.provider.GetConfig(ctx, key)
	if err == nil {
		source := "database"
		if !isDynamic {
			source = "file" // 静态配置从文件加载
		}
		return value, source, nil
	}

	// 数据库中不存在，从加载器元数据获取默认值
	meta, exists := s.loader.GetMetadata(key)
	if !exists {
		return "", "", fmt.Errorf("unknown config key: %s", key)
	}

	if meta.DefaultVal != "" {
		return meta.DefaultVal, "default", nil
	}

	return "", "", fmt.Errorf("config key has no value: %s", key)
}

// UpdateConfig 更新动态配置项
func (s *Service) UpdateConfig(ctx context.Context, key string, value string) error {
	// 验证 key 是否允许数据库存储
	if !s.loader.AllowDatabaseStorage(key) {
		return fmt.Errorf("config key '%s' is not allowed to be stored in database (db:false)", key)
	}

	// 验证值是否符合规则
	meta, exists := s.loader.GetMetadata(key)
	if !exists {
		return fmt.Errorf("unknown config key: %s", key)
	}

	// TODO: 实现基于 validate 标签的值验证
	// 当前简化实现，仅检查非空
	if value == "" && meta.ValidateTag != "" {
		return fmt.Errorf("config value cannot be empty for key: %s", key)
	}

	// 持久化到数据库
	if err := s.provider.SetConfig(ctx, key, value); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 重新加载配置以应用变更
	newCfg, err := s.loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to reload config after update: %w", err)
	}

	// 验证新配置
	if err := s.validator.Struct(newCfg); err != nil {
		// 回滚：删除刚刚设置的值
		_ = s.provider.DeleteConfig(ctx, key)
		return fmt.Errorf("new config validation failed, rolled back: %w", err)
	}

	s.cfg = newCfg
	return nil
}

// ListDynamicConfigs 列出所有动态配置项
func (s *Service) ListDynamicConfigs(ctx context.Context) (map[string]string, error) {
	return s.provider.ListDynamicConfigs(ctx)
}

// DeleteDynamicConfig 删除动态配置项
func (s *Service) DeleteDynamicConfig(ctx context.Context, key string) error {
	// 验证 key 是否允许数据库存储
	if !s.loader.AllowDatabaseStorage(key) {
		return fmt.Errorf("config key '%s' is not a dynamic config (db:false)", key)
	}

	if err := s.provider.DeleteConfig(ctx, key); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// 重新加载配置
	newCfg, err := s.loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to reload config after deletion: %w", err)
	}

	s.cfg = newCfg
	return nil
}

// GetConfigAsJSON 获取完整配置的 JSON 表示
func (s *Service) GetConfigAsJSON() (string, error) {
	if s.cfg == nil {
		return "", fmt.Errorf("config not loaded")
	}

	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(data), nil
}

// GetAllowedDynamicKeys 获取所有允许动态配置的键（db:true）
func (s *Service) GetAllowedDynamicKeys() []string {
	var keys []string
	for key, meta := range s.loader.metadata {
		if meta.AllowDB {
			keys = append(keys, key)
		}
	}
	return keys
}
