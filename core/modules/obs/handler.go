package obs

import (
	"net/http"

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
// @Security Bearer
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
// @Security Bearer
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
// @Security Bearer
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
// @Security Bearer
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
