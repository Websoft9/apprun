package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBugFix_DatabasePassword 验证 Bug #1: database.password 应该从文件中读取
func TestBugFix_DatabasePassword(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建包含 database.password 的配置文件
	defaultYAML := `
app:
  name: "test-app"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "test-password-from-file-12345"
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

	// 测试：获取 database.password 应该成功（从文件读取）
	value, source, err := service.GetConfigValue(ctx, "database.password")
	require.NoError(t, err, "Bug #1: database.password should be readable from file")
	assert.Equal(t, "test-password-from-file-12345", value)
	assert.Equal(t, "file", source)
}

// TestBugFix_IsDynamic 验证 Bug #2: is_dynamic 应该只基于 db tag
func TestBugFix_IsDynamic(t *testing.T) {
	tmpDir := t.TempDir()

	defaultYAML := `
app:
  name: "my-app-from-file"
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
	cfg, err := service.LoadConfig(ctx)
	require.NoError(t, err)

	// 验证配置已加载
	assert.Equal(t, "my-app-from-file", cfg.App.Name)

	// 测试：app.name 有 db:true tag，即使值来自文件，isDynamic也应该为true
	// 通过 AllowDatabaseStorage 检查
	isDynamic := loader.AllowDatabaseStorage("app.name")
	assert.True(t, isDynamic, "Bug #2: app.name with db:true should show isDynamic=true")

	// 测试：app.version 有 db:false tag，isDynamic应该为false
	isDynamicVersion := loader.AllowDatabaseStorage("app.version")
	assert.False(t, isDynamicVersion, "app.version with db:false should show isDynamic=false")

	// 验证通过 GetConfigValue 获取的 source
	value, source, err := service.GetConfigValue(ctx, "app.name")
	require.NoError(t, err)
	assert.Equal(t, "my-app-from-file", value)
	assert.Equal(t, "file", source) // 值来自文件

	// 但 isDynamic 标记应该是 true（因为有 db:true）
	// 这个逻辑在 handler 中，我们通过 AllowDatabaseStorage 验证
	assert.True(t, loader.AllowDatabaseStorage("app.name"))
}

// TestBugFix_ConfigPriority 验证删除数据库配置后正确回退到文件值
func TestBugFix_ConfigPriority(t *testing.T) {
	tmpDir := t.TempDir()

	defaultYAML := `
app:
  name: "original-name-from-file"
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

	// Step 1: 验证初始值来自文件
	value, source, err := service.GetConfigValue(ctx, "app.name")
	require.NoError(t, err)
	assert.Equal(t, "original-name-from-file", value)
	assert.Equal(t, "file", source)

	// Step 2: 更新为数据库值
	err = service.UpdateConfig(ctx, "app.name", "updated-name-in-database")
	require.NoError(t, err)

	value, source, err = service.GetConfigValue(ctx, "app.name")
	require.NoError(t, err)
	assert.Equal(t, "updated-name-in-database", value)
	assert.Equal(t, "database", source)

	// Step 3: 删除数据库值，应该回退到文件值
	err = service.DeleteDynamicConfig(ctx, "app.name")
	require.NoError(t, err)

	value, source, err = service.GetConfigValue(ctx, "app.name")
	require.NoError(t, err)
	assert.Equal(t, "original-name-from-file", value, "After delete, should fallback to file value")
	assert.Equal(t, "file", source)
}
