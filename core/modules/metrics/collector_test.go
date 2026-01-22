package metrics

import (
	"context"
	"testing"
	"time"

	"apprun/ent"
	"apprun/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	return client
}

func TestMetricsCollector_CollectUserMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	collector := NewMetricsCollector(client, nil)

	// Create test users
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	sevenDaysAgo := now.AddDate(0, 0, -7)

	// Create 10 users with various states
	for i := 0; i < 10; i++ {
		builder := client.User.Create().
			SetEmail("user" + string(rune(i+'0')) + "@example.com").
			SetPasswordHash("hash").
			SetStatus(1) // active

		// Set different roles
		switch {
		case i < 2:
			builder = builder.SetRole("platform_admin")
		default:
			builder = builder.SetRole("platform_user")
		}

		// Set creation times
		switch {
		case i < 3:
			builder = builder.SetCreatedAt(today) // today
		case i < 8:
			builder = builder.SetCreatedAt(sevenDaysAgo.Add(time.Hour * 24)) // within 7 days
		default:
			builder = builder.SetCreatedAt(sevenDaysAgo.Add(-time.Hour * 48)) // older
		}

		// Create one banned user
		if i == 9 {
			builder = builder.SetStatus(0)
		}

		_, err := builder.Save(ctx)
		require.NoError(t, err)
	}

	// Collect metrics
	metrics, err := collector.CollectUserMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify metrics
	assert.Equal(t, 10, metrics.TotalUsers, "Total users should be 10")
	assert.Equal(t, 9, metrics.ActiveUsers, "Active users should be 9 (10 - 1 banned)")
	assert.Equal(t, 2, metrics.AdminUsers, "Admin users should be 2")
	assert.Equal(t, 1, metrics.BannedUsers, "Banned users should be 1")
	assert.Equal(t, 3, metrics.NewUsersToday, "New users today should be 3")
	assert.Equal(t, 8, metrics.UserRegistrationsLast7Days, "Users in last 7 days should be 8")
	assert.False(t, metrics.Timestamp.IsZero(), "Timestamp should be set")
}

func TestMetricsCollector_CollectSystemMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	collector := NewMetricsCollector(client, nil)

	// Collect system metrics
	metrics, err := collector.CollectSystemMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify metrics
	assert.True(t, metrics.UptimeSeconds >= 0, "Uptime should be non-negative")
	assert.True(t, metrics.MemoryUsageMB > 0, "Memory usage should be positive")
	assert.True(t, metrics.CPUUsagePercent >= 0, "CPU usage should be non-negative")
	assert.True(t, metrics.DiskUsagePercent >= 0, "Disk usage should be non-negative")
	assert.True(t, metrics.Goroutines > 0, "Goroutines should be positive")
	assert.False(t, metrics.Timestamp.IsZero(), "Timestamp should be set")
}

func TestMetricsCollector_CollectAuthMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	collector := NewMetricsCollector(client, nil)

	// Collect auth metrics (stub implementation)
	metrics, err := collector.CollectAuthMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify stub returns zeros
	assert.Equal(t, int64(0), metrics.LoginAttemptsTotal)
	assert.Equal(t, float64(0), metrics.LoginSuccessRate)
	assert.False(t, metrics.Timestamp.IsZero(), "Timestamp should be set")
}

func TestMetricsCollector_CollectPerformanceMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	collector := NewMetricsCollector(client, nil)

	// Collect performance metrics (stub implementation)
	metrics, err := collector.CollectPerformanceMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify stub returns zeros
	assert.Equal(t, int64(0), metrics.APIRequestsTotal)
	assert.Equal(t, 0, metrics.APIResponseTimeP95)
	assert.False(t, metrics.Timestamp.IsZero(), "Timestamp should be set")
}

func TestMetricsCollector_CollectAllMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	ctx := context.Background()
	collector := NewMetricsCollector(client, nil)

	// Create some test data
	_, err := client.User.Create().
		SetEmail("admin@example.com").
		SetPasswordHash("hash").
		SetRole("platform_admin").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// Collect all metrics
	metrics, err := collector.CollectAllMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify aggregated metrics
	assert.Equal(t, 1, metrics.TotalUsers)
	assert.Equal(t, 1, metrics.ActiveUsers)
	assert.Equal(t, 1, metrics.AdminUsers)
	assert.True(t, metrics.UptimeSeconds >= 0)
	assert.False(t, metrics.Timestamp.IsZero())
}
