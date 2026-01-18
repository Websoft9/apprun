package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test: JSON 序列化时是否能识别 mapstructure 标签？
func TestJSONWithMapstructureTag(t *testing.T) {
	// 场景 1: 只有 mapstructure 标签
	type ConfigOnlyMapstructure struct {
		SecretKey string `mapstructure:"secret_key"`
		PoolSize  int    `mapstructure:"pool_size"`
	}

	// 场景 2: 同时有 mapstructure 和 json 标签
	type ConfigWithBothTags struct {
		SecretKey string `mapstructure:"secret_key" json:"secret_key"`
		PoolSize  int    `mapstructure:"pool_size" json:"pool_size"`
	}

	// 场景 3: json 标签与 mapstructure 不同
	type ConfigDifferentTags struct {
		SecretKey string `mapstructure:"secret_key" json:"secretKey"` // JSON 用驼峰
		PoolSize  int    `mapstructure:"pool_size" json:"poolSize"`   // JSON 用驼峰
	}

	t.Run("只有mapstructure标签-JSON序列化", func(t *testing.T) {
		cfg := ConfigOnlyMapstructure{
			SecretKey: "my-secret",
			PoolSize:  10,
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		// ❌ 期望: {"secret_key":"my-secret","pool_size":10}
		// ✅ 实际: {"SecretKey":"my-secret","PoolSize":10}
		// 结论: encoding/json 不认 mapstructure 标签！
		t.Logf("只有mapstructure标签时JSON输出: %s", string(jsonBytes))

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		// 验证：JSON 使用 Go 默认规则（首字母大写的字段名）
		assert.Contains(t, result, "SecretKey")     // ✅ Go 默认
		assert.NotContains(t, result, "secret_key") // ❌ mapstructure 标签被忽略
	})

	t.Run("同时有两个标签-JSON序列化", func(t *testing.T) {
		cfg := ConfigWithBothTags{
			SecretKey: "my-secret",
			PoolSize:  10,
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		// ✅ 期望并实际: {"secret_key":"my-secret","pool_size":10}
		t.Logf("有json标签时JSON输出: %s", string(jsonBytes))

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		// 验证：JSON 使用 json 标签的值
		assert.Contains(t, result, "secret_key")   // ✅ 使用 json 标签
		assert.NotContains(t, result, "SecretKey") // ❌ 不使用 Go 默认
	})

	t.Run("json标签与mapstructure不同", func(t *testing.T) {
		cfg := ConfigDifferentTags{
			SecretKey: "my-secret",
			PoolSize:  10,
		}

		jsonBytes, err := json.Marshal(cfg)
		assert.NoError(t, err)

		// ✅ 输出: {"secretKey":"my-secret","poolSize":10}
		// 说明：json 和 mapstructure 可以完全独立！
		t.Logf("不同标签时JSON输出: %s", string(jsonBytes))

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "secretKey")     // ✅ 使用 json 标签（驼峰）
		assert.NotContains(t, result, "secret_key") // ❌ 不使用 mapstructure 标签（下划线）
	})

	t.Run("反序列化-只有mapstructure标签", func(t *testing.T) {
		// 尝试从 JSON 反序列化（使用下划线命名）
		jsonStr := `{"secret_key":"from-json","pool_size":20}`

		var cfg ConfigOnlyMapstructure
		err := json.Unmarshal([]byte(jsonStr), &cfg)
		assert.NoError(t, err)

		// ❌ 预期会填充，但实际字段为空！
		// 因为 encoding/json 不认 mapstructure 标签
		t.Logf("反序列化结果: SecretKey=%s, PoolSize=%d", cfg.SecretKey, cfg.PoolSize)
		assert.Empty(t, cfg.SecretKey)   // ❌ 无法匹配 secret_key → SecretKey
		assert.Equal(t, 0, cfg.PoolSize) // ❌ 无法匹配 pool_size → PoolSize
	})

	t.Run("反序列化-有json标签", func(t *testing.T) {
		jsonStr := `{"secret_key":"from-json","pool_size":20}`

		var cfg ConfigWithBothTags
		err := json.Unmarshal([]byte(jsonStr), &cfg)
		assert.NoError(t, err)

		// ✅ 成功匹配
		assert.Equal(t, "from-json", cfg.SecretKey)
		assert.Equal(t, 20, cfg.PoolSize)
	})
}

// Test: 实际场景模拟 - 配置中心 API
func TestConfigCenterAPIScenario(t *testing.T) {
	type CacheConfig struct {
		Host     string `mapstructure:"host"` // 只有 mapstructure
		PoolSize int    `mapstructure:"pool_size"`
	}

	type JWTConfig struct {
		Secret string `mapstructure:"secret" json:"secret"` // 两个都有
		Expiry string `mapstructure:"expiry" json:"expiry"`
	}

	t.Run("配置中心API返回cache配置-问题演示", func(t *testing.T) {
		// 假设配置中心返回 cache 配置
		cfg := CacheConfig{Host: "redis.prod", PoolSize: 50}

		// API Handler: c.JSON(200, cfg)
		jsonBytes, _ := json.Marshal(cfg)

		// ❌ 问题：前端收到的字段名是 Go 风格而不是配置文件风格
		// 输出: {"Host":"redis.prod","PoolSize":50}
		// 期望: {"host":"redis.prod","pool_size":50}
		t.Logf("API 返回给前端: %s", string(jsonBytes))

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "Host")    // ❌ Go 风格
		assert.NotContains(t, result, "host") // ❌ 不是 YAML 风格
	})

	t.Run("配置中心API返回JWT配置-正确示例", func(t *testing.T) {
		cfg := JWTConfig{Secret: "secret-key", Expiry: "24h"}

		jsonBytes, _ := json.Marshal(cfg)

		// ✅ 正确：字段名与 YAML 配置一致
		// 输出: {"secret":"secret-key","expiry":"24h"}
		t.Logf("API 返回给前端: %s", string(jsonBytes))

		var result map[string]interface{}
		json.Unmarshal(jsonBytes, &result)

		assert.Contains(t, result, "secret") // ✅ 与 YAML 一致
		assert.Contains(t, result, "expiry")
	})
}
