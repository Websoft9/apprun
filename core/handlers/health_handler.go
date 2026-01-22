// Package handlers provides HTTP request handlers for the application.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apprun/ent"
	"apprun/pkg/cache"
	"apprun/pkg/health"
	"apprun/pkg/health/checkers"
	"apprun/pkg/logger"
	"apprun/pkg/version"
)

// HealthHandler handles comprehensive health check requests
type HealthHandler struct {
	checker *health.HealthChecker
}

// NewHealthHandler creates a new health handler with all component checkers
func NewHealthHandler(db *ent.Client, cacheClient cache.Client, metricsEnabled bool) *HealthHandler {
	// Create component checkers
	componentCheckers := []health.ComponentChecker{
		checkers.NewDatabaseChecker(db),
		checkers.NewRedisChecker(cacheClient),
		checkers.NewMetricsChecker(metricsEnabled),
	}

	// Create health checker with all components
	checker := health.NewHealthChecker("apprun", version.Version, componentCheckers...)

	return &HealthHandler{
		checker: checker,
	}
}

// Check handles GET /health requests
//
//	@Summary		Comprehensive Health Check
//	@Description	Check the health of apprun service and all its dependencies (database, cache, metrics storage)
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	health.HealthReport	"All components healthy"
//	@Success		207	{object}	health.HealthReport	"System degraded (non-critical component unhealthy)"
//	@Failure		503	{object}	health.HealthReport	"System unhealthy (critical component failed)"
//	@Router			/health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	// Create a timeout context for the entire health check (5 seconds max)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Perform health checks
	report := h.checker.CheckAll(ctx)

	// Log health check result
	logger.L().WithContext(ctx).Info("Health check completed",
		logger.Field{Key: "status", Value: string(report.Status)},
		logger.Field{Key: "components", Value: len(report.Components)},
	)

	// Pre-serialize to buffer to catch encoding errors before sending headers
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(report); err != nil {
		logger.L().WithContext(ctx).Error("Failed to encode health report",
			logger.Field{Key: "error", Value: err},
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send response with appropriate HTTP status code
	statusCode := report.Status.HTTPStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.L().WithContext(ctx).Error("Failed to write health response",
			logger.Field{Key: "error", Value: err},
		)
	}
}
