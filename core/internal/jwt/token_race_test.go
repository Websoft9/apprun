package jwt

import (
	"sync"
	"testing"
	"time"
)

// TestTokenService_ConcurrentAccess tests that concurrent access to token generation is safe
func TestTokenService_ConcurrentAccess(t *testing.T) {
	// Reset global state for testing
	globalTokenService = nil
	once = sync.Once{}

	const numGoroutines = 100
	const numIterations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a mock config provider
	mockCfg := &mockConfigProvider{
		data: map[string]interface{}{
			"jwt.secret":                   "test-secret-key-32-characters!!",
			"jwt.access_token_expiration":  "24h",
			"jwt.refresh_token_expiration": "168h",
		},
	}

	// Initialize once before test
	InitGlobalService(mockCfg)

	// Run concurrent token generation
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				userClaims := map[string]interface{}{
					"worker_id": workerID,
					"iteration": j,
				}
				_, _, _, err := GenerateTokenPair(int64(workerID), userClaims)
				if err != nil {
					t.Errorf("Worker %d iteration %d failed: %v", workerID, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

// mockConfigProvider implements a simple config provider for testing
type mockConfigProvider struct {
	data map[string]interface{}
}

func (m *mockConfigProvider) Get(key string) interface{} {
	return m.data[key]
}

func (m *mockConfigProvider) GetString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

func (m *mockConfigProvider) GetInt(key string) int {
	if v, ok := m.data[key].(int); ok {
		return v
	}
	return 0
}

func (m *mockConfigProvider) GetInt64(key string) int64 {
	if v, ok := m.data[key].(int64); ok {
		return v
	}
	return 0
}

func (m *mockConfigProvider) GetDuration(key string) time.Duration {
	if v, ok := m.data[key].(string); ok {
		d, _ := time.ParseDuration(v)
		return d
	}
	if v, ok := m.data[key].(time.Duration); ok {
		return v
	}
	return 0
}

func (m *mockConfigProvider) GetBool(key string) bool {
	if v, ok := m.data[key].(bool); ok {
		return v
	}
	return false
}

func (m *mockConfigProvider) GetFloat64(key string) float64 {
	if v, ok := m.data[key].(float64); ok {
		return v
	}
	return 0.0
}

func (m *mockConfigProvider) GetStringSlice(key string) []string {
	if v, ok := m.data[key].([]string); ok {
		return v
	}
	return nil
}

func (m *mockConfigProvider) IsSet(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *mockConfigProvider) AllSettings() map[string]interface{} {
	return m.data
}

func (m *mockConfigProvider) Set(key string, value interface{}) {
	m.data[key] = value
}
