package config

import (
	"apprun/pkg/response"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_GetConfig 测试查询单个配置项
func TestHandler_GetConfig(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mockProvider := &mockConfigProvider{configs: make(map[string]string)}
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	// 预设配置数据
	mockProvider.configs["app.name"] = "test-app"

	// 创建 HTTP 请求
	req := httptest.NewRequest(http.MethodGet, "/api/config?key=app.name", nil)
	w := httptest.NewRecorder()

	// Act
	handler.GetConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp response.Response
	err = json.NewDecoder(w.Body).Decode(&apiResp)
	require.NoError(t, err)
	assert.True(t, apiResp.Success)

	// Parse the GetConfigResponse from Data
	dataBytes, err := json.Marshal(apiResp.Data)
	require.NoError(t, err)
	var configResp GetConfigResponse
	err = json.Unmarshal(dataBytes, &configResp)
	require.NoError(t, err)
	assert.Equal(t, "app.name", configResp.Key)
	assert.Equal(t, "test-app", configResp.Value)
}

// TestHandler_GetConfig_MissingKey 测试缺少 key 参数
func TestHandler_GetConfig_MissingKey(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	loader, _ := NewLoader(tmpDir, nil)
	service := NewService(loader, nil)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.GetConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code) // 422 for validation errors

	var resp response.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error.Message, "missing 'key'")
}

// TestHandler_UpdateConfig 测试更新配置
func TestHandler_UpdateConfig(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// 创建 default.yaml，提供必填配置
	defaultYAML := `
app:
  name: "apprun"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "test-password-12345"
  dbname: "apprun"
poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/apprun_poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	mockProvider := &mockConfigProvider{configs: make(map[string]string)}
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	// 创建更新请求
	reqBody := UpdateConfigRequest{
		Key:   "app.name",
		Value: "updated-app",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.UpdateConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp response.Response
	err = json.NewDecoder(w.Body).Decode(&apiResp)
	require.NoError(t, err)
	assert.True(t, apiResp.Success)

	// Extract the actual UpdateConfigResponse from Data
	respData, ok := apiResp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "app.name", respData["key"])
	assert.Equal(t, "updated-app", respData["value"])

	// 验证数据已保存
	assert.Equal(t, "updated-app", mockProvider.configs["app.name"])
}

// TestHandler_UpdateConfig_DBFalse 测试更新 db:false 配置（应拒绝）
func TestHandler_UpdateConfig_DBFalse(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// 创建 default.yaml，提供必填配置
	defaultYAML := `
app:
  name: "apprun"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "test-password-12345"
  dbname: "apprun"
poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/apprun_poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	mockProvider := &mockConfigProvider{configs: make(map[string]string)}
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	// 尝试更新 db:false 配置
	reqBody := UpdateConfigRequest{
		Key:   "app.version", // app.version 的 db tag 是 false
		Value: "2.0.0",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.UpdateConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	t.Logf("Error response: %+v", resp)
	// 检查错误消息包含 "not allowed" 或 "db:false"
	errorMsg := ""
	if resp.Error != nil {
		errorMsg = resp.Error.Message
	}
	assert.True(t,
		strings.Contains(errorMsg, "not allowed") ||
			strings.Contains(errorMsg, "db:false"),
		"Expected error message to contain 'not allowed' or 'db:false', got: %s", errorMsg)
}

// TestHandler_UpdateConfig_InvalidJSON 测试无效的 JSON 请求体
func TestHandler_UpdateConfig_InvalidJSON(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	loader, _ := NewLoader(tmpDir, nil)
	service := NewService(loader, nil)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.UpdateConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error.Message, "invalid request body")
}

// TestHandler_ListConfigs 测试列出所有动态配置
func TestHandler_ListConfigs(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mockProvider := &mockConfigProvider{configs: map[string]string{
		"app.name":    "test-app",
		"poc.enabled": "true",
	}}
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/config/list", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ListConfigs(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp response.Response
	err = json.NewDecoder(w.Body).Decode(&apiResp)
	require.NoError(t, err)
	assert.True(t, apiResp.Success)

	// Parse the ListConfigsResponse from Data
	dataBytes, err := json.Marshal(apiResp.Data)
	require.NoError(t, err)
	var listResp ListConfigsResponse
	err = json.Unmarshal(dataBytes, &listResp)
	require.NoError(t, err)
	assert.Equal(t, 2, listResp.Count)
	assert.Len(t, listResp.Configs, 2)
}

