package obs

import (
	"net/http"
	"strconv"
	"time"

	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// MetricsHandler handles metrics HTTP requests
type MetricsHandler struct {
	service *MetricsService
}

// NewMetricsHandler creates a new metrics handler instance
func NewMetricsHandler(service *MetricsService) *MetricsHandler {
	return &MetricsHandler{
		service: service,
	}
}

// GetAll handles GET /api/metrics - returns all metrics
// @Summary Get All Metrics
// @Description Get all platform metrics including users, system, auth, and performance
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=AllMetrics} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics [get]
func (h *MetricsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetAllMetrics(ctx)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, metrics)
}

// GetUsers handles GET /api/metrics/users - returns user metrics
// @Summary Get User Metrics
// @Description Get user statistics including total, active, admin, banned, and registration metrics
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=UserMetrics} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/users [get]
func (h *MetricsHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetUserMetrics(ctx)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, metrics)
}

// GetSystem handles GET /api/metrics/system - returns system health metrics
// @Summary Get System Metrics
// @Description Get system health metrics including CPU, memory, and disk usage
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=SystemMetrics} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/system [get]
func (h *MetricsHandler) GetSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetSystemMetrics(ctx)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, metrics)
}

// GetPerformance handles GET /api/metrics/performance - returns performance metrics
// @Summary Get Performance Metrics
// @Description Get application performance metrics including response times and throughput
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=PerformanceMetrics} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/performance [get]
func (h *MetricsHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetPerformanceMetrics(ctx)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, metrics)
}

// GetHistory handles GET /api/metrics/history - returns historical metrics
// @Summary Get Historical Metrics
// @Description Get historical metrics data from storage (Story 9.1 - Storage integration)
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param name query string true "Metric name (e.g., user_count_total, system_cpu_percent)"
// @Param duration query string false "Time range duration (e.g., 1h, 24h, 7d)" default(24h)
// @Param start query string false "Start time (RFC3339 format)"
// @Param end query string false "End time (RFC3339 format)"
// @Param limit query int false "Maximum results to return" default(1000)
// @Success 200 {object} response.Response{data=HistoryResponse} "Success"
// @Failure 400 {object} response.Response "Bad Request - Missing metric name"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/history [get]
func (h *MetricsHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	// Validate required parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		response.ErrorWithRequest(w, r, 400, "INVALID_REQUEST", "Metric name is required")
		return
	}

	// Parse time range
	var start, end time.Time
	var err error

	// Try explicit start/end first
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr != "" && endStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			response.ErrorWithRequest(w, r, 400, "INVALID_TIME_FORMAT", "Invalid start time format")
			return
		}
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			response.ErrorWithRequest(w, r, 400, "INVALID_TIME_FORMAT", "Invalid end time format")
			return
		}
	} else {
		// Use duration (default 24h)
		durationStr := r.URL.Query().Get("duration")
		if durationStr == "" {
			durationStr = "24h"
		}
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			response.ErrorWithRequest(w, r, 400, "INVALID_DURATION", "Invalid duration format")
			return
		}
		end = time.Now()
		start = end.Add(-duration)
	}

	// Parse limit (default 1000)
	limit := 1000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Query from storage with degradation strategy
	history, err := h.service.GetMetricsHistory(ctx, name, start, end, limit)
	if err != nil {
		// Story 9.1: Degradation - return empty data instead of error
		log.Warn("Storage unavailable, returning empty history",
			logger.Field{Key: "metric", Value: name},
			logger.Field{Key: "error", Value: err},
		)
		history = &HistoryResponse{
			Metrics: []MetricPoint{},
			Count:   0,
			Start:   start,
			End:     end,
			HasMore: false,
		}
	}

	response.SuccessWithRequest(w, r, history)
}
