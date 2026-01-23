package testutils

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Metrics Test Helpers for Story 9.1
// These helpers support ATDD tests for metrics exposure

// RandomMetricName generates a random metric name for testing
func RandomMetricName() string {
	prefixes := []string{"user", "system", "api", "db"}
	suffixes := []string{"count", "latency", "rate", "percent", "total"}

	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]

	return prefix + "_" + suffix
}

// RandomMetricValue generates a realistic metric value
func RandomMetricValue() float64 {
	return rand.Float64() * 1000
}

// RandomTags generates random metric tags for testing
func RandomTags() map[string]string {
	envs := []string{"production", "staging", "development"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}

	return map[string]string{
		"env":      envs[rand.Intn(len(envs))],
		"instance": fmt.Sprintf("apprun-%04d", rand.Intn(10000)),
		"region":   regions[rand.Intn(len(regions))],
	}
}

// MetricDataPoint represents a test metric data point
type MetricDataPoint struct {
	Name      string
	Value     float64
	Timestamp time.Time
	Tags      map[string]string
}

// GenerateMetricTimeSeries creates a time series of metric data points
// Useful for testing historical queries
func GenerateMetricTimeSeries(name string, count int, intervalMinutes int) []MetricDataPoint {
	series := make([]MetricDataPoint, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		series[i] = MetricDataPoint{
			Name:      name,
			Value:     RandomMetricValue(),
			Timestamp: now.Add(-time.Duration((count-i-1)*intervalMinutes) * time.Minute),
			Tags: map[string]string{
				"env":      "production",
				"instance": "apprun-01",
			},
		}
	}

	return series
}

// StandardUserMetrics returns standard user metric values for testing
func StandardUserMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_users":                    100,
		"active_users":                   85,
		"admin_users":                    5,
		"banned_users":                   2,
		"new_users_today":                10,
		"user_registrations_last_7_days": 45,
	}
}

// StandardSystemMetrics returns standard system metric values for testing
func StandardSystemMetrics() map[string]interface{} {
	return map[string]interface{}{
		"uptime_seconds":     86400,
		"memory_usage_mb":    512,
		"cpu_usage_percent":  35.2,
		"disk_usage_percent": 42.8,
		"goroutines":         245,
	}
}

// CreateMetricTags creates consistent tags for metrics
func CreateMetricTags(env, instance, source string) map[string]string {
	tags := map[string]string{
		"env":      env,
		"instance": instance,
	}

	if source != "" {
		tags["source"] = source
	}

	return tags
}

// AssertMetricTagsComplete verifies all required tags are present
func AssertMetricTagsComplete(t *testing.T, tags map[string]string) {
	t.Helper()

	requiredTags := []string{"env", "instance"}
	for _, tag := range requiredTags {
		if _, ok := tags[tag]; !ok {
			t.Errorf("Required tag missing: %s", tag)
		}
	}
}
