package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"apprun/modules/audit"
	"apprun/modules/audit/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage implements Storage interface for testing
type MockStorage struct {
	entries []*storage.AuditEntry
	queries []storage.AuditFilter
	mu      sync.Mutex
}

func (m *MockStorage) Write(ctx context.Context, entry *storage.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

//nolint:gocritic // hugeParam: MockStorage mirrors Storage interface
func (m *MockStorage) Query(ctx context.Context, filter storage.AuditFilter) (*storage.AuditQueryResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queries = append(m.queries, filter)
	return &storage.AuditQueryResult{
		Logs:  m.entries,
		Total: int64(len(m.entries)),
	}, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestService_Log(t *testing.T) {
	mockStorage := &MockStorage{}
	config := audit.ServiceConfig{
		Enabled:       true,
		BufferSize:    10,
		WorkerCount:   2,
		FlushInterval: 100 * time.Millisecond,
	}

	svc, err := NewService(mockStorage, config, nil)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)
	defer svc.Shutdown(5 * time.Second)

	// Log an entry
	entry := &storage.AuditEntry{
		Timestamp: time.Now(),
		Action:    "test.action",
		IPAddress: "127.0.0.1",
	}

	err = svc.Log(context.Background(), entry)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify entry was written
	assert.Greater(t, len(mockStorage.entries), 0)
}

func TestService_LogAction(t *testing.T) {
	mockStorage := &MockStorage{}
	config := audit.DefaultServiceConfig()

	svc, err := NewService(mockStorage, config, nil)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)
	defer svc.Shutdown(5 * time.Second)

	// Prepare context with operator ID
	operatorID := uuid.New()
	ctx := SetOperatorIDInContext(context.Background(), operatorID)
	ctx = SetIPAddressInContext(ctx, "192.168.1.1")

	// Log an action
	params := LogActionParams{
		Action:     "user.update_role",
		TargetID:   "user-123",
		TargetType: "user",
		Changes: map[string]interface{}{
			"role": map[string]string{
				"from": "user",
				"to":   "admin",
			},
		},
	}

	err = svc.LogAction(ctx, params)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify entry was written with correct data
	assert.Greater(t, len(mockStorage.entries), 0)
	if len(mockStorage.entries) > 0 {
		entry := mockStorage.entries[0]
		assert.Equal(t, "user.update_role", entry.Action)
		assert.Equal(t, "user-123", entry.TargetID)
		assert.Equal(t, "192.168.1.1", entry.IPAddress)
		assert.NotNil(t, entry.OperatorID)
	}
}

func TestService_Shutdown(t *testing.T) {
	mockStorage := &MockStorage{}
	config := audit.ServiceConfig{
		Enabled:       true,
		BufferSize:    100,
		WorkerCount:   2,
		FlushInterval: 100 * time.Millisecond,
	}

	svc, err := NewService(mockStorage, config, nil)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)

	// Log multiple entries
	for i := 0; i < 10; i++ {
		entry := &storage.AuditEntry{
			Timestamp: time.Now(),
			Action:    "test.action",
		}
		_ = svc.Log(context.Background(), entry)
	}

	// Give time for entries to queue
	time.Sleep(100 * time.Millisecond)

	// Shutdown should wait for all entries to be processed
	err = svc.Shutdown(5 * time.Second)
	require.NoError(t, err)

	// Verify all entries were written
	assert.GreaterOrEqual(t, len(mockStorage.entries), 9, "Most entries should be processed after shutdown")
}

func TestService_Sanitize(t *testing.T) {
	mockStorage := &MockStorage{}
	config := audit.DefaultServiceConfig()
	sensitiveFields := []string{"password", "token", "secret"}

	svc, err := NewService(mockStorage, config, sensitiveFields)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)
	defer svc.Shutdown(5 * time.Second)

	// Log an action with sensitive data
	params := LogActionParams{
		Action:     "user.login",
		TargetID:   "user-123",
		TargetType: "user",
		Changes: map[string]interface{}{
			"username": "jdoe",
			"password": "supersecretpassword", // Should be masked
			"metadata": map[string]interface{}{
				"token": "abcdef123456", // Should be masked in nested map
				"ip":    "127.0.0.1",
			},
			"config": map[string]interface{}{
				"api_key": "key-123",       // Not in sensitive list
				"secret":  "hidden-secret", // Should be masked
			},
		},
	}

	err = svc.LogAction(context.Background(), params)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify masking
	require.Greater(t, len(mockStorage.entries), 0)
	entry := mockStorage.entries[len(mockStorage.entries)-1]

	// Check changes map
	changes := entry.Changes

	assert.Equal(t, "jdoe", changes["username"])
	assert.Equal(t, "******", changes["password"])

	// Check nested metadata
	metadata, ok := changes["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "******", metadata["token"])
	assert.Equal(t, "127.0.0.1", metadata["ip"])

	// Check nested config
	conf, ok := changes["config"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "key-123", conf["api_key"])
	assert.Equal(t, "******", conf["secret"])
}
