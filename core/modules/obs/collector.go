// Package obs provides observability features including metrics collection and exposure.
package obs

import (
	"context"
	"runtime"
	"time"

	"apprun/ent"
	"apprun/ent/user"
	"apprun/pkg/errors"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
)

var startTime = time.Now()

// MetricsCollector collects various system and application metrics
type MetricsCollector struct {
	entClient *ent.Client
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector(client *ent.Client) *MetricsCollector {
	return &MetricsCollector{
		entClient: client,
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

	return &UserMetrics{
		TotalUsers:                 totalUsers,
		ActiveUsers:                activeUsers,
		AdminUsers:                 adminUsers,
		BannedUsers:                bannedUsers,
		NewUsersToday:              newUsersToday,
		UserRegistrationsLast7Days: registrationsLast7Days,
		Timestamp:                  time.Now(),
	}, nil
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

	return &SystemMetrics{
		UptimeSeconds:    uptime,
		MemoryUsageMB:    memoryUsageMB,
		CPUUsagePercent:  cpuPercent[0],
		DiskUsagePercent: diskUsagePercent,
		Goroutines:       goroutines,
		Timestamp:        time.Now(),
	}, nil
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
