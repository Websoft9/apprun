package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"apprun/ent"
	"apprun/ent/configitem"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var (
	viperInstance *viper.Viper
	validate      *validator.Validate
	entClient     *ent.Client
)

// ConfigItem 用于API响应
type ConfigItem struct {
	Path        string      `json:"path"`
	Value       interface{} `json:"value"`
	DBStorable  bool        `json:"dbStorable"`
	Description string      `json:"description,omitempty"`
}

// DBProvider 实现Viper的RemoteProvider接口，从数据库加载配置
type DBProvider struct {
	client *ent.Client
}

func (p *DBProvider) Get(rp viper.RemoteProvider) ([]byte, error) {
	if p.client == nil {
		return nil, fmt.Errorf("database client not initialized")
	}

	items, err := p.client.Configitem.Query().
		Where(configitem.IsDynamic(true)).
		All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query config from DB: %w", err)
	}

	configMap := make(map[string]interface{})
	for _, item := range items {
		var value interface{}
		if err := json.Unmarshal([]byte(item.Value), &value); err != nil {
			log.Printf("Warning: failed to unmarshal config value for key %s: %v", item.Key, err)
			configMap[item.Key] = item.Value
		} else {
			configMap[item.Key] = value
		}
	}

	return json.Marshal(configMap)
}

func (p *DBProvider) Watch(rp viper.RemoteProvider) ([]byte, error) {
	return p.Get(rp)
}

func (p *DBProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	// 简化实现，不支持热重载
	return nil, nil
}

// InitConfig 初始化配置系统，包括数据库客户端
func InitConfig(client *ent.Client) error {
	entClient = client
	validate = validator.New()
	viperInstance = viper.New()

	return nil
}

// LoadConfig 使用Viper加载配置，支持多领域文件动态扫描和数据库集成
// 优先级：环境变量 > DB > conf_d/*.yaml > 领域配置文件 > default.yaml > 结构体默认值
func LoadConfig() (*Config, error) {
	if viperInstance == nil {
		viperInstance = viper.New()
		validate = validator.New()
	}

	v := viperInstance
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	// 步骤1: 设置结构体默认值（最低优先级）
	setDefaultsFromTags()

	// 步骤2: 加载 default.yaml
	v.SetConfigName("default")
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: could not read default config: %v", err)
	} else {
		log.Println("Loaded default.yaml")
	}

	// 步骤3: 动态扫描并加载领域配置文件（按字母排序）
	configDir := "./config"
	if entries, err := os.ReadDir(configDir); err == nil {
		var domainFiles []string
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}
			// 排除 default.yaml 和 conf_d 目录
			if entry.Name() == "default.yaml" {
				continue
			}
			domainFiles = append(domainFiles, entry.Name())
		}
		// 字母排序
		sort.Strings(domainFiles)

		for _, file := range domainFiles {
			v.SetConfigFile(filepath.Join(configDir, file))
			if err := v.MergeInConfig(); err != nil {
				log.Printf("Warning: could not merge domain config %s: %v", file, err)
			} else {
				log.Printf("Loaded domain config: %s", file)
			}
		}
	}

	// 步骤4: 加载 conf_d 目录下的配置文件
	confDPath := filepath.Join(configDir, "conf_d")
	if entries, err := os.ReadDir(confDPath); err == nil {
		var confDFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
				confDFiles = append(confDFiles, entry.Name())
			}
		}
		sort.Strings(confDFiles)

		for _, file := range confDFiles {
			v.SetConfigFile(filepath.Join(confDPath, file))
			if err := v.MergeInConfig(); err != nil {
				log.Printf("Warning: could not merge conf_d config %s: %v", file, err)
			} else {
				log.Printf("Loaded conf_d config: %s", file)
			}
		}
	}

	// 步骤5: 配置环境变量支持（在DB加载前设置，确保最高优先级）
	// 环境变量前缀为 W9_，路径分隔符 . 转换为 _
	// 示例: app.name → W9_APP_NAME
	v.SetEnvPrefix("W9")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 步骤6: 从数据库加载动态配置（如果数据库已初始化）
	// 注意：只有当环境变量不存在时，才使用DB中的值
	if entClient != nil {
		provider := &DBProvider{client: entClient}
		if data, err := provider.Get(nil); err != nil {
			log.Printf("Warning: failed to load config from DB, using file config: %v", err)
		} else {
			var dbConfig map[string]interface{}
			if err := json.Unmarshal(data, &dbConfig); err != nil {
				log.Printf("Warning: failed to unmarshal DB config: %v", err)
			} else {
				// 合并数据库配置，但不覆盖环境变量
				for key, value := range dbConfig {
					// 检查是否有对应的环境变量
					envKey := "W9_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
					if os.Getenv(envKey) == "" {
						// 环境变量不存在，使用DB值
						v.Set(key, value)
					} else {
						log.Printf("Environment variable %s overrides DB value for key: %s", envKey, key)
					}
				}
				log.Println("Loaded config from database (ENV variables have higher priority)")
			}
		}
	}

	// 步骤7: 反序列化到结构体
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 步骤8: 校验配置
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Println("Configuration loaded successfully with priority: ENV > DB > conf_d > domain files > default.yaml > struct tags")
	return &config, nil
}

// setDefaultsFromTags 从Config结构体的default标签设置默认值
func setDefaultsFromTags() {
	var cfg Config
	setDefaultsRecursive(reflect.ValueOf(&cfg).Elem(), "")
}

