package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"apprun/internal/config"

	"github.com/spf13/viper"
)

// Loader 配置加载器，实现 6 层优先级系统
type Loader struct {
	configDir string                // 配置文件目录
	provider  ConfigProvider        // 数据库配置提供者
	viper     *viper.Viper          // Viper 实例
	metadata  map[string]*fieldMeta // 字段元数据（从反射提取）
}

// fieldMeta 字段元数据
type fieldMeta struct {
	Key         string // 配置键路径，如 "app.name"
	DefaultVal  string // 默认值（从 default 标签）
	AllowDB     bool   // 是否允许数据库存储（db 标签）
	ValidateTag string // 验证规则（validate 标签）
}

// NewLoader 创建配置加载器
func NewLoader(configDir string, provider ConfigProvider) (*Loader, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv() // 自动绑定环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	loader := &Loader{
		configDir: configDir,
		provider:  provider,
		viper:     v,
		metadata:  make(map[string]*fieldMeta),
	}

	// 使用反射提取字段元数据
	if err := loader.extractMetadata(); err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	return loader, nil
}

// extractMetadata 使用反射提取 Config 结构体的标签元数据
func (l *Loader) extractMetadata() error {
	cfg := config.Config{}
	t := reflect.TypeOf(cfg)

	return l.walkStruct(t, "")
}

// walkStruct 递归遍历结构体字段
func (l *Loader) walkStruct(t reflect.Type, prefix string) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		// 获取 yaml 标签作为键名
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			yamlTag = strings.ToLower(field.Name)
		}

		// 构建配置路径
		var path string
		if prefix == "" {
			path = yamlTag
		} else {
			path = prefix + "." + yamlTag
		}

		// 如果是嵌套结构体，递归处理
		if fieldType.Kind() == reflect.Struct {
			if err := l.walkStruct(fieldType, path); err != nil {
				return err
			}
			continue
		}

		// 提取标签
		defaultVal := field.Tag.Get("default")
		dbTag := field.Tag.Get("db")
		validateTag := field.Tag.Get("validate")

		// 解析 db 标签（默认为 false）
		allowDB := false
		if dbTag != "" {
			var err error
			allowDB, err = strconv.ParseBool(dbTag)
			if err != nil {
				return fmt.Errorf("invalid db tag for field %s: %s", field.Name, dbTag)
			}
		}

		// 保存元数据
		l.metadata[path] = &fieldMeta{
			Key:         path,
			DefaultVal:  defaultVal,
			AllowDB:     allowDB,
			ValidateTag: validateTag,
		}
	}

	return nil
}

// Load 加载配置，按照 6 层优先级顺序
// 优先级：Layer 1 (标签默认值) < Layer 2 (default.yaml) < Layer 3 (专用文件)
//
//	< Layer 4 (conf_d) < Layer 5 (数据库) < Layer 6 (环境变量)
func (l *Loader) Load(ctx context.Context) (*config.Config, error) {
	cfg := &config.Config{}

	// Layer 1: 应用标签默认值
	if err := l.applyTagDefaults(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply tag defaults: %w", err)
	}

	// Layer 2: 加载 default.yaml
	if err := l.loadDefaultYAML(); err != nil {
		return nil, fmt.Errorf("failed to load default.yaml: %w", err)
	}

	// Layer 3: 加载专用配置文件（如 database.yaml, server.yaml）
	if err := l.loadSpecializedFiles(); err != nil {
		return nil, fmt.Errorf("failed to load specialized files: %w", err)
	}

	// Layer 4: 加载 conf_d 目录下的配置文件
	if err := l.loadConfD(); err != nil {
		return nil, fmt.Errorf("failed to load conf_d: %w", err)
	}

	// 将 Viper 配置解析到结构体
	if err := l.viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Layer 5: 从数据库覆盖动态配置（只覆盖 db:true 的字段）
	if err := l.applyDatabaseConfig(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to apply database config: %w", err)
	}

	// Layer 6: 环境变量（Viper 自动处理，优先级最高）
	// 已由 AutomaticEnv() 启用

	return cfg, nil
}

// applyTagDefaults 应用标签默认值（Layer 1）
func (l *Loader) applyTagDefaults(cfg *config.Config) error {
	for key, meta := range l.metadata {
		if meta.DefaultVal != "" {
			l.viper.SetDefault(key, meta.DefaultVal)
		}
	}
	return nil
}

// loadDefaultYAML 加载 default.yaml（Layer 2）
func (l *Loader) loadDefaultYAML() error {
	defaultFile := filepath.Join(l.configDir, "default.yaml")
	if _, err := os.Stat(defaultFile); err == nil {
		l.viper.SetConfigFile(defaultFile)
		if err := l.viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read default.yaml: %w", err)
		}
	}
	return nil
}

// loadSpecializedFiles 加载专用配置文件（Layer 3）
func (l *Loader) loadSpecializedFiles() error {
	specialFiles := []string{"database.yaml", "server.yaml", "poc.yaml"}
	for _, fname := range specialFiles {
		fpath := filepath.Join(l.configDir, fname)
		if _, err := os.Stat(fpath); err == nil {
			// 临时设置配置文件路径并合并
			tmpViper := viper.New()
			tmpViper.SetConfigFile(fpath)
			if err := tmpViper.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to read %s: %w", fname, err)
			}
			// 合并到主 viper 实例
			if err := l.viper.MergeConfigMap(tmpViper.AllSettings()); err != nil {
				return fmt.Errorf("failed to merge %s: %w", fname, err)
			}
		}
	}
	return nil
}

// loadConfD 加载 conf_d 目录下的配置文件（Layer 4）
func (l *Loader) loadConfD() error {
	confDDir := filepath.Join(l.configDir, "conf_d")
	if _, err := os.Stat(confDDir); os.IsNotExist(err) {
		return nil // conf_d 目录不存在，跳过
	}

	entries, err := os.ReadDir(confDDir)
	if err != nil {
		return fmt.Errorf("failed to read conf_d directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		fpath := filepath.Join(confDDir, entry.Name())
		// 使用临时 viper 实例读取并合并
		tmpViper := viper.New()
		tmpViper.SetConfigFile(fpath)
		if err := tmpViper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}
		if err := l.viper.MergeConfigMap(tmpViper.AllSettings()); err != nil {
			return fmt.Errorf("failed to merge %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// applyDatabaseConfig 从数据库覆盖动态配置（Layer 5）
func (l *Loader) applyDatabaseConfig(ctx context.Context, cfg *config.Config) error {
	if l.provider == nil {
		return nil // 没有数据库提供者，跳过
	}

	dbConfigs, err := l.provider.ListDynamicConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list database configs: %w", err)
	}

	// 只覆盖 db:true 的字段
	for key, value := range dbConfigs {
		meta, exists := l.metadata[key]
		if !exists {
			continue // 未知配置项，跳过
		}

		if !meta.AllowDB {
			continue // db:false，不允许数据库覆盖
		}

		// 设置到 Viper（覆盖之前的值）
		l.viper.Set(key, value)
	}

	// 重新解析到结构体
	if err := l.viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to re-unmarshal after database config: %w", err)
	}

	return nil
}

// GetMetadata 获取字段元数据（用于验证和服务层）
func (l *Loader) GetMetadata(key string) (*fieldMeta, bool) {
	meta, exists := l.metadata[key]
	return meta, exists
}

// AllowDatabaseStorage 检查配置项是否允许数据库存储
func (l *Loader) AllowDatabaseStorage(key string) bool {
	meta, exists := l.metadata[key]
	if !exists {
		return false
	}
	return meta.AllowDB
}
