// Package metrics provides observability features including metrics collection and exposure.
package metrics

import (
	"context"
	"runtime"
	"time"

	"apprun/ent"
	"apprun/ent/user"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/metricstore"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
)

var startTime = time.Now()

// MetricsCollector collects various system and application metrics
type MetricsCollector struct {
	entClient *ent.Client
	repo      *metricstore.Repository // Story 9.1: Storage integration
	logger    logger.Logger
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector(client *ent.Client, repo *metricstore.Repository) *MetricsCollector {
	return &MetricsCollector{
		entClient: client,
		repo:      repo,
		logger:    logger.L(),
	}
}

// CollectUserMetrics collects user-related metrics
func (c *MetricsCollector) CollectUserMetrics(ctx context.Context) (*UserMetrics, error) {
	// Query 1: Total users
	totalUsers, err := c.entClient.User.Query().Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count total users")
	}

	// Query 2: Active users (status = 1)
	activeUsers, err := c.entClient.User.Query().
		Where(user.StatusEQ(1)).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count active users")
	}

	// Query 3: Admin users (role = platform_admin)
	adminUsers, err := c.entClient.User.Query().
		Where(user.RoleEQ("platform_admin")).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count admin users")
	}

	// Query 4: Banned users (status = 0)
	bannedUsers, err := c.entClient.User.Query().
		Where(user.StatusEQ(0)).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count banned users")
	}

	// Query 5: New users today (created_at >= today 00:00:00)
	today := time.Now().Truncate(24 * time.Hour)
	newUsersToday, err := c.entClient.User.Query().
		Where(user.CreatedAtGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count new users today")
	}

	// Query 6: Registrations last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	registrationsLast7Days, err := c.entClient.User.Query().
		Where(user.CreatedAtGTE(sevenDaysAgo)).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to count user registrations last 7 days")
	}

	metrics := &UserMetrics{
		TotalUsers:                 totalUsers,
		ActiveUsers:                activeUsers,
		AdminUsers:                 adminUsers,
		BannedUsers:                bannedUsers,
		NewUsersToday:              newUsersToday,
		UserRegistrationsLast7Days: registrationsLast7Days,
		Timestamp:                  time.Now(),
	}

	// Story 9.1: Async persist to storage (non-blocking)
	if c.repo != nil {
		go c.persistUserMetrics(context.Background(), metrics)
	}

	return metrics, nil
}

// CollectSystemMetrics collects system health metrics
func (c *MetricsCollector) CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	// Uptime
	uptime := int64(time.Since(startTime).Seconds())

	// Memory (Go runtime)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryUsageMB := memStats.Alloc / 1024 / 1024

	// CPU Usage (gopsutil)
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		// Non-critical error, log and use default value
		cpuPercent = []float64{0}
	}
	if len(cpuPercent) == 0 {
		cpuPercent = []float64{0}
	}

	// Disk Usage (gopsutil)
	diskStat, err := disk.Usage("/")
	diskUsagePercent := 0.0
	if err == nil {
		diskUsagePercent = diskStat.UsedPercent
	}
	// Non-critical error, continue with default value 0.0

	// Goroutines
	goroutines := runtime.NumGoroutine()

	metrics := &SystemMetrics{
		UptimeSeconds:    uptime,
		MemoryUsageMB:    memoryUsageMB,
		CPUUsagePercent:  cpuPercent[0],
		DiskUsagePercent: diskUsagePercent,
		Goroutines:       goroutines,
		Timestamp:        time.Now(),
	}

	// Story 9.1: Async persist to storage (non-blocking)
	if c.repo != nil {
		go c.persistSystemMetrics(context.Background(), metrics)
	}

	return metrics, nil
}

// CollectAuthMetrics collects authentication metrics
// Note: Returns zero values if audit_logs not available
func (c *MetricsCollector) CollectAuthMetrics(ctx context.Context) (*AuthMetrics, error) {
	// Phase 1: Return stub data (audit logs not yet implemented)
	// TODO: Implement after Story 5.9 (Audit Logging) is completed
	return &AuthMetrics{
		LoginAttemptsTotal:  0,
		LoginSuccessRate:    0,
		FailedLoginAttempts: 0,
		TokenIssuedTotal:    0,
		Timestamp:           time.Now(),
	}, nil
}

