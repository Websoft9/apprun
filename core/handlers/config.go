package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"apprun/internal/config"

	"github.com/go-chi/chi/v5"
)

// ConfigHandler 处理配置相关的HTTP请求
type ConfigHandler struct {
	// 可以添加依赖注入，如数据库client等
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// GetConfig 处理 GET /config 请求
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	items, err := config.GetAllConfigItems()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode config items: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateConfig 处理 PUT /config 请求
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(updates) == 0 {
		http.Error(w, "No updates provided", http.StatusBadRequest)
		return
	}

	// 更新配置
	if err := config.UpdateConfig(updates); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "config key not allowed to be modified" {
			status = http.StatusForbidden
		} else if err.Error() == "config validation failed after update" {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	// 返回更新后的配置
	items, err := config.GetAllConfigItems()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode updated config: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetConfigItem 处理 GET /config/{key} 请求
func (h *ConfigHandler) GetConfigItem(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "Config key is required", http.StatusBadRequest)
		return
	}

	// 获取所有配置项并查找指定key
	items, err := config.GetAllConfigItems()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusInternalServerError)
		return
	}

	// 查找指定的配置项
	for _, item := range items {
		if item.Path == key {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			return
		}
	}

	http.Error(w, "Config key not found", http.StatusNotFound)
}

// HealthCheck 处理健康检查请求
func (h *ConfigHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "apprun",
	})
}
