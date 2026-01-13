package config

import (
	"encoding/json"
	"testing"

	"apprun/modules/auth"
	"apprun/pkg/cache"
	"apprun/pkg/database"
	"apprun/pkg/i18n"
	"apprun/pkg/jwt"
	"apprun/pkg/logger"
	"apprun/pkg/server"

	"github.com/stretchr/testify/assert"
)

// TestAllConfigsHaveJSONTags 验证所有配置结构体都有 json 标签
// 确保 API 返回的 JSON 字段名与 YAML 配置键名一致
func TestAllConfigsHaveJSONTags(t *testing.T) {
	t.Run("cache.Config序列化验证", func(t *testing.T) {
		cfg := cache.Config{
			Host:     "redis.prod",
			Port:     "6379",
			PoolSize: 50,
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		// ✅ 验证使用 json 标签（snake_case），不是 Go 默认（PascalCase）
		assert.Contains(t, result, "host", "应使用 json 标签值 'host'")
		assert.Contains(t, result, "port", "应使用 json 标签值 'port'")
		assert.Contains(t, result, "pool_size", "应使用 json 标签值 'pool_size'")

		assert.NotContains(t, result, "Host", "不应使用 Go 默认字段名")
		assert.NotContains(t, result, "PoolSize", "不应使用 Go 默认字段名")

		t.Logf("✅ cache.Config JSON: %s", string(jsonBytes))
	})

	t.Run("database.Config序列化验证", func(t *testing.T) {
		cfg := database.Config{
			Driver: "postgres",
			Host:   "db.prod",
			Port:   5432,
			DBName: "apprun",
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "driver")
		assert.Contains(t, result, "host")
		assert.Contains(t, result, "port")
		assert.Contains(t, result, "db_name", "应使用 snake_case")

		assert.NotContains(t, result, "DBName")

		t.Logf("✅ database.Config JSON: %s", string(jsonBytes))
	})

	t.Run("logger.Config序列化验证", func(t *testing.T) {
		cfg := logger.Config{
			Level: logger.LevelInfo,
			Output: logger.OutputConfig{
				Targets: []string{"stdout"},
			},
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "level")
		assert.Contains(t, result, "output")

		assert.NotContains(t, result, "Level")
		assert.NotContains(t, result, "Output")

		t.Logf("✅ logger.Config JSON: %s", string(jsonBytes))
	})

	t.Run("server.Config序列化验证", func(t *testing.T) {
		cfg := server.Config{
			HTTPPort:  "8080",
			HTTPSPort: "8443",
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "http_port", "应使用 snake_case")
		assert.Contains(t, result, "https_port", "应使用 snake_case")

		assert.NotContains(t, result, "HTTPPort")
		assert.NotContains(t, result, "HTTPSPort")

		t.Logf("✅ server.Config JSON: %s", string(jsonBytes))
	})

	t.Run("jwt.Config序列化验证", func(t *testing.T) {
		cfg := jwt.Config{
			Secret: "test-secret-key-must-be-32-chars",
			Issuer: "apprun",
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "secret")
		assert.Contains(t, result, "issuer")

		assert.NotContains(t, result, "Secret")
		assert.NotContains(t, result, "Issuer")

		t.Logf("✅ jwt.Config JSON: %s", string(jsonBytes))
	})

	t.Run("i18n.Config序列化验证", func(t *testing.T) {
		cfg := i18n.Config{
			DefaultLanguage:    "en-US",
			SupportedLanguages: []string{"en-US", "zh-CN"},
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "default_language", "应使用 snake_case")
		assert.Contains(t, result, "supported_languages", "应使用 snake_case")

		assert.NotContains(t, result, "DefaultLanguage")
		assert.NotContains(t, result, "SupportedLanguages")

		t.Logf("✅ i18n.Config JSON: %s", string(jsonBytes))
	})

	t.Run("auth.Config序列化验证", func(t *testing.T) {
		cfg := auth.Config{
			JWT: jwt.Config{
				Secret: "test-secret-key-must-be-32-chars",
			},
			Security: auth.SecurityConfig{
				BcryptCost:        10,
				MaxFailedAttempts: 5,
			},
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "jwt")
		assert.Contains(t, result, "security")

		// 验证嵌套结构
		security := result["security"].(map[string]interface{})
		assert.Contains(t, security, "bcrypt_cost", "应使用 snake_case")
		assert.Contains(t, security, "max_failed_attempts", "应使用 snake_case")

		assert.NotContains(t, result, "JWT")
		assert.NotContains(t, result, "Security")

		t.Logf("✅ auth.Config JSON: %s", string(jsonBytes))
	})
}

// TestJSONTagsMatchMapstructureTags 验证 json 标签值与 mapstructure 标签值一致
// 确保配置加载和 API 响应使用相同的字段名
func TestJSONTagsMatchMapstructureTags(t *testing.T) {
	t.Run("字段名一致性验证", func(t *testing.T) {
		// 使用 cache.Config 作为代表性测试
		cfg := cache.Config{
			Host:       "localhost",
			Port:       "6379",
			PoolSize:   10,
			MaxRetries: 3,
		}

		// 序列化为 JSON
		jsonBytes, _ := json.Marshal(cfg)
		var jsonResult map[string]interface{}
		json.Unmarshal(jsonBytes, &jsonResult)

		// 验证：JSON 字段名应该与 YAML 配置键名相同
		// YAML: cache.pool_size → JSON API: {"pool_size": 10}
		expectedFields := []string{
			"host",        // 与 YAML key 一致
			"port",        // 与 YAML key 一致
			"pool_size",   // 与 YAML key 一致（snake_case）
			"max_retries", // 与 YAML key 一致（snake_case）
		}

		for _, field := range expectedFields {
			assert.Contains(t, jsonResult, field,
				"JSON 字段名应与 YAML 配置键名一致，确保前后端一致性")
		}

		t.Logf("✅ 字段名一致性验证通过：JSON API 与 YAML 配置使用相同命名")
	})
}
