// Package middleware provides HTTP middleware for automatic audit logging.
package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"apprun/modules/audit"
	"apprun/modules/audit/service"
	"apprun/modules/audit/storage"

	"github.com/google/uuid"
)

type Middleware struct {
	service *service.Service
	config  audit.MiddlewareConfig
}

func New(svc *service.Service, config audit.MiddlewareConfig) *Middleware {
	return &Middleware{
		service: svc,
		config:  config,
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.Enabled || m.isExcluded(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		startTime := time.Now()
		operatorID := m.extractOperatorID(r)
		ipAddress := m.extractIPAddress(r)

		ctx := r.Context()
		if operatorID != nil {
			ctx = service.SetOperatorIDInContext(ctx, *operatorID)
		}
		ctx = service.SetIPAddressInContext(ctx, ipAddress)
		r = r.WithContext(ctx)

		wrapper := &responseWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		go func() {
			entry := &storage.AuditEntry{
				Timestamp:      startTime,
				OperatorID:     operatorID,
				Method:         r.Method,
				Path:           r.URL.Path,
				IPAddress:      ipAddress,
				UserAgent:      r.UserAgent(),
				StatusCode:     wrapper.statusCode,
				ResponseTimeMs: int(time.Since(startTime).Milliseconds()),
			}

			entry.Action = m.determineAction(r.Method, r.URL.Path, wrapper.statusCode)

			if err := m.service.Log(r.Context(), entry); err != nil {
				log.Printf("[ERROR] Failed to log audit entry: %v", err)
			}
		}()
	})
}

func (m *Middleware) isExcluded(path string) bool {
	for _, excludePath := range m.config.ExcludePaths {
		if strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	return false
}

func (m *Middleware) extractOperatorID(r *http.Request) *uuid.UUID {
	if userID, ok := r.Context().Value("user_uuid").(uuid.UUID); ok {
		return &userID
	}
	if userIDStr, ok := r.Context().Value("user_uuid").(string); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return &userID
		}
	}
	return nil
}

func (m *Middleware) extractIPAddress(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (m *Middleware) determineAction(method, path string, statusCode int) string {
	if strings.Contains(path, "/auth/login") {
		if statusCode >= 200 && statusCode < 300 {
			return "auth.login"
		}
		return "auth.login_failed"
	}
	if strings.Contains(path, "/auth/logout") {
		return "auth.logout"
	}
	if strings.Contains(path, "/auth/refresh") {
		return "auth.token_refresh"
	}
	if statusCode == http.StatusForbidden {
		return "permission.denied"
	}
	switch method {
	case http.MethodPost:
		return "resource.create"
	case http.MethodPut, http.MethodPatch:
		return "resource.update"
	case http.MethodDelete:
		return "resource.delete"
	default:
		return "resource.access"
	}
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *responseWrapper) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
