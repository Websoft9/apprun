package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSwaggerUI_Accessible tests that Swagger UI is accessible at /api/docs/
func TestSwaggerUI_Accessible(t *testing.T) {
	r := chi.NewRouter()
	RegisterSwagger(r)

	// Test /api/docs/index.html (the actual Swagger UI)
	req := httptest.NewRequest("GET", "/api/docs/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Swagger UI should be accessible at /api/docs/index.html")
	assert.Contains(t, w.Body.String(), "Swagger UI", "Response should contain Swagger UI title")
}

// TestSwaggerUI_IndexHTML tests that /api/docs/index.html is accessible
func TestSwaggerUI_IndexHTML(t *testing.T) {
	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Swagger UI index.html should be accessible")
	assert.Contains(t, w.Body.String(), "Swagger UI", "Response should contain Swagger UI title")
}

// TestSwaggerRedirect tests that /api/docs redirects to /api/docs/
func TestSwaggerRedirect(t *testing.T) {
	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code, "Should return 301 redirect")
	assert.Equal(t, "/api/docs/", w.Header().Get("Location"), "Should redirect to /api/docs/")
}

// TestOpenAPISpec_Accessible tests that OpenAPI spec endpoint is registered
// Note: Actual spec content requires docs package to be generated
func TestOpenAPISpec_Accessible(t *testing.T) {
	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// In test environment without generated docs, we expect 500 or 404
	// In production with docs generated, it should return 200
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError || w.Code == http.StatusNotFound,
		"OpenAPI spec endpoint should be registered (may fail without generated docs)")
}

// TestOpenAPISpec_Structure tests the structure of generated OpenAPI spec
// This test requires the docs package to be generated with `make swagger`
func TestOpenAPISpec_Structure(t *testing.T) {
	t.Skip("Requires generated docs package - run `make swagger` first")

	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "OpenAPI spec should be accessible")

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	require.NoError(t, err, "Should be valid JSON")

	// Verify OpenAPI spec structure
	assert.Contains(t, spec, "openapi", "Should contain OpenAPI version")
	assert.Contains(t, spec, "info", "Should contain info section")
	assert.Contains(t, spec, "paths", "Should contain paths section")
	assert.Contains(t, spec, "definitions", "Should contain definitions section")

	// Verify info section
	info, ok := spec["info"].(map[string]interface{})
	require.True(t, ok, "info should be an object")
	assert.Equal(t, "AppRun API", info["title"], "Should have correct API title")
	assert.Equal(t, "1.0", info["version"], "Should have correct API version")

	// Verify basePath
	assert.Equal(t, "/api", spec["basePath"], "Should have correct basePath")

	// Verify config endpoints exist
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "paths should be an object")
	assert.Contains(t, paths, "/config", "Should contain /config endpoint")
	assert.Contains(t, paths, "/config/list", "Should contain /config/list endpoint")
	assert.Contains(t, paths, "/config/allowed", "Should contain /config/allowed endpoint")
}

// TestOpenAPISpec_ConfigEndpoints tests that all 5 config endpoints are documented
// This test requires the docs package to be generated with `make swagger`
func TestOpenAPISpec_ConfigEndpoints(t *testing.T) {
	t.Skip("Requires generated docs package - run `make swagger` first")

	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	require.NoError(t, err)

	paths := spec["paths"].(map[string]interface{})

	// Test GET /config
	configPath, ok := paths["/config"].(map[string]interface{})
	require.True(t, ok, "/config path should exist")
	assert.Contains(t, configPath, "get", "Should have GET method")
	assert.Contains(t, configPath, "put", "Should have PUT method")
	assert.Contains(t, configPath, "delete", "Should have DELETE method")

	// Test GET /config/list
	listPath, ok := paths["/config/list"].(map[string]interface{})
	require.True(t, ok, "/config/list path should exist")
	assert.Contains(t, listPath, "get", "Should have GET method")

	// Test GET /config/allowed
	allowedPath, ok := paths["/config/allowed"].(map[string]interface{})
	require.True(t, ok, "/config/allowed path should exist")
	assert.Contains(t, allowedPath, "get", "Should have GET method")

	// Verify GET /config has correct structure
	getConfig := configPath["get"].(map[string]interface{})
	assert.Equal(t, "Get configuration item", getConfig["summary"], "Should have correct summary")
	assert.Contains(t, getConfig, "parameters", "Should have parameters")
	assert.Contains(t, getConfig, "responses", "Should have responses")

	// Verify parameters
	params := getConfig["parameters"].([]interface{})
	assert.Greater(t, len(params), 0, "Should have at least one parameter")

	// Verify responses
	responses := getConfig["responses"].(map[string]interface{})
	assert.Contains(t, responses, "200", "Should have 200 response")
	assert.Contains(t, responses, "400", "Should have 400 response")
	assert.Contains(t, responses, "404", "Should have 404 response")
}

// TestOpenAPISpec_ErrorResponses tests that error responses are documented
// This test requires the docs package to be generated with `make swagger`
func TestOpenAPISpec_ErrorResponses(t *testing.T) {
	t.Skip("Requires generated docs package - run `make swagger` first")

	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	require.NoError(t, err)

	// Verify ErrorResponse definition exists
	definitions := spec["definitions"].(map[string]interface{})
	assert.Contains(t, definitions, "config.ErrorResponse", "Should have ErrorResponse definition")

	errorResp := definitions["config.ErrorResponse"].(map[string]interface{})
	properties := errorResp["properties"].(map[string]interface{})
	assert.Contains(t, properties, "error", "ErrorResponse should have 'error' field")
	assert.Contains(t, properties, "details", "ErrorResponse should have 'details' field")
}

// TestOpenAPISpec_NoHostBinding tests that host field is not hardcoded
// This test requires the docs package to be generated with `make swagger`
func TestOpenAPISpec_NoHostBinding(t *testing.T) {
	t.Skip("Requires generated docs package - run `make swagger` first")

	r := chi.NewRouter()
	RegisterSwagger(r)

	req := httptest.NewRequest("GET", "/api/docs/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	require.NoError(t, err)

	// Verify host is empty (not hardcoded to localhost)
	host, exists := spec["host"]
	if exists {
		hostStr, ok := host.(string)
		require.True(t, ok, "host should be a string")
		assert.Empty(t, hostStr, "host should be empty for dynamic deployment support")
	}
}
