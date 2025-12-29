package config

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler 配置管理 HTTP 处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器实例
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes 注册路由到 chi.Router
// 注意：此方法应在 /api 路由组内调用，会注册 /config 子路由
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/config", func(r chi.Router) {
		r.Get("/", h.GetConfig)             // GET /api/config?key=xxx
		r.Put("/", h.UpdateConfig)          // PUT /api/config
		r.Get("/list", h.ListConfigs)       // GET /api/config/list
		r.Delete("/", h.DeleteConfig)       // DELETE /api/config?key=xxx
		r.Get("/allowed", h.GetAllowedKeys) // GET /api/config/allowed
	})
}

// GetConfig 获取配置值（查询单个配置项）
// GET /api/config?key=app.name
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		h.respondError(w, http.StatusBadRequest, "missing 'key' query parameter", "")
		return
	}

	value, source, err := h.service.GetConfigValue(r.Context(), key)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "config not found", err.Error())
		return
	}

	// 判断是否为动态配置
	isDynamic := h.service.loader.AllowDatabaseStorage(key) && source == "database"

	resp := GetConfigResponse{
		Key:       key,
		Value:     value,
		IsDynamic: isDynamic,
		Source:    source,
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// UpdateConfig 更新动态配置项
// PUT /api/config
// Body: {"key": "poc.enabled", "value": "false"}
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// 验证请求
	if req.Key == "" {
		h.respondError(w, http.StatusBadRequest, "missing 'key' field", "")
		return
	}
	if req.Value == "" {
		h.respondError(w, http.StatusBadRequest, "missing 'value' field", "")
		return
	}

	// 更新配置
	if err := h.service.UpdateConfig(r.Context(), req.Key, req.Value); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to update config", err.Error())
		return
	}

	resp := UpdateConfigResponse{
		Success: true,
		Message: "config updated successfully",
		Key:     req.Key,
		Value:   req.Value,
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// ListConfigs 列出所有动态配置项
// GET /api/config/list
func (h *Handler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.service.ListDynamicConfigs(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list configs", err.Error())
		return
	}

	resp := ListConfigsResponse{
		Configs: configs,
		Count:   len(configs),
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// DeleteConfig 删除动态配置项
// DELETE /api/config?key=poc.enabled
func (h *Handler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		h.respondError(w, http.StatusBadRequest, "missing 'key' query parameter", "")
		return
	}

	if err := h.service.DeleteDynamicConfig(r.Context(), key); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to delete config", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "config deleted successfully",
		"key":     key,
	})
}

// GetAllowedKeys 获取所有允许动态配置的键（db:true）
// GET /api/config/allowed
func (h *Handler) GetAllowedKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.service.GetAllowedDynamicKeys()

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"allowed_keys": keys,
		"count":        len(keys),
	})
}

// respondJSON 发送 JSON 响应
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError 发送错误响应
func (h *Handler) respondError(w http.ResponseWriter, status int, message string, details string) {
	resp := ErrorResponse{
		Error:   message,
		Details: details,
	}
	h.respondJSON(w, status, resp)
}
