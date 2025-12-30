package config

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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

// GetConfigValue retrieves config value by key with source information
func (s *Service) GetConfigValue(ctx context.Context, key string) (string, string, error) {
	// Try to get from database first (for dynamic configs)
	value, isDynamic, err := s.provider.GetConfig(ctx, key)
	if err == nil && isDynamic {
		return value, "database", nil
	}

	// Get from loaded config instance (file, env, or defaults)
	if s.cfg != nil {
		if val := s.getValueFromConfig(key); val != "" {
			return val, "file", nil
		}
	}

	// Fallback to tag default value
	meta, exists := s.loader.GetMetadata(key)
	if exists && meta.DefaultVal != "" {
		return meta.DefaultVal, "default", nil
	}

	return "", "", fmt.Errorf("config key has no value: %s", key)
}

// getValueFromConfig extracts value from loaded config using reflection
func (s *Service) getValueFromConfig(key string) string {
	if s.cfg == nil {
		return ""
	}

	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return ""
	}

	v := reflect.ValueOf(s.cfg).Elem()

	// Navigate through nested structs
	for i, part := range parts {
		// Find field by matching yaml tag or field name (case-insensitive)
		field, found := s.findField(v, part)
		if !found {
			return ""
		}

		// If last part, get the value
		if i == len(parts)-1 {
			return formatValue(field)
		}

		// Continue to nested struct
		if field.Kind() == reflect.Struct {
			v = field
		} else {
			return ""
		}
	}

	return ""
}

// findField finds a struct field by name (case-insensitive) or yaml tag
func (s *Service) findField(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	nameLower := strings.ToLower(name)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Check yaml tag
		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
			if strings.Split(yamlTag, ",")[0] == nameLower {
				return fieldValue, true
			}
		}

		// Check field name (case-insensitive)
		if strings.ToLower(field.Name) == nameLower {
			return fieldValue, true
		}
	}

	return reflect.Value{}, false
}

// formatValue converts reflect.Value to string
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())
	default:
		return ""
	}
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

	// 使用 validator 进行值验证（如果有 validate 标签）
	if meta.ValidateTag != "" {
		if err := s.validator.Var(value, meta.ValidateTag); err != nil {
			return fmt.Errorf("validation failed for key '%s': %w", key, err)
		}
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
