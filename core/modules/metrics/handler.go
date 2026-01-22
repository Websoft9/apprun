package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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

// GetSnapshot handles GET /api/metrics/snapshot - returns aggregated metrics view
// @Summary Get Metrics Snapshot
// @Description Get aggregated metrics view for specific scope (system, users, or all)
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param scope query string false "Scope filter: system, users, all" default(all)
// @Success 200 {object} response.Response{data=SnapshotResponse} "Success"
// @Failure 400 {object} response.Response "Bad Request - Invalid scope"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/snapshot [get]
func (h *MetricsHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse scope parameter
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "all"
	}

	// Validate scope
	validScopes := map[string]bool{
		"system": true,
		"users":  true,
		"all":    true,
	}
	if !validScopes[scope] {
		response.ErrorWithRequest(w, r, 400, "INVALID_SCOPE", "Invalid scope. Must be: system, users, or all")
		return
	}

	snapshot, err := h.service.GetSnapshot(ctx, scope)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, snapshot)
}

// GetKeys handles GET /api/metrics/keys - returns available metric keys
// @Summary Get Metric Keys
// @Description List all available metric names for history queries
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param source query string false "Filter by source: system or user"
// @Success 200 {object} response.Response{data=KeysResponse} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/keys [get]
func (h *MetricsHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	source := r.URL.Query().Get("source")
	keys, err := h.service.GetMetricKeys(ctx, source)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, keys)
}

// GetScopes handles GET /api/metrics/scopes - returns available snapshot scopes
// @Summary Get Snapshot Scopes
// @Description List all available scopes for snapshot queries and their included metrics
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=ScopesResponse} "Success"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires platform_admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/scopes [get]
func (h *MetricsHandler) GetScopes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	scopes, err := h.service.GetScopes(ctx)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, scopes)
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
		duration, err := time.ParseDuration(durationStr) //nolint:govet  //nolint:govet 
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
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 { //nolint:govet // Shadow acceptable in multi-return
			limit = parsedLimit
		}
	}

	// Parse tags filter (optional)
	tags := make(map[string]string)
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "tags[") && strings.HasSuffix(key, "]") {
			tagKey := strings.TrimSuffix(strings.TrimPrefix(key, "tags["), "]")
			if len(values) > 0 {
				tags[tagKey] = values[0]
			}
		}
	}

	// Query from storage with degradation strategy
	history, err := h.service.GetMetricsHistory(ctx, name, start, end, limit, tags)
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

// Ingest handles POST /api/metrics/ingest - ingests metrics batch
// @Summary Ingest Metrics
// @Description Batch ingest metrics with optional tags and timestamps
// @Tags metrics
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param metrics body []IngestRequest true "Array of metric data"
// @Success 200 {object} response.Response{data=IngestBatchResponse} "Success"
// @Failure 400 {object} response.Response "Bad Request - Invalid metrics format"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires write permission"
// @Failure 429 {object} response.Response "Too Many Requests"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/metrics/ingest [post]
func (h *MetricsHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var requests []IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body: "+err.Error())
		return
	}

	if len(requests) == 0 {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "At least one metric required")
		return
	}

	if len(requests) > 1000 {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Too many metrics in batch (max 1000)")
		return
	}

	// Process batch
	result, err := h.service.IngestBatch(ctx, requests)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, result)
}
