package storage

import (
	"context"
	"fmt"

	"apprun/ent"
	"apprun/ent/auditlog"

	"github.com/google/uuid"
)

// DatabaseStorage implements Storage interface using Ent ORM.
type DatabaseStorage struct {
	client *ent.Client
}

// NewDatabaseStorage creates a new database storage instance.
func NewDatabaseStorage(client *ent.Client) *DatabaseStorage {
	return &DatabaseStorage{
		client: client,
	}
}

// Write stores a single audit log entry to the database.
func (s *DatabaseStorage) Write(ctx context.Context, entry *AuditEntry) error {
	if entry == nil {
		return fmt.Errorf("audit entry cannot be nil")
	}

	// Generate ID if not set
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	// Build the create query
	builder := s.client.AuditLog.Create().
		SetID(entry.ID).
		SetTimestamp(entry.Timestamp).
		SetAction(entry.Action)

	// Set optional fields
	if entry.OperatorID != nil {
		builder.SetOperatorID(*entry.OperatorID)
	}
	if entry.TargetID != "" {
		builder.SetTargetID(entry.TargetID)
	}
	if entry.TargetType != "" {
		builder.SetTargetType(entry.TargetType)
	}
	if len(entry.Changes) > 0 {
		builder.SetChanges(entry.Changes)
	}
	if entry.IPAddress != "" {
		builder.SetIPAddress(entry.IPAddress)
	}
	if entry.UserAgent != "" {
		builder.SetUserAgent(entry.UserAgent)
	}
	if entry.StatusCode != 0 {
		builder.SetStatusCode(entry.StatusCode)
	}
	if entry.ResponseTimeMs != 0 {
		builder.SetResponseTimeMs(entry.ResponseTimeMs)
	}
	if entry.Method != "" {
		builder.SetMethod(entry.Method)
	}
	if entry.Path != "" {
		builder.SetPath(entry.Path)
	}

	// Execute the insert
	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// Query retrieves audit logs based on the provided filter.
//
//nolint:gocritic // hugeParam: AuditFilter is acceptable at 80 bytes
func (s *DatabaseStorage) Query(ctx context.Context, filter AuditFilter) (*AuditQueryResult, error) {
	// Build the base query
	query := s.client.AuditLog.Query()

	// Apply filters
	if filter.StartTime != nil {
		query = query.Where(auditlog.TimestampGTE(*filter.StartTime))
	}
	if filter.EndTime != nil {
		query = query.Where(auditlog.TimestampLTE(*filter.EndTime))
	}
	if filter.OperatorID != nil {
		query = query.Where(auditlog.OperatorIDEQ(*filter.OperatorID))
	}
	if filter.Action != "" {
		query = query.Where(auditlog.ActionEQ(filter.Action))
	}
	if filter.TargetType != "" {
		query = query.Where(auditlog.TargetTypeEQ(filter.TargetType))
	}

	// Order by timestamp descending (most recent first)
	query = query.Order(ent.Desc(auditlog.FieldTimestamp))

	// Count total if requested
	var total int64
	if filter.IncludeTotal {
		count, err := query.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count audit logs: %w", err)
		}
		total = int64(count)
	}

	// Apply pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 50
	}
	if filter.PageSize > 200 {
		filter.PageSize = 200
	}

	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// Execute the query
	logs, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}

	// Convert Ent models to AuditEntry
	entries := make([]*AuditEntry, 0, len(logs))
	for _, log := range logs {
		entry := &AuditEntry{
			ID:             log.ID,
			Timestamp:      log.Timestamp,
			Action:         log.Action,
			TargetID:       log.TargetID,
			TargetType:     log.TargetType,
			Changes:        log.Changes,
			IPAddress:      log.IPAddress,
			UserAgent:      log.UserAgent,
			StatusCode:     log.StatusCode,
			ResponseTimeMs: log.ResponseTimeMs,
			Method:         log.Method,
			Path:           log.Path,
		}
		if log.OperatorID != uuid.Nil {
			opID := log.OperatorID
			entry.OperatorID = &opID
		}
		entries = append(entries, entry)
	}

	return &AuditQueryResult{
		Logs:  entries,
		Total: total,
	}, nil
}

// Close closes the database connection.
func (s *DatabaseStorage) Close() error {
	// Ent client is managed externally, no-op here
	return nil
}
