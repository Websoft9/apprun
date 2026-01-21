package obs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"apprun/pkg/metrics"
	"apprun/pkg/response"
)

// metricNamePattern defines valid metric name format (alphanumeric, underscore, hyphen, dot)
var metricNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]*$`)

// validateMetricName validates metric name format to prevent injection attacks
// Rules:
// - Must start with a letter
// - Can contain: letters, numbers, underscore, hyphen, dot
// - Cannot contain: colon (used as key separator), special chars
func validateMetricName(name string) error {
	if !metricNamePattern.MatchString(name) {
		return fmt.Errorf("invalid metric name format: must start with letter and contain only [a-zA-Z0-9_.-]")
	}
	return nil
}

// StorageHandler handles metrics storage HTTP requests
type StorageHandler struct {
	repo *metrics.Repository
}

// NewStorageHandler creates a new storage handler instance
func NewStorageHandler(repo *metrics.Repository) *StorageHandler {
	return &StorageHandler{
		repo: repo,
	}
}

// IngestRequest represents the request body for metric ingestion
type IngestRequest struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp *time.Time        `json:"timestamp,omitempty"`
}

// IngestResponse represents the response for metric ingestion
type IngestResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	MetricID string `json:"metric_id"`
}

// QueryResponse represents the response for metric queries
type QueryResponse struct {
	Metrics []MetricData `json:"metrics"`
	Count   int          `json:"count"`
	Start   time.Time    `json:"start"`
	End     time.Time    `json:"end"`
	HasMore bool         `json:"has_more"`
}

// MetricData represents a single metric data point
type MetricData struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// HealthResponse represents the storage health response
type HealthResponse struct {
	Status       string                 `json:"status"`
	Backend      string                 `json:"backend"`
	DiskUsageMB  float64                `json:"disk_usage_mb,omitempty"`
	MetricsCount int                    `json:"metrics_count,omitempty"`
	Stats        map[string]interface{} `json:"stats,omitempty"`
}

// Ingest handles POST /api/observability/metrics/ingest
// @Summary Ingest Metric
// @Description Store a metric with optional tags and timestamp
// @Tags observability
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param metric body IngestRequest true "Metric data"
// @Success 200 {object} response.Response{data=IngestResponse} "Success"
// @Failure 400 {object} response.Response "Bad Request - Invalid metric format"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires admin role"
// @Failure 429 {object} response.Response "Too Many Requests"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/observability/metrics/ingest [post]
func (h *StorageHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body: "+err.Error())
		return
	}

	// Validate input
	if req.Name == "" {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Metric name is required")
		return
	}
	if len(req.Name) > 256 {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Metric name too long (max 256 characters)")
		return
	}
	// Validate metric name format to prevent injection
	if err := validateMetricName(req.Name); err != nil {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if len(req.Tags) > 10 {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Too many tags (max 10)")
		return
	}

	// Use provided timestamp or current time
	ts := time.Now()
	if req.Timestamp != nil {
		ts = *req.Timestamp
	}

	// Record metric
	err := h.repo.RecordMetric(ctx, req.Name, req.Value, req.Tags)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Generate metric ID
	metricID := req.Name + ":" + strconv.FormatInt(ts.UnixNano(), 10)

	resp := IngestResponse{
		Success:  true,
		Message:  "Metric stored",
		MetricID: metricID,
	}

	response.SuccessWithRequest(w, r, resp)
}

// Query handles GET /api/observability/metrics/query
// @Summary Query Metrics
// @Description Retrieve metrics by name and time range
// @Tags observability
// @Security BearerAuth
// @Produce json
// @Param name query string true "Metric name"
// @Param duration query string false "Time range (e.g., 1h, 24h)" default(24h)
// @Param start query string false "Start timestamp (RFC3339)"
// @Param end query string false "End timestamp (RFC3339)"
// @Param limit query int false "Max results" default(1000)
// @Success 200 {object} response.Response{data=QueryResponse} "Success"
// @Failure 400 {object} response.Response "Bad Request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - requires admin role"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/observability/metrics/query [get]
func (h *StorageHandler) Query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	name := r.URL.Query().Get("name")
	if name == "" {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Metric name is required")
		return
	}
	// Validate metric name format
	if err := validateMetricName(name); err != nil {
		response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Parse time range
	var start, end time.Time
	var err error

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	durationStr := r.URL.Query().Get("duration")

	if durationStr == "" {
		durationStr = "24h"
	}

	if startStr != "" && endStr != "" {
		// Use explicit start/end
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid start timestamp: "+err.Error())
			return
		}
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid end timestamp: "+err.Error())
			return
		}
	} else {
		// Use duration
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid duration: "+err.Error())
			return
		}
		end = time.Now()
		start = end.Add(-duration)
	}

	// Parse limit (default 1000, max 10000)
	limit := 1000
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			response.ErrorWithRequest(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid limit")
			return
		}
		limit = parsedLimit
		if limit > 10000 {
			limit = 10000 // enforce max
		}
	}

	// Query metrics with limit
	metrics, err := h.repo.GetMetricsByRange(ctx, name, start, end, limit+1)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Check if there are more results
	hasMore := len(metrics) > limit
	if hasMore {
		metrics = metrics[:limit]
	}

	// Convert to response format
	metricData := make([]MetricData, len(metrics))
	for i, m := range metrics {
		metricData[i] = MetricData{
			Name:      m.Name,
			Value:     m.Value,
			Tags:      m.Tags,
			Timestamp: m.Timestamp,
		}
	}

	resp := QueryResponse{
		Metrics: metricData,
		Count:   len(metricData),
		Start:   start,
		End:     end,
		HasMore: hasMore,
	}

	response.SuccessWithRequest(w, r, resp)
}

// Health handles GET /api/observability/metrics/health
// @Summary Storage Health Check
// @Description Check the health status of the metrics storage backend
// @Tags observability
// @Produce json
// @Success 200 {object} response.Response{data=HealthResponse} "Healthy"
// @Failure 503 {object} response.Response{data=HealthResponse} "Unhealthy"
// @Router /api/observability/metrics/health [get]
func (h *StorageHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.repo.Health(ctx)

	resp := HealthResponse{
		Backend: h.repo.GetBackend(),
	}

	if err != nil {
		resp.Status = "unhealthy"
		response.ErrorWithRequest(w, r, http.StatusServiceUnavailable, "STORAGE_UNHEALTHY", "Storage unhealthy: "+err.Error())
		return
	}

	resp.Status = "healthy"
	response.SuccessWithRequest(w, r, resp)
}
