package obs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMetricName(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid: starts with letter, alphanumeric",
			metricName:  "testMetric123",
			expectError: false,
		},
		{
			name:        "valid: with underscores",
			metricName:  "test_metric_name",
			expectError: false,
		},
		{
			name:        "valid: with hyphens",
			metricName:  "test-metric-name",
			expectError: false,
		},
		{
			name:        "valid: with dots",
			metricName:  "test.metric.name",
			expectError: false,
		},
		{
			name:        "valid: complex name",
			metricName:  "http_requests_total-count.value",
			expectError: false,
		},
		{
			name:        "invalid: starts with number",
			metricName:  "123metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: contains colon (injection risk)",
			metricName:  "test:metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: contains special chars @",
			metricName:  "test@metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: contains special chars $",
			metricName:  "test$metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: contains space",
			metricName:  "test metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: contains newline",
			metricName:  "test\nmetric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: starts with underscore",
			metricName:  "_test_metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: starts with hyphen",
			metricName:  "-test-metric",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
		{
			name:        "invalid: empty string",
			metricName:  "",
			expectError: true,
			errorMsg:    "invalid metric name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetricName(tt.metricName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
