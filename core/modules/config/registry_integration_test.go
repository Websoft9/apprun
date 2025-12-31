package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"apprun/pkg/logger"
)

// TestRegistryIntegration_LoggerModule tests the full registry workflow with logger module
func TestRegistryIntegration_LoggerModule(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()
	defaultConfigPath := filepath.Join(tempDir, "default.yaml")

	// 写入基础配置文件
	configContent := `
app:
  name: "testapp"
  version: "1.0.0"

database:
  driver: "postgres"
  host: "127.0.0.1"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "testdb"

poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/test"
  apikey: "testapikey1234567890"

logger:
  level: "debug"
  output:
    targets: ["stdout", "file"]
    file_path: "/var/log/apprun.log"
    max_size: 100
`
	err := os.WriteFile(defaultConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 创建配置注册表
	registry := NewRegistry()

	// 注册 logger 模块配置
	loggerConfig := &logger.Config{}
	err = registry.Register("logger", loggerConfig)
	require.NoError(t, err)

	// 验证注册成功
	assert.True(t, registry.Has("logger"))
	assert.Equal(t, 1, registry.Count())

	// 创建带注册表的加载器
	loader, err := NewLoaderWithRegistry(tempDir, nil, registry)
	require.NoError(t, err)

	// 加载配置
	ctx := context.Background()
	config, err := loader.Load(ctx)
	require.NoError(t, err)

	// 验证基础配置正确加载
	assert.Equal(t, "testapp", config.App.Name)
	assert.Equal(t, "127.0.0.1", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)

	// 验证注册表中的配置元数据被提取

	// 验证 logger.level 的元数据
	levelMeta, exists := loader.GetMetadata("logger.level")
	require.True(t, exists, "logger.level metadata should exist")
	assert.True(t, levelMeta.AllowDB, "logger.level should allow DB updates")
	assert.Equal(t, "oneof=debug info warn error", levelMeta.ValidateTag)
	assert.Equal(t, "info", levelMeta.DefaultVal)

	// 验证 logger.output.targets 的元数据
	targetsMeta, exists := loader.GetMetadata("logger.output.targets")
	require.True(t, exists, "logger.output.targets metadata should exist")
	assert.True(t, targetsMeta.AllowDB, "logger.output.targets should allow DB updates")
	assert.Contains(t, targetsMeta.ValidateTag, "dive")
}

// TestRegistryIntegration_UpdateValidation tests db tag validation for registered modules
func TestRegistryIntegration_UpdateValidation(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()
	defaultConfigPath := filepath.Join(tempDir, "default.yaml")

	// 写入基础配置文件
	configContent := `
app:
  name: "testapp"
  version: "1.0.0"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "testdb"

poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/test"
  apikey: "testapikey1234567890"

logger:
  level: "info"
  output:
    targets: ["stdout"]
`
	err := os.WriteFile(defaultConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 创建配置注册表并注册 logger 模块
	registry := NewRegistry()
	err = registry.Register("logger", &logger.Config{})
	require.NoError(t, err)

	// 创建带注册表的加载器
	loader, err := NewLoaderWithRegistry(tempDir, nil, registry)
	require.NoError(t, err)

	// 加载配置
	ctx := context.Background()
	_, err = loader.Load(ctx)
	require.NoError(t, err)

	// 测试案例：logger.level 有 db:"true" 标签，应该允许动态更新
	levelMeta, exists := loader.GetMetadata("logger.level")
	require.True(t, exists)
	assert.True(t, levelMeta.AllowDB, "logger.level should allow DB updates")

	// 测试案例：验证 validate 标签正确
	assert.Equal(t, "oneof=debug info warn error", levelMeta.ValidateTag)

	// 模拟验证逻辑（实际在 Service.UpdateConfig 中）
	// 这里只验证元数据是否正确提取
	assert.NotEmpty(t, levelMeta.ValidateTag)
	assert.Equal(t, "info", levelMeta.DefaultVal)
}

// TestRegistryIntegration_MultipleModules tests multiple module registration
func TestRegistryIntegration_MultipleModules(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()
	defaultConfigPath := filepath.Join(tempDir, "default.yaml")

	// 写入多模块配置文件
	configContent := `
app:
  name: "testapp"
  version: "1.0.0"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "testdb"

poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/test"
  apikey: "testapikey1234567890"

logger:
  level: "info"

user:
  max_sessions: 5
  session_timeout: 3600
`
	err := os.WriteFile(defaultConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 定义 user 模块配置结构
	type UserConfig struct {
		MaxSessions    int `yaml:"max_sessions" default:"3" db:"true" validate:"min=1,max=10"`
		SessionTimeout int `yaml:"session_timeout" default:"1800" db:"true" validate:"min=60"`
	}

	// 创建配置注册表并注册多个模块
	registry := NewRegistry()
	err = registry.Register("logger", &logger.Config{})
	require.NoError(t, err)
	err = registry.Register("user", &UserConfig{})
	require.NoError(t, err)

	// 验证注册成功
	assert.Equal(t, 2, registry.Count())
	assert.True(t, registry.Has("logger"))
	assert.True(t, registry.Has("user"))

	// 创建带注册表的加载器
	loader, err := NewLoaderWithRegistry(tempDir, nil, registry)
	require.NoError(t, err)

	// 加载配置
	ctx := context.Background()
	config, err := loader.Load(ctx)
	require.NoError(t, err)

	// 验证基础配置正确加载
	assert.Equal(t, "testapp", config.App.Name)

	// 验证两个模块的元数据都被提取
	loggerMeta, exists := loader.GetMetadata("logger.level")
	require.True(t, exists)
	assert.True(t, loggerMeta.AllowDB)

	userMeta, exists := loader.GetMetadata("user.max_sessions")
	require.True(t, exists)
	assert.True(t, userMeta.AllowDB)
	assert.Equal(t, "min=1,max=10", userMeta.ValidateTag)
	assert.Equal(t, "3", userMeta.DefaultVal)
}

// TestRegistryIntegration_BackwardCompatibility tests that system works without registry
func TestRegistryIntegration_BackwardCompatibility(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()
	defaultConfigPath := filepath.Join(tempDir, "default.yaml")

	// 写入配置文件
	configContent := `
app:
  name: "testapp"
  version: "1.0.0"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword"
  dbname: "testdb"

poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/test"
  apikey: "testapikey1234567890"
`
	err := os.WriteFile(defaultConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 不使用注册表，使用旧的 NewLoader 方法
	loader, err := NewLoader(tempDir, nil)
	require.NoError(t, err)

	// 加载配置
	ctx := context.Background()
	config, err := loader.Load(ctx)
	require.NoError(t, err)

	// 验证配置正确加载
	assert.Equal(t, "testapp", config.App.Name)
	assert.Equal(t, "postgres", config.Database.Driver)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
}
