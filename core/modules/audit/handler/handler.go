// Package handler provides HTTP handlers for audit log query operations.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"apprun/ent"
	"apprun/ent/user"
	"apprun/modules/audit/service"
	"apprun/modules/audit/storage"
	"apprun/pkg/response"

	"github.com/google/uuid"
)

type Handler struct {
	service  *service.Service
	dbClient *ent.Client
}

func New(svc *service.Service, dbClient *ent.Client) *Handler {
	return &Handler{
		service:  svc,
		dbClient: dbClient,
	}
}

// QueryLogs godoc
// @Summary Query audit logs
// @Description Query system audit logs with filtering and pagination. Requires platform_admin role. Returns audit logs with operator information including email address. All query parameters are optional filters.
// @Tags audit
// @Accept json
// @Produce json
// @Param start_time query string false "Start time filter in RFC3339 format" example(2026-01-01T00:00:00Z)
// @Param end_time query string false "End time filter in RFC3339 format" example(2026-01-20T23:59:59Z)
// @Param operator_id query string false "Filter by operator UUID" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Param action query string false "Filter by action type. Available options: login, logout, token_refresh, user.create, user.update, user.delete, role.assign, access_denied" Enums(login, logout, token_refresh, user.create, user.update, user.delete, role.assign, access_denied) example(login)
// @Param target_type query string false "Filter by target resource type. Available options: user, role, config, project" Enums(user, role, config, project) example(user)
// @Param page query int false "Page number, starts from 1" minimum(1) default(1) example(1)
// @Param page_size query int false "Page size, default 50, maximum 200" minimum(1) maximum(200) default(50) example(20)
// @Success 200 {object} response.Response{data=object{logs=[]map[string]interface{},total=int64,page=int,page_size=int}} "Returns audit logs with pagination info. Each log contains: id, timestamp, operator_id, operator_email, action, target_id, target_type, changes, ip_address, user_agent, status_code, response_time_ms, method, path"
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - platform_admin role required"
// @Failure 500 {object} response.Response "Internal server error"
// @Security BearerAuth
// @Router /api/admin/audit-logs [get]
func (h *Handler) QueryLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := storage.AuditFilter{
		Page:         parseIntParam(r, "page", 1),
		PageSize:     parseIntParam(r, "page_size", 50),
		IncludeTotal: true,
	}

	if filter.PageSize > 200 {
		filter.PageSize = 200
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	if filter.Page < 1 {
		filter.Page = 1
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	if operatorIDStr := r.URL.Query().Get("operator_id"); operatorIDStr != "" {
		if operatorID, err := uuid.Parse(operatorIDStr); err == nil {
			filter.OperatorID = &operatorID
		}
	}

	filter.Action = r.URL.Query().Get("action")
	filter.TargetType = r.URL.Query().Get("target_type")

	result, err := h.service.Query(ctx, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "AUDIT_QUERY_FAILED", err.Error())
		return
	}

	// Collect operator IDs to avoid N+1 query
	operatorIDs := make([]uuid.UUID, 0)
	for _, log := range result.Logs {
		if log.OperatorID != nil {
			operatorIDs = append(operatorIDs, *log.OperatorID)
		}
	}

	// Batch query users
	userMap := make(map[uuid.UUID]string)
	if len(operatorIDs) > 0 {
		users, err := h.dbClient.User.Query().
			Where(user.UUIDIn(operatorIDs...)).
			All(ctx)
		if err == nil {
			for _, u := range users {
				userMap[u.UUID] = u.Email
			}
		}
	}

	enrichedLogs := make([]map[string]interface{}, 0, len(result.Logs))
	for _, log := range result.Logs {
		logMap := map[string]interface{}{
			"id":               log.ID,
			"timestamp":        log.Timestamp,
			"action":           log.Action,
			"target_id":        log.TargetID,
			"target_type":      log.TargetType,
			"changes":          log.Changes,
			"ip_address":       log.IPAddress,
			"user_agent":       log.UserAgent,
			"status_code":      log.StatusCode,
			"response_time_ms": log.ResponseTimeMs,
			"method":           log.Method,
			"path":             log.Path,
		}

		if log.OperatorID != nil {
			logMap["operator_id"] = log.OperatorID
			if email, ok := userMap[*log.OperatorID]; ok {
				logMap["operator_email"] = email
			}
		}

		enrichedLogs = append(enrichedLogs, logMap)
	}

	response.Success(w, map[string]interface{}{
		"logs":      enrichedLogs,
		"total":     result.Total,
		"page":      filter.Page,
		"page_size": filter.PageSize,
	})
}

func parseIntParam(r *http.Request, param string, defaultValue int) int {
	if valueStr := r.URL.Query().Get(param); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil && value > 0 {
			return value
		}
	}
	return defaultValue
}
