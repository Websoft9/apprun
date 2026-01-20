// Package storage provides audit log storage implementations.
package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	ID             uuid.UUID              `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	OperatorID     *uuid.UUID             `json:"operator_id,omitempty"`
	Action         string                 `json:"action"`
	TargetID       string                 `json:"target_id,omitempty"`
	TargetType     string                 `json:"target_type,omitempty"`
	Changes        map[string]interface{} `json:"changes,omitempty"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	StatusCode     int                    `json:"status_code,omitempty"`
	ResponseTimeMs int                    `json:"response_time_ms,omitempty"`
	Method         string                 `json:"method,omitempty"`
	Path           string                 `json:"path,omitempty"`
}

// AuditFilter represents query filter for audit logs.
type AuditFilter struct {
	StartTime    *time.Time
	EndTime      *time.Time
	OperatorID   *uuid.UUID
	Action       string
	TargetType   string
	Page         int
	PageSize     int
	IncludeTotal bool
}

// AuditQueryResult represents the result of an audit query.
type AuditQueryResult struct {
	Logs  []*AuditEntry
	Total int64
}

// Storage defines the interface for audit log storage implementations.
type Storage interface {
	// Write stores a single audit log entry.
	Write(ctx context.Context, entry *AuditEntry) error

	// Query retrieves audit logs based on the provided filter.
	//nolint:gocritic // hugeParam: AuditFilter is acceptable at 80 bytes
	Query(ctx context.Context, filter AuditFilter) (*AuditQueryResult, error)

	// Close closes the storage connection and releases resources.
	Close() error
}
