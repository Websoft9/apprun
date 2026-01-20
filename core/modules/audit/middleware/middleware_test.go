package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apprun/modules/audit"
	"apprun/modules/audit/service"
	"apprun/modules/audit/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage implements Storage interface for testing
type MockStorage struct {
	entries []*storage.AuditEntry
}

func (m *MockStorage) Write(ctx context.Context, entry *storage.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

//nolint:gocritic // hugeParam: MockStorage mirrors Storage interface
func (m *MockStorage) Query(ctx context.Context, filter storage.AuditFilter) (*storage.AuditQueryResult, error) {
	return nil, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestMiddleware_Handler_RecordLog(t *testing.T) {
	mockStorage := &MockStorage{}
	svcConfig := audit.ServiceConfig{Enabled: true, BufferSize: 10, WorkerCount: 1}
	svc, err := service.NewService(mockStorage, svcConfig, nil)
	require.NoError(t, err)
	err = svc.Start()
	require.NoError(t, err)
	defer svc.Shutdown(time.Second)

	mwConfig := audit.MiddlewareConfig{Enabled: true}
	mw := New(svc, mwConfig)

	// Create request
	req := httptest.NewRequest("POST", "/api/users", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.1")
	req.Header.Set("User-Agent", "TestAgent")
	w := httptest.NewRecorder()

	// Handler that does nothing
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	mw.Handler(next).ServeHTTP(w, req)

	// Wait for async log
	time.Sleep(100 * time.Millisecond)

	// Verify
	assert.Equal(t, 1, len(mockStorage.entries))
	if len(mockStorage.entries) > 0 {
		entry := mockStorage.entries[0]
		assert.Equal(t, "10.0.0.1", entry.IPAddress)
		assert.Equal(t, "resource.create", entry.Action)
		assert.Equal(t, "/api/users", entry.Path)
		assert.Equal(t, 201, entry.StatusCode)
	}
}

func TestMiddleware_Handler_ExcludePath(t *testing.T) {
	mockStorage := &MockStorage{}
	svcConfig := audit.ServiceConfig{Enabled: true, BufferSize: 10, WorkerCount: 1}
	svc, _ := service.NewService(mockStorage, svcConfig, nil)
	svc.Start()
	defer svc.Shutdown(time.Second)

	mwConfig := audit.MiddlewareConfig{
		Enabled:      true,
		ExcludePaths: []string{"/health"},
	}
	mw := New(svc, mwConfig)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mw.Handler(next).ServeHTTP(w, req)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, len(mockStorage.entries))
}

func TestMiddleware_ExtractIP(t *testing.T) {
	svcConfig := audit.ServiceConfig{Enabled: true}
	svc, _ := service.NewService(&MockStorage{}, svcConfig, nil)
	mw := New(svc, audit.MiddlewareConfig{Enabled: true})

	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "10.0.0.1, 10.0.0.2"},
			expected: "10.0.0.1",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "10.0.0.2"},
			expected: "10.0.0.2",
		},
		{
			name:     "RemoteAddr",
			remote:   "10.0.0.3:12345",
			expected: "10.0.0.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.remote != "" {
				req.RemoteAddr = tt.remote
			}

			ip := mw.extractIPAddress(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestMiddleware_DetermineAction(t *testing.T) {
	svc, _ := service.NewService(&MockStorage{}, audit.ServiceConfig{Enabled: true}, nil)
	mw := New(svc, audit.MiddlewareConfig{Enabled: true})

	tests := []struct {
		method   string
		path     string
		code     int
		expected string
	}{
		{"POST", "/auth/login", 200, "auth.login"},
		{"POST", "/auth/login", 401, "auth.login_failed"},
		{"GET", "/api/users", 403, "permission.denied"},
		{"POST", "/api/users", 201, "resource.create"},
		{"GET", "/api/users", 200, "resource.access"},
		{"PUT", "/api/users/1", 200, "resource.update"},
		{"DELETE", "/api/users/1", 204, "resource.delete"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			action := mw.determineAction(tt.method, tt.path, tt.code)
			assert.Equal(t, tt.expected, action)
		})
	}
}
