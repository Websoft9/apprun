package config

import (
	"context"
)

// mockConfigProvider 模拟配置提供者（测试辅助，共享给所有测试文件）
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
