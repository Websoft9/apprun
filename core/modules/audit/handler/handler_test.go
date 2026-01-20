package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apprun/ent/enttest"
	"apprun/modules/audit"
	"apprun/modules/audit/service"
	"apprun/modules/audit/storage"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage for handler tests
type MockStorage struct {
	entries []*storage.AuditEntry
}

func (m *MockStorage) Write(ctx context.Context, entry *storage.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

//nolint:gocritic // hugeParam: MockStorage mirrors Storage interface
func (m *MockStorage) Query(ctx context.Context, filter storage.AuditFilter) (*storage.AuditQueryResult, error) {
	return &storage.AuditQueryResult{
		Logs:  m.entries,
		Total: int64(len(m.entries)),
	}, nil
}

func (m *MockStorage) Close() error { return nil }

func TestHandler_QueryLogs(t *testing.T) {
	// Setup Ent
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ctx := context.Background()

	// Create a user
	u, err := client.User.Create().
		SetEmail("admin@example.com").
		SetPasswordHash("secret").
		Save(ctx)
	require.NoError(t, err)

	// Setup Audit Service with MockStorage
	operatorID := u.UUID
	mockEntry := &storage.AuditEntry{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		OperatorID: &operatorID,
		Action:     "test.action",
		IPAddress:  "127.0.0.1",
	}
	mockStorage := &MockStorage{
		entries: []*storage.AuditEntry{mockEntry},
	}
	svc, _ := service.NewService(mockStorage, audit.ServiceConfig{Enabled: true}, nil)

	// Setup Handler
	h := New(svc, client)

	// Make Request
	req := httptest.NewRequest("GET", "/api/admin/audit-logs?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	h.QueryLogs(w, req)

	// Verify Response Code
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify Response Body
	var resp struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// Verify Logs
	logs, ok := resp.Data["logs"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, len(logs))

	log0 := logs[0].(map[string]interface{})
	assert.Equal(t, "test.action", log0["action"])

	// Verify Enrichment (N+1 optimization check)
	assert.Equal(t, "admin@example.com", log0["operator_email"])
}