// CollectPerformanceMetrics collects API performance metrics
// Note: Returns stub data (requires middleware instrumentation)
func (c *MetricsCollector) CollectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	// Phase 1: Return stub data
	// TODO: Implement middleware to collect real performance data
	return &PerformanceMetrics{
		APIRequestsTotal:         0,
		APIResponseTimeP95:       0,
		APIErrorRate:             0,
		DatabaseQueryDurationAvg: 0,
		Timestamp:                time.Now(),
	}, nil
}

// CollectAllMetrics collects all metrics at once
func (c *MetricsCollector) CollectAllMetrics(ctx context.Context) (*AllMetrics, error) {
	userMetrics, err := c.CollectUserMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect user metrics")
	}

	systemMetrics, err := c.CollectSystemMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect system metrics")
	}

	authMetrics, err := c.CollectAuthMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect auth metrics")
	}

	perfMetrics, err := c.CollectPerformanceMetrics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "failed to collect performance metrics")
	}

	return &AllMetrics{
		TotalUsers:                 userMetrics.TotalUsers,
		ActiveUsers:                userMetrics.ActiveUsers,
		AdminUsers:                 userMetrics.AdminUsers,
		BannedUsers:                userMetrics.BannedUsers,
		NewUsersToday:              userMetrics.NewUsersToday,
		UserRegistrationsLast7Days: userMetrics.UserRegistrationsLast7Days,
		LoginAttemptsTotal:         authMetrics.LoginAttemptsTotal,
		LoginSuccessRate:           authMetrics.LoginSuccessRate,
		FailedLoginAttempts:        authMetrics.FailedLoginAttempts,
		TokenIssuedTotal:           authMetrics.TokenIssuedTotal,
		APIRequestsTotal:           perfMetrics.APIRequestsTotal,
		APIResponseTimeP95:         perfMetrics.APIResponseTimeP95,
		APIErrorRate:               perfMetrics.APIErrorRate,
		UptimeSeconds:              systemMetrics.UptimeSeconds,
		MemoryUsageMB:              systemMetrics.MemoryUsageMB,
		CPUUsagePercent:            systemMetrics.CPUUsagePercent,
		DiskUsagePercent:           systemMetrics.DiskUsagePercent,
		Timestamp:                  time.Now(),
	}, nil
}

// persistUserMetrics writes user metrics to storage asynchronously
// Errors are logged but don't affect the real-time response (degradation strategy)
func (c *MetricsCollector) persistUserMetrics(ctx context.Context, m *UserMetrics) {
	systemTags := map[string]string{"source": "system"}
	if err := c.repo.RecordMetric(ctx, MetricNameUserTotal, float64(m.TotalUsers), systemTags); err != nil {
		c.logger.Warn("Failed to persist user_count_total", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameUserActive, float64(m.ActiveUsers), systemTags); err != nil {
		c.logger.Warn("Failed to persist user_count_active", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameUserAdmin, float64(m.AdminUsers), systemTags); err != nil {
		c.logger.Warn("Failed to persist user_count_admin", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameUserBanned, float64(m.BannedUsers), systemTags); err != nil {
		c.logger.Warn("Failed to persist user_count_banned", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameUserNewToday, float64(m.NewUsersToday), systemTags); err != nil {
		c.logger.Warn("Failed to persist user_count_new_today", logger.Field{Key: "error", Value: err})
	}
}

// persistSystemMetrics writes system metrics to storage asynchronously
func (c *MetricsCollector) persistSystemMetrics(ctx context.Context, m *SystemMetrics) {
	systemTags := map[string]string{"source": "system"}
	if err := c.repo.RecordMetric(ctx, MetricNameSystemMemory, float64(m.MemoryUsageMB), systemTags); err != nil {
		c.logger.Warn("Failed to persist system_memory_mb", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameSystemCPU, m.CPUUsagePercent, systemTags); err != nil {
		c.logger.Warn("Failed to persist system_cpu_percent", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameSystemDisk, m.DiskUsagePercent, systemTags); err != nil {
		c.logger.Warn("Failed to persist system_disk_percent", logger.Field{Key: "error", Value: err})
	}
	if err := c.repo.RecordMetric(ctx, MetricNameSystemUptime, float64(m.UptimeSeconds), systemTags); err != nil {
		c.logger.Warn("Failed to persist system_uptime_seconds", logger.Field{Key: "error", Value: err})
	}
}
