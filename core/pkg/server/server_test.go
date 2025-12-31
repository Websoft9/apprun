package server

import (
"net/http"
"net/http/httptest"
"testing"
"time"

"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
cfg := DefaultConfig()
assert.Equal(t, "8080", cfg.HTTPPort)
assert.Equal(t, "8443", cfg.HTTPSPort)
assert.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
assert.True(t, cfg.EnableHTTPWithHTTPS)
}

func TestServerConfig(t *testing.T) {
cfg := &Config{
HTTPPort:            "9090",
HTTPSPort:           "9443",
ShutdownTimeout:     10 * time.Second,
EnableHTTPWithHTTPS: false,
}

assert.Equal(t, "9090", cfg.HTTPPort)
assert.Equal(t, "9443", cfg.HTTPSPort)
assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
assert.False(t, cfg.EnableHTTPWithHTTPS)
}

func TestServerHandler(t *testing.T) {
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
w.Write([]byte("OK"))
})

req := httptest.NewRequest("GET", "/", nil)
w := httptest.NewRecorder()
handler.ServeHTTP(w, req)

assert.Equal(t, http.StatusOK, w.Code)
assert.Equal(t, "OK", w.Body.String())
}
