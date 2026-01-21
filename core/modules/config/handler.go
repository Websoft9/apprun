// Package config provides HTTP handlers for configuration management endpoints.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"apprun/ent"
	jwtpkg "apprun/internal/jwt"
	apperrors "apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// Handler 配置管理 HTTP 处理器
type Handler struct {
	service  *Service
	dbClient *ent.Client
}

// NewHandler 创建处理器实例
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		dbClient: service.provider.(*Repository).client,
	}
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
//
//	@Summary		Get configuration item
//	@Description	Query a single configuration item by key, returns value, source and dynamic flag
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string				true	"Configuration key, e.g. app.name"
//	@Success		200	{object}	GetConfigResponse	"Configuration retrieved successfully"
//	@Failure		400	{object}	response.Response	"Missing key parameter"
//	@Failure		401	{object}	response.Response	"Unauthorized - missing or invalid JWT token"
//	@Failure		403	{object}	response.Response	"Forbidden - insufficient permissions"
//	@Failure		404	{object}	response.Response	"Configuration not found"
//	@Security		BearerAuth
//	@Router			/api/config [get]
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		response.ValidationErrorWithRequest(w, r, "key", "missing 'key' query parameter")
		return
	}

	value, source, err := h.service.GetConfigValue(r.Context(), key)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
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

	response.SuccessWithRequest(w, r, resp)
}

// UpdateConfig 更新动态配置项
//
//	@Summary		Update configuration item
//	@Description	Update a single dynamic configuration item (only for db:true configs).
//	@Description	Static configurations (db:false) cannot be updated via API.
//	@Description	Changes are persisted to database and take effect immediately.
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UpdateConfigRequest		true	"Configuration update request"	example({"key":"poc.enabled","value":"true"})
//	@Success		200		{object}	UpdateConfigResponse	"Configuration updated successfully"
//	@Failure		400		{object}	response.Response		"Invalid request or config not allowed to store in database"
//	@Failure		401		{object}	response.Response		"Unauthorized - missing or invalid JWT token"
//	@Failure		403		{object}	response.Response		"Forbidden - insufficient permissions (requires platform:config:write)"
//	@Security		BearerAuth
//	@Router			/api/config [put]
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.AppErrorWithRequest(w, r, apperrors.New(apperrors.ErrCodeConfigInvalidKey, "Invalid request body"))
		return
	}

	// 验证请求
	if req.Key == "" {
		response.ValidationErrorWithRequest(w, r, "key", "missing 'key' field")
		return
	}
	if req.Value == "" {
		response.ValidationErrorWithRequest(w, r, "value", "missing 'value' field")
		return
	}

	// 获取用户ID用于审计日志
	userID := jwtpkg.GetUserID(ctx)

	// 获取旧值用于审计日志
	oldValue, _, _ := h.service.GetConfigValue(ctx, req.Key)

	// 更新配置
	if err := h.service.UpdateConfig(ctx, req.Key, req.Value); err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// 记录审计日志 (Story 3-2-1)
	h.createAuditLog(ctx, userID, "update", req.Key, oldValue, req.Value, r.RemoteAddr, r.UserAgent())

	resp := UpdateConfigResponse(req)

	response.SuccessWithRequest(w, r, resp)
}

// ListConfigs 列出所有动态配置项
//
//	@Summary		List dynamic configurations
//	@Description	Returns all dynamic configuration items stored in database.
//	@Description	This does not include static configurations from files.
//	@Description	Use this to see which configs have been overridden dynamically.
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ListConfigsResponse	"Configuration list"
//	@Failure		401	{object}	response.Response	"Unauthorized - missing or invalid JWT token"
//	@Failure		403	{object}	response.Response	"Forbidden - insufficient permissions"
//	@Failure		500	{object}	response.Response	"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/config/list [get]
func (h *Handler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.service.ListDynamicConfigs(r.Context())
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	resp := ListConfigsResponse{
		Configs: configs,
		Count:   len(configs),
	}

	response.SuccessWithRequest(w, r, resp)
}

// DeleteConfig 删除动态配置项
//
//	@Summary		Delete configuration item
//	@Description	Delete a dynamic configuration item from database (config will fallback to file or default value).
//	@Description	Only dynamic configurations (db:true) can be deleted.
//	@Description	After deletion, the config will use the value from config files or built-in defaults.
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string					true	"Configuration key"	example(poc.enabled)
//	@Success		200	{object}	map[string]interface{}	"Deletion successful"
//	@Failure		400	{object}	response.Response		"Missing key parameter or deletion failed"
//	@Failure		401	{object}	response.Response		"Unauthorized - missing or invalid JWT token"
//	@Failure		403	{object}	response.Response		"Forbidden - insufficient permissions (requires platform:config:write)"
//	@Security		BearerAuth
//	@Router			/api/config [delete]
func (h *Handler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.URL.Query().Get("key")
	if key == "" {
		response.ValidationErrorWithRequest(w, r, "key", "missing 'key' query parameter")
		return
	}

	// 获取用户ID用于审计日志
	userID := jwtpkg.GetUserID(ctx)

	// 获取旧值用于审计日志
	oldValue, _, _ := h.service.GetConfigValue(ctx, key)

	if err := h.service.DeleteDynamicConfig(ctx, key); err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// 记录审计日志 (Story 3-2-1)
	h.createAuditLog(ctx, userID, "delete", key, oldValue, "", r.RemoteAddr, r.UserAgent())

	response.SuccessWithRequest(w, r, map[string]interface{}{
		"key": key,
	})
}

// GetAllowedKeys 获取所有允许动态配置的键（db:true）
//
//	@Summary		Get allowed configuration keys
//	@Description	Returns all configuration keys marked as db:true (can be modified dynamically via API).
//	@Description	Use this endpoint to discover which configs can be updated through the API.
//	@Description	Configs not in this list cannot be modified dynamically.
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"List of allowed configuration keys"
//	@Failure		401	{object}	response.Response	"Unauthorized - missing or invalid JWT token"
//	@Failure		403	{object}	response.Response	"Forbidden - insufficient permissions"
//	@Security		BearerAuth
//	@Router			/api/config/allowed [get]
func (h *Handler) GetAllowedKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.service.GetAllowedDynamicKeys()

	response.SuccessWithRequest(w, r, map[string]interface{}{
		"allowed_keys": keys,
		"count":        len(keys),
	})
}

// createAuditLog creates an audit log entry for configuration changes (Story 3-2-1)
// This function is non-blocking - audit log failures don't block config operations
func (h *Handler) createAuditLog(ctx context.Context, userID int64, action, key, oldValue, newValue, ipAddress, userAgent string) {
	// Build changes JSON
	changes := map[string]interface{}{
		"key": key,
	}
	if action == "update" {
		changes["old_value"] = oldValue
		changes["new_value"] = newValue
	} else if action == "delete" {
		changes["old_value"] = oldValue
	}

	// Create audit log entry using the current schema
	_, err := h.dbClient.AuditLog.Create().
		SetAction(fmt.Sprintf("config.%s", action)).
		SetTargetID(key).
		SetTargetType("config").
		SetChanges(changes).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent).
		SetStatusCode(200).
		Save(ctx)

	if err != nil {
		// Log error but don't fail the config operation
		logger.Error("Failed to create audit log for config change",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "action", Value: action},
			logger.Field{Key: "key", Value: key},
			logger.Field{Key: "error", Value: err.Error()},
		)
	}
}
