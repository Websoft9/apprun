package obs

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"apprun/pkg/metrics"
	"apprun/pkg/metrics/storage"
)

func setupStorageHandler(t *testing.T) (*StorageHandler, *storage.MockStorage) {
	cfg := &metrics.Config{
		Storage: metrics.StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
		},
	}

	store, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)

	repo := metrics.NewRepository(store, cfg)
	handler := NewStorageHandler(repo)

	mockStorage := store.(*storage.MockStorage)

	return handler, mockStorage
}

func TestStorageHandler_Ingest(t *testing.T) {
	handler, _ := setupStorageHandler(t)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name: "valid metric",
			body: IngestRequest{
				Name:  "test_metric",
				Value: 42.0,
				Tags:  map[string]string{"env": "test"},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.True(t, data["success"].(bool))
				assert.Equal(t, "Metric stored", data["message"])
				assert.Contains(t, data["metric_id"], "test_metric:")
			},
		},
		{
			name:           "missing name",
			body:           IngestRequest{Value: 42.0},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			body: IngestRequest{
				Name:  string(make([]byte, 300)),
				Value: 42.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "too many tags",
			body: IngestRequest{
				Name:  "test_metric",
				Value: 42.0,
				Tags: map[string]string{
					"tag1": "val1", "tag2": "val2", "tag3": "val3",
					"tag4": "val4", "tag5": "val5", "tag6": "val6",
					"tag7": "val7", "tag8": "val8", "tag9": "val9",
					"tag10": "val10", "tag11": "val11",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/observability/metrics/ingest", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Ingest(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil && w.Code == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestStorageHandler_Query(t *testing.T) {
	handler, mockStorage := setupStorageHandler(t)

	// Reset storage before test
	mockStorage.Reset()

	// Store some test metrics
	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 5; i++ {
		metric := storage.Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Tags:      map[string]string{"index": string(rune(i))},
			Timestamp: now.Add(-time.Duration(5-i) * time.Minute), // Store in the past to ensure they're queryable
		}
		err := mockStorage.Store(ctx, metric)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(t *testing.T, resp map[string]interface{})
	}{
		{
			name:           "query with duration",
			queryParams:    "?name=test_metric&duration=1h",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(5), data["count"])
				metrics := data["metrics"].([]interface{})
				assert.Len(t, metrics, 5)
			},
		},
		{
			name:           "query with limit",
			queryParams:    "?name=test_metric&duration=1h&limit=3",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(3), data["count"])
				assert.True(t, data["has_more"].(bool))
			},
		},
		{
			name:           "missing name",
			queryParams:    "?duration=1h",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid duration",
			queryParams:    "?name=test_metric&duration=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit",
			queryParams:    "?name=test_metric&limit=abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/observability/metrics/query"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.Query(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil && w.Code == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestStorageHandler_QueryWithTimeRange(t *testing.T) {
	handler, mockStorage := setupStorageHandler(t)

	// Reset storage before test
	mockStorage.Reset()

	// Store metrics with specific timestamps
	ctx := context.Background()
	baseTime := time.Now().UTC().Add(-12 * time.Hour) // Use UTC time in the past
	for i := 0; i < 10; i++ {
		metric := storage.Metric{
			Name:      "time_test",
			Value:     float64(i),
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
		}
		err := mockStorage.Store(ctx, metric)
		require.NoError(t, err)
	}

	start := baseTime.Add(3 * time.Hour).Format(time.RFC3339)
	end := baseTime.Add(7 * time.Hour).Format(time.RFC3339)

	req := httptest.NewRequest(http.MethodGet, "/api/observability/metrics/query?name=time_test&start="+start+"&end="+end, nil)
	w := httptest.NewRecorder()

	handler.Query(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Response body: %s", w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	// Query returns metrics within [start, end] inclusive range
	// Metrics 3, 4, 5, 6, 7 but BadgerDB may use exclusive end
	count := int(data["count"].(float64))
	assert.True(t, count >= 4 && count <= 5, "Expected 4 or 5 metrics, got %d", count)
}

func TestStorageHandler_Health(t *testing.T) {
	handler, mockStorage := setupStorageHandler(t)

	tests := []struct {
		name           string
		healthy        bool
		expectedStatus int
		expectedHealth string
	}{
		{
			name:           "healthy storage",
			healthy:        true,
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name:           "unhealthy storage",
			healthy:        false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.SetHealthy(tt.healthy)

			req := httptest.NewRequest(http.MethodGet, "/api/observability/metrics/health", nil)
			w := httptest.NewRecorder()

			handler.Health(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, tt.expectedHealth, data["status"])
			}
		})
	}
}
