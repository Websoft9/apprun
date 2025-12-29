package config

import (
	"context"
)

// ConfigProvider 定义配置持久化接口
// 实现反腐层模式，隔离 Ent 模型
type ConfigProvider interface {
	// GetConfig 根据 key 获取配置项
	GetConfig(ctx context.Context, key string) (value string, isDynamic bool, err error)

	// SetConfig 设置动态配置项（只允许 db:true 的项）
	SetConfig(ctx context.Context, key string, value string) error

	// ListDynamicConfigs 列出所有动态配置项
	ListDynamicConfigs(ctx context.Context) (map[string]string, error)

	// DeleteConfig 删除动态配置项
	DeleteConfig(ctx context.Context, key string) error
}

// GetConfigRequest GET /api/config 请求参数
type GetConfigRequest struct {
	Key string `json:"key" validate:"required"` // 配置键，如 "app.name" 或 "poc.enabled"
}

// GetConfigResponse GET /api/config 响应
type GetConfigResponse struct {
	Key       string `json:"key" example:"app.name"`     // Configuration key
	Value     string `json:"value" example:"apprun"`     // Configuration value
	IsDynamic bool   `json:"is_dynamic" example:"false"` // Whether it's a dynamic configuration
	Source    string `json:"source" example:"default"`   // Source: "database", "file", "env", "default"
}

// UpdateConfigRequest PUT /api/config 请求体
type UpdateConfigRequest struct {
	Key   string `json:"key" validate:"required" example:"poc.enabled"` // Configuration key
	Value string `json:"value" validate:"required" example:"true"`      // New value
}

// UpdateConfigResponse PUT /api/config 响应
type UpdateConfigResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message,omitempty" example:"config updated successfully"`
	Key     string `json:"key" example:"poc.enabled"`
	Value   string `json:"value" example:"true"`
}

// ListConfigsResponse GET /api/configs 响应（列出所有动态配置）
type ListConfigsResponse struct {
	Configs map[string]string `json:"configs"`           // Key-value mapping of dynamic configurations
	Count   int               `json:"count" example:"3"` // Number of configuration items
}

// ErrorResponse 通用错误响应
type ErrorResponse struct {
	Error   string `json:"error" example:"missing 'key' query parameter"`
	Details string `json:"details,omitempty" example:""`
}
