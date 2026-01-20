package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// HTTPTestClient wraps httptest for easier API testing
type HTTPTestClient struct {
	handler   http.Handler
	authToken string
}

// NewHTTPTestClient creates a new HTTP test client
func NewHTTPTestClient(handler http.Handler) *HTTPTestClient {
	return &HTTPTestClient{
		handler: handler,
	}
}

// SetAuthToken sets the JWT token for authenticated requests
func (c *HTTPTestClient) SetAuthToken(token string) {
	c.authToken = token
}

// GET performs a GET request
func (c *HTTPTestClient) GET(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	return w
}

// POST performs a POST request with JSON body
func (c *HTTPTestClient) POST(path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", path, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	return w
}

// PUT performs a PUT request with JSON body
func (c *HTTPTestClient) PUT(path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", path, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	return w
}

// DELETE performs a DELETE request
func (c *HTTPTestClient) DELETE(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)
	return w
}

// AssertStatusCode checks HTTP status code
func AssertStatusCode(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	require.Equal(t, expected, resp.Code, "unexpected status code. Body: %s", resp.Body.String())
}

// AssertJSONResponse parses and returns JSON response
func AssertJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, v interface{}) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	err = json.Unmarshal(body, v)
	require.NoError(t, err, "failed to parse JSON response: %s", string(body))
}

// AssertErrorResponse checks error response structure
func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	var result map[string]interface{}
	AssertJSONResponse(t, resp, &result)

	require.False(t, result["success"].(bool), "expected success=false")
	require.NotNil(t, result["error"], "expected error field")

	errorMap := result["error"].(map[string]interface{})
	require.Equal(t, expectedCode, errorMap["code"], "unexpected error code")
}

// AssertSuccessResponse checks success response structure
func AssertSuccessResponse(t *testing.T, resp *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	AssertJSONResponse(t, resp, &result)

	require.True(t, result["success"].(bool), "expected success=true")
	require.NotNil(t, result["data"], "expected data field")

	return result["data"].(map[string]interface{})
}
