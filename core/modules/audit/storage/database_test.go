package storage

import (
	"context"
	"testing"
	"time"

	"apprun/ent/enttest"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage_Write(t *testing.T) {
	// Use SQLite for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	storage := NewDatabaseStorage(client)
	ctx := context.Background()

	operatorID := uuid.New()
	entry := &AuditEntry{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		OperatorID: &operatorID,
		Action:     "user.create",
		TargetID:   "target-123",
		TargetType: "user",
		Changes: map[string]interface{}{
			"email": "test@example.com",
		},
		IPAddress:      "192.168.1.1",
		UserAgent:      "TestAgent/1.0",
		StatusCode:     201,
		ResponseTimeMs: 50,
		Method:         "POST",
		Path:           "/api/users",
	}

	err := storage.Write(ctx, entry)
	require.NoError(t, err)

	// Verify entry was written
	filter := AuditFilter{
		OperatorID:   &operatorID,
		Page:         1,
		PageSize:     10,
		IncludeTotal: true,
	}

	result, err := storage.Query(ctx, filter)
	require.NoError(t, err)
	assert.Greater(t, result.Total, int64(0))
	assert.NotEmpty(t, result.Logs)
}

func TestDatabaseStorage_Query(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	storage := NewDatabaseStorage(client)
	ctx := context.Background()

	// Write test entries
	operatorID := uuid.New()
	for i := 0; i < 5; i++ {
		entry := &AuditEntry{
			ID:         uuid.New(),
			Timestamp:  time.Now(),
			OperatorID: &operatorID,
			Action:     "test.action",
			IPAddress:  "127.0.0.1",
		}
		_ = storage.Write(ctx, entry)
	}

	// Query with pagination
	filter := AuditFilter{
		OperatorID:   &operatorID,
		Page:         1,
		PageSize:     3,
		IncludeTotal: true,
	}

	result, err := storage.Query(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Logs))
	assert.GreaterOrEqual(t, result.Total, int64(5))
}

func TestDatabaseStorage_QueryByAction(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	storage := NewDatabaseStorage(client)
	ctx := context.Background()

	// Write entries with different actions
	entries := []*AuditEntry{
		{ID: uuid.New(), Timestamp: time.Now(), Action: "user.create"},
		{ID: uuid.New(), Timestamp: time.Now(), Action: "user.update"},
		{ID: uuid.New(), Timestamp: time.Now(), Action: "user.delete"},
	}

	for _, entry := range entries {
		_ = storage.Write(ctx, entry)
	}

	// Query for specific action
	filter := AuditFilter{
		Action:       "user.create",
		Page:         1,
		PageSize:     10,
		IncludeTotal: true,
	}

	result, err := storage.Query(ctx, filter)
	require.NoError(t, err)
	assert.Greater(t, result.Total, int64(0))

	// Verify all returned logs have the correct action
	for _, log := range result.Logs {
		assert.Equal(t, "user.create", log.Action)
	}
}
