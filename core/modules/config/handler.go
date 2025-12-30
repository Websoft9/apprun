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
// @Summary      Get configuration item
// @Description  Query a single configuration item by key, returns value, source and dynamic flag
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        key  query  string  true  "Configuration key, e.g. app.name"
// @Success      200  {object}  GetConfigResponse  "Configuration retrieved successfully"
// @Failure      400  {object}  ErrorResponse      "Missing key parameter"
// @Failure      404  {object}  ErrorResponse      "Configuration not found"
// @Router       /config [get]
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

	// isDynamic means the config CAN be modified via API (has db:"true" tag)
	// regardless of its current source
	isDynamic := h.service.loader.AllowDatabaseStorage(key)

	resp := GetConfigResponse{
		Key:       key,
		Value:     value,
		IsDynamic: isDynamic,
		Source:    source,
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// UpdateConfig 更新动态配置项
// @Summary      Update configuration item
// @Description  Update a single dynamic configuration item (only for db:true configs).
// @Description  Static configurations (db:false) cannot be updated via API.
// @Description  Changes are persisted to database and take effect immediately.
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        request  body  UpdateConfigRequest  true  "Configuration update request"  example({"key":"poc.enabled","value":"true"})
// @Success      200  {object}  UpdateConfigResponse  "Configuration updated successfully"
// @Failure      400  {object}  ErrorResponse         "Invalid request or config not allowed to store in database"
// @Router       /config [put]
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
// @Summary      List dynamic configurations
// @Description  Returns all dynamic configuration items stored in database.
// @Description  This does not include static configurations from files.
// @Description  Use this to see which configs have been overridden dynamically.
// @Tags         config
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListConfigsResponse  "Configuration list"
// @Failure      500  {object}  ErrorResponse        "Internal server error"
// @Router       /config/list [get]
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
// @Summary      Delete configuration item
// @Description  Delete a dynamic configuration item from database (config will fallback to file or default value).
// @Description  Only dynamic configurations (db:true) can be deleted.
// @Description  After deletion, the config will use the value from config files or built-in defaults.
// @Tags         config
// @Accept       json
// @Produce      json
// @Param        key  query  string  true  "Configuration key"  example(poc.enabled)
// @Success      200  {object}  map[string]interface{}  "Deletion successful"
// @Failure      400  {object}  ErrorResponse           "Missing key parameter or deletion failed"
// @Router       /config [delete]
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
// @Summary      Get allowed configuration keys
// @Description  Returns all configuration keys marked as db:true (can be modified dynamically via API).
// @Description  Use this endpoint to discover which configs can be updated through the API.
// @Description  Configs not in this list cannot be modified dynamically.
// @Tags         config
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "List of allowed configuration keys"
// @Router       /config/allowed [get]
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
