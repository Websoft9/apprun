package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestService_LoadConfig 测试配置加载和验证
func TestService_LoadConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建有效的配置文件
	validYAML := `
app:
  name: "test-service"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword123"
  dbname: "testdb"
poc:
  enabled: true
  database: "http://localhost:5432/poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(validYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	service := NewService(loader, nil)

	ctx := context.Background()
	cfg, err := service.LoadConfig(ctx)
	if err != nil {
		t.Logf("LoadConfig error: %v", err)
		if cfg != nil {
			t.Logf("POC.APIKey: '%s'", cfg.POC.APIKey)
			t.Logf("POC.Enabled: %v", cfg.POC.Enabled)
			t.Logf("POC.Database: '%s'", cfg.POC.Database)
		}
	}
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-service", cfg.App.Name)
}

// TestService_LoadConfig_ValidationFailure 测试配置验证失败
func TestService_LoadConfig_ValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建无效的配置文件（password 少于 8 个字符）
	invalidYAML := `
app:
  name: "test"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "short"
  dbname: "testdb"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	service := NewService(loader, nil)

	ctx := context.Background()
	_, err = service.LoadConfig(ctx)
	assert.Error(t, err, "should fail validation due to short password")
	assert.Contains(t, err.Error(), "validation failed")
}

// TestService_UpdateConfig 测试更新动态配置
func TestService_UpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建基础配置
	defaultYAML := `
app:
  name: "test-app"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword123"
  dbname: "testdb"
poc:
  enabled: false
  database: "http://localhost:5432/poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	mockProvider := newMockProvider()
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)

	ctx := context.Background()
	_, err = service.LoadConfig(ctx)
	require.NoError(t, err)

	// 更新动态配置（app.name 是 db:true）
	err = service.UpdateConfig(ctx, "app.name", "updated-app-name")
	require.NoError(t, err)

	// 验证更新后的值
	value, source, err := service.GetConfigValue(ctx, "app.name")
	require.NoError(t, err)
	assert.Equal(t, "updated-app-name", value)
	assert.Equal(t, "database", source)
}

// TestService_UpdateConfig_DBFalse 测试更新 db:false 的配置项应失败
func TestService_UpdateConfig_DBFalse(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建基础配置
	defaultYAML := `
app:
  name: "test-app"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword123"
  dbname: "testdb"
poc:
  enabled: false
  database: "http://localhost:5432/poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	mockProvider := newMockProvider()
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)

	ctx := context.Background()
	_, err = service.LoadConfig(ctx)
	require.NoError(t, err)

	// 尝试更新 db:false 的配置项（app.version）
	err = service.UpdateConfig(ctx, "app.version", "2.0.0")
	assert.Error(t, err, "should fail to update db:false config")
	assert.Contains(t, err.Error(), "not allowed to be stored in database")
}

// TestService_DeleteDynamicConfig 测试删除动态配置
func TestService_DeleteDynamicConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建基础配置
	defaultYAML := `
app:
  name: "test-app"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpassword123"
  dbname: "testdb"
poc:
  enabled: false
  database: "http://localhost:5432/poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	mockProvider := newMockProvider()
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)

	ctx := context.Background()
	_, err = service.LoadConfig(ctx)
	require.NoError(t, err)

	// 先设置动态配置
	err = service.UpdateConfig(ctx, "poc.enabled", "true")
	require.NoError(t, err)

	// 删除动态配置
	err = service.DeleteDynamicConfig(ctx, "poc.enabled")
	require.NoError(t, err)

	// 验证已删除（应回退到默认值）
	value, source, err := service.GetConfigValue(ctx, "poc.enabled")
	require.NoError(t, err)
	assert.NotEqual(t, "true", value) // 应该不再是数据库中的 "true"
	assert.NotEqual(t, "database", source)
}

// TestService_GetAllowedDynamicKeys 测试获取允许动态配置的键
func TestService_GetAllowedDynamicKeys(t *testing.T) {
	tmpDir := t.TempDir()

	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	service := NewService(loader, nil)

	keys := service.GetAllowedDynamicKeys()
	assert.NotEmpty(t, keys)

	// 验证包含已知的 db:true 键
	assert.Contains(t, keys, "app.name")
	assert.Contains(t, keys, "poc.enabled")
	assert.Contains(t, keys, "poc.database")
	assert.Contains(t, keys, "poc.apikey")

	// 验证不包含 db:false 键
	assert.NotContains(t, keys, "app.version")
	assert.NotContains(t, keys, "database.password")
}