func setDefaultsRecursive(v reflect.Value, prefix string) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)

		// 构建配置路径
		fieldName := strings.ToLower(typeField.Name)
		path := fieldName
		if prefix != "" {
			path = prefix + "." + fieldName
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Struct {
			setDefaultsRecursive(field, path)
			continue
		}

		// 获取default标签
		if defaultValue := typeField.Tag.Get("default"); defaultValue != "" {
			viperInstance.SetDefault(path, parseDefaultValue(defaultValue, field.Kind()))
		}
	}
}

func parseDefaultValue(value string, kind reflect.Kind) interface{} {
	switch kind {
	case reflect.Bool:
		return value == "true"
	case reflect.Int, reflect.Int64:
		var i int
		fmt.Sscanf(value, "%d", &i)
		return i
	default:
		return value
	}
}

// GetAllConfigItems 获取所有配置项（用于GET /config API）
func GetAllConfigItems() ([]ConfigItem, error) {
	var items []ConfigItem
	var cfg Config

	collectConfigItems(reflect.ValueOf(&cfg).Elem(), reflect.TypeOf(cfg), "", &items)

	return items, nil
}

func collectConfigItems(v reflect.Value, t reflect.Type, prefix string, items *[]ConfigItem) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)

		fieldName := strings.ToLower(typeField.Name)
		path := fieldName
		if prefix != "" {
			path = prefix + "." + fieldName
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Struct {
			collectConfigItems(field, typeField.Type, path, items)
			continue
		}

		// 获取当前值
		value := viperInstance.Get(path)

		// 检查是否可存储到数据库
		dbStorable := typeField.Tag.Get("db") == "true"

		*items = append(*items, ConfigItem{
			Path:       path,
			Value:      value,
			DBStorable: dbStorable,
		})
	}
}

// UpdateConfig 更新配置项（用于PUT /config API）
func UpdateConfig(updates map[string]interface{}) error {
	if entClient == nil {
		return fmt.Errorf("database client not initialized")
	}

	// 验证所有要更新的配置项是否允许修改
	var cfg Config
	allowedKeys := make(map[string]bool)
	collectAllowedKeys(reflect.TypeOf(cfg), "", allowedKeys)

	for key := range updates {
		if !allowedKeys[key] {
			return fmt.Errorf("config key '%s' is not allowed to be modified (db:false or not exists)", key)
		}
	}

	// 开启事务
	ctx := context.Background()
	tx, err := entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// 保存旧值用于回滚
	oldValues := make(map[string]interface{})
	for key := range updates {
		oldValues[key] = viperInstance.Get(key)
	}

	// 更新内存配置
	for key, value := range updates {
		viperInstance.Set(key, value)
	}

	// 验证更新后的配置
	var newConfig Config
	if err := viperInstance.Unmarshal(&newConfig); err != nil {
		// 回滚内存配置
		for key, oldValue := range oldValues {
			viperInstance.Set(key, oldValue)
		}
		return fmt.Errorf("config unmarshal failed after update: %w", err)
	}

	if err := validate.Struct(&newConfig); err != nil {
		// 回滚内存配置
		for key, oldValue := range oldValues {
			viperInstance.Set(key, oldValue)
		}
		return fmt.Errorf("config validation failed after update: %w", err)
	}

	// 持久化到数据库
	for key, value := range updates {
		valueJSON, err := json.Marshal(value)
		if err != nil {
			tx.Rollback()
			// 回滚内存配置
			for k, oldValue := range oldValues {
				viperInstance.Set(k, oldValue)
			}
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		// 检查是否已存在
		existing, err := tx.Configitem.Query().Where(configitem.KeyEQ(key)).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			// 回滚内存配置
			for k, oldValue := range oldValues {
				viperInstance.Set(k, oldValue)
			}
			return fmt.Errorf("failed to query existing config for key %s: %w", key, err)
		}

		if existing != nil {
			// 更新现有记录
			_, err = tx.Configitem.UpdateOneID(existing.ID).
				SetValue(string(valueJSON)).
				SetIsDynamic(true).
				Save(ctx)
		} else {
			// 创建新记录
			_, err = tx.Configitem.Create().
				SetKey(key).
				SetValue(string(valueJSON)).
				SetIsDynamic(true).
				Save(ctx)
		}

		if err != nil {
			tx.Rollback()
			// 回滚内存配置
			for k, oldValue := range oldValues {
				viperInstance.Set(k, oldValue)
			}
			return fmt.Errorf("failed to save config to DB for key %s: %w", key, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		// 回滚内存配置
		for key, oldValue := range oldValues {
			viperInstance.Set(key, oldValue)
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully updated %d config items", len(updates))
	return nil
}

func collectAllowedKeys(t reflect.Type, prefix string, allowed map[string]bool) {
	for i := 0; i < t.NumField(); i++ {
		typeField := t.Field(i)

		fieldName := strings.ToLower(typeField.Name)
		path := fieldName
		if prefix != "" {
			path = prefix + "." + fieldName
		}

		// 处理嵌套结构体
		if typeField.Type.Kind() == reflect.Struct {
			collectAllowedKeys(typeField.Type, path, allowed)
			continue
		}

		// 只有db:"true"的字段才允许修改
		if typeField.Tag.Get("db") == "true" {
			allowed[path] = true
		}
	}
}
