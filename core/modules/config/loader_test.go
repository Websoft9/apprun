package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConfigProvider 模拟配置提供者（用于测试）
type mockConfigProvider struct {
	configs map[string]string
}

func newMockProvider() *mockConfigProvider {
	return &mockConfigProvider{
		configs: make(map[string]string),
	}
}

func (m *mockConfigProvider) GetConfig(ctx context.Context, key string) (string, bool, error) {
	val, exists := m.configs[key]
	return val, exists, nil
}

func (m *mockConfigProvider) SetConfig(ctx context.Context, key string, value string) error {
	m.configs[key] = value
	return nil
}

func (m *mockConfigProvider) ListDynamicConfigs(ctx context.Context) (map[string]string, error) {
	return m.configs, nil
}

func (m *mockConfigProvider) DeleteConfig(ctx context.Context, key string) error {
	delete(m.configs, key)
	return nil
}

// TestLoader_TagDefaults 测试 Layer 1: 标签默认值
func TestLoader_TagDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	// 检查元数据是否正确提取
	meta, exists := loader.GetMetadata("app.name")
	assert.True(t, exists, "app.name metadata should exist")
	assert.Equal(t, "apprun", meta.DefaultVal, "default value should match tag")
	assert.True(t, meta.AllowDB, "app.name should allow DB (db:true)")

	meta, exists = loader.GetMetadata("app.version")
	assert.True(t, exists)
	assert.Equal(t, "1.0.0", meta.DefaultVal)
	assert.False(t, meta.AllowDB, "app.version should not allow DB (db:false)")
}

// TestLoader_DefaultYAML 测试 Layer 2: default.yaml 覆盖
func TestLoader_DefaultYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 default.yaml
	defaultYAML := `
app:
  name: "test-app"
  version: "2.0.0"
database:
  host: "db.example.com"
  port: 3306
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)
	require.NoError(t, err)

	// default.yaml 应覆盖标签默认值
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "2.0.0", cfg.App.Version)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 3306, cfg.Database.Port)
}

// TestLoader_SpecializedFiles 测试 Layer 3: 专用配置文件
func TestLoader_SpecializedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 default.yaml
	defaultYAML := `
database:
  host: "localhost"
  port: 5432
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	// 创建 database.yaml（应覆盖 default.yaml）
	databaseYAML := `
database:
  host: "prod-db.example.com"
  port: 5433
`
	err = os.WriteFile(filepath.Join(tmpDir, "database.yaml"), []byte(databaseYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)
	require.NoError(t, err)

	// database.yaml 应覆盖 default.yaml
	assert.Equal(t, "prod-db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
}

// TestLoader_ConfD 测试 Layer 4: conf_d 目录
func TestLoader_ConfD(t *testing.T) {
	tmpDir := t.TempDir()
	confDDir := filepath.Join(tmpDir, "conf_d")
	require.NoError(t, os.Mkdir(confDDir, 0755))

	// 创建 default.yaml
	defaultYAML := `
poc:
  enabled: false
  database: "default-db"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	// 创建 conf_d/custom-poc.yaml（应覆盖 default.yaml）
	customYAML := `
poc:
  enabled: true
  database: "custom-poc-db"
`
	err = os.WriteFile(filepath.Join(confDDir, "custom-poc.yaml"), []byte(customYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)
	require.NoError(t, err)

	// conf_d 应覆盖 default.yaml
	assert.True(t, cfg.POC.Enabled)
	assert.Equal(t, "custom-poc-db", cfg.POC.Database)
}

// TestLoader_DatabaseOverride 测试 Layer 5: 数据库覆盖
func TestLoader_DatabaseOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 default.yaml
	defaultYAML := `
app:
  name: "file-app"
poc:
  enabled: false
  api_key: "file-key"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	// 模拟数据库配置（只有 db:true 的字段）
	mockProvider := newMockProvider()
	mockProvider.SetConfig(context.Background(), "app.name", "db-app")
	mockProvider.SetConfig(context.Background(), "poc.enabled", "true")

	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)
	require.NoError(t, err)

	// 数据库应覆盖文件配置（app.name 和 poc.enabled 都是 db:true）
	assert.Equal(t, "db-app", cfg.App.Name)
	assert.True(t, cfg.POC.Enabled)
}

// TestLoader_EnvOverride 测试 Layer 6: 环境变量覆盖
func TestLoader_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// 设置环境变量（Viper 自动将 . 转换为 _）
	os.Setenv("APP_NAME", "env-app")
	os.Setenv("DATABASE_HOST", "env-db.example.com")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("DATABASE_HOST")
	}()

	// 创建 default.yaml
	defaultYAML := `
app:
  name: "file-app"
database:
  host: "file-db.example.com"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := loader.Load(ctx)
	require.NoError(t, err)

	// 环境变量应覆盖文件配置
	assert.Equal(t, "env-app", cfg.App.Name)
	assert.Equal(t, "env-db.example.com", cfg.Database.Host)
}

// TestLoader_AllowDatabaseStorage 测试 db 标签控制
func TestLoader_AllowDatabaseStorage(t *testing.T) {
	tmpDir := t.TempDir()

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	// app.name 标记为 db:true
	assert.True(t, loader.AllowDatabaseStorage("app.name"))

	// app.version 标记为 db:false
	assert.False(t, loader.AllowDatabaseStorage("app.version"))

	// database.password 标记为 db:false
	assert.False(t, loader.AllowDatabaseStorage("database.password"))

	// poc.enabled 标记为 db:true
	assert.True(t, loader.AllowDatabaseStorage("poc.enabled"))
}