// TestHandler_DeleteConfig 测试删除动态配置
func TestHandler_DeleteConfig(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mockProvider := &mockConfigProvider{configs: map[string]string{
		"app.name": "test-app",
	}}
	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodDelete, "/api/config?key=app.name", nil)
	w := httptest.NewRecorder()

	// Act
	handler.DeleteConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp response.Response
	err = json.NewDecoder(w.Body).Decode(&apiResp)
	require.NoError(t, err)
	assert.True(t, apiResp.Success)

	// Parse the map from Data
	dataMap, ok := apiResp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "app.name", dataMap["key"])

	// 验证已删除
	_, exists := mockProvider.configs["app.name"]
	assert.False(t, exists)
}

// TestHandler_GetAllowedKeys 测试获取允许动态配置的键
func TestHandler_GetAllowedKeys(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	loader, err := NewLoader(tmpDir, nil)
	require.NoError(t, err)

	service := NewService(loader, nil)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/config/allowed", nil)
	w := httptest.NewRecorder()

	// Act
	handler.GetAllowedKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp response.Response
	err = json.NewDecoder(w.Body).Decode(&apiResp)
	require.NoError(t, err)
	assert.True(t, apiResp.Success)

	// Parse the map from Data
	dataMap, ok := apiResp.Data.(map[string]interface{})
	require.True(t, ok)
	allowedKeys, ok := dataMap["allowed_keys"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(allowedKeys), 0)
	assert.Contains(t, allowedKeys, "app.name") // app.name 的 db tag 是 true
}

// TestHandler_IntegrationFlow 集成测试：完整的 CRUD 流程
func TestHandler_IntegrationFlow(t *testing.T) {
	// Arrange - 创建完整的路由和 handler
	tmpDir := t.TempDir()

	// 创建基本的 default.yaml，提供必填配置项
	defaultYAML := `
app:
  name: "apprun"
  version: "1.0.0"
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "test-password-12345"
  dbname: "apprun"
poc:
  enabled: true
  database: "postgres://user:pass@localhost:5432/apprun_poc"
  apikey: "test-api-key-12345"
`
	err := os.WriteFile(filepath.Join(tmpDir, "default.yaml"), []byte(defaultYAML), 0644)
	require.NoError(t, err)

	// 提供必填配置项以通过验证
	mockProvider := &mockConfigProvider{
		configs: make(map[string]string),
	}

	loader, err := NewLoader(tmpDir, mockProvider)
	require.NoError(t, err)

	service := NewService(loader, mockProvider)
	handler := NewHandler(service)

	// 创建路由
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	testKey := "app.name"
	testValue := "test-app-name" // Test 1: 创建配置
	t.Run("Create Config", func(t *testing.T) {
		reqBody := UpdateConfigRequest{Key: testKey, Value: testValue}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/config", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("Error response: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test 2: 查询配置
	t.Run("Get Config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/config?key="+testKey, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var apiResp response.Response
		err := json.NewDecoder(w.Body).Decode(&apiResp)
		require.NoError(t, err)
		assert.True(t, apiResp.Success)

		// Parse GetConfigResponse from Data
		dataBytes, err := json.Marshal(apiResp.Data)
		require.NoError(t, err)
		var configResp GetConfigResponse
		err = json.Unmarshal(dataBytes, &configResp)
		require.NoError(t, err)
		assert.Equal(t, testValue, configResp.Value)
	})

	// Test 3: 列出所有配置
	t.Run("List Configs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/config/list", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var apiResp response.Response
		err := json.NewDecoder(w.Body).Decode(&apiResp)
		require.NoError(t, err)
		assert.True(t, apiResp.Success)

		// Parse ListConfigsResponse from Data
		dataBytes, err := json.Marshal(apiResp.Data)
		require.NoError(t, err)
		var listResp ListConfigsResponse
		err = json.Unmarshal(dataBytes, &listResp)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, listResp.Count, 1)
	})

	// Test 4: 删除配置
	t.Run("Delete Config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/config?key="+testKey, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test 5: 验证已删除（应返回 Not Found 或默认值）
	t.Run("Verify Deleted", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/config?key="+testKey, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// 删除后仍可能从配置文件或默认值返回，所以不检查 404
		// 只验证不再是我们设置的值
		if w.Code == http.StatusOK {
			var response GetConfigResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			// 验证返回的是默认值，而不是我们设置的值
			assert.NotEqual(t, testValue, response.Value)
		}
	})
}
