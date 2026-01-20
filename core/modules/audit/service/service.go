// Package service provides audit logging service with async processing.
package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"apprun/modules/audit"
	"apprun/modules/audit/storage"

	"github.com/google/uuid"
)

// Service handles audit logging with async processing and graceful shutdown.
type Service struct {
	storage         storage.Storage
	fallbackStorage storage.Storage
	config          audit.ServiceConfig

	queue           chan *storage.AuditEntry
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	started         bool
	mu              sync.Mutex
	sensitiveFields []string
}

// NewService creates a new audit service instance.
func NewService(primaryStorage storage.Storage, config audit.ServiceConfig, sensitiveFields []string) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize fallback storage if configured
	var fallbackStorage storage.Storage
	if config.FallbackFilePath != "" {
		fileStorage, err := storage.NewFileStorage(config.FallbackFilePath)
		if err != nil {
			log.Printf("[WARN] Failed to initialize fallback file storage: %v", err)
		} else {
			fallbackStorage = fileStorage
		}
	}

	svc := &Service{
		storage:         primaryStorage,
		fallbackStorage: fallbackStorage,
		config:          config,
		queue:           make(chan *storage.AuditEntry, config.BufferSize),
		ctx:             ctx,
		cancel:          cancel,
		sensitiveFields: sensitiveFields,
	}

	return svc, nil
}

// Start starts the worker pool for processing audit logs.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("audit service already started")
	}

	if !s.config.Enabled {
		log.Println("[INFO] Audit service is disabled")
		return nil
	}

	// Start worker goroutines
	for i := 0; i < s.config.WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	s.started = true
	log.Printf("[INFO] Audit service started with %d workers", s.config.WorkerCount)
	return nil
}

// worker processes audit log entries from the queue.
func (s *Service) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			// Drain remaining items from queue before exiting
			for {
				select {
				case entry := <-s.queue:
					s.writeWithFallback(s.ctx, entry)
				default:
					return
				}
			}
		case entry := <-s.queue:
			s.writeWithFallback(s.ctx, entry)
		}
	}
}

// writeWithFallback writes to primary storage with fallback on failure.
func (s *Service) writeWithFallback(ctx context.Context, entry *storage.AuditEntry) {
	// Try primary storage
	if err := s.storage.Write(ctx, entry); err != nil {
		log.Printf("[ERROR] Failed to write audit log to primary storage: %v", err)

		// Try fallback storage
		if s.fallbackStorage != nil {
			if err := s.fallbackStorage.Write(ctx, entry); err != nil {
				log.Printf("[ERROR] Failed to write audit log to fallback storage: %v", err)
			} else {
				log.Printf("[WARN] Audit log written to fallback storage (primary failed)")
			}
		}
	}
}

// Log records an audit log entry (used by middleware).
func (s *Service) Log(ctx context.Context, entry *storage.AuditEntry) error {
	if !s.config.Enabled {
		return nil
	}

	// Generate ID if not set
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	// Set timestamp if not set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Sanitize sensitive fields
	s.sanitize(entry)

	// Non-blocking send to queue
	select {
	case s.queue <- entry:
		return nil
	default:
		// Queue is full - apply backpressure
		return fmt.Errorf("audit queue full (buffer size: %d)", s.config.BufferSize)
	}
}

// LogActionParams holds parameters for manual audit logging.
type LogActionParams struct {
	Action     string                 `json:"action"`
	TargetID   string                 `json:"target_id,omitempty"`
	TargetType string                 `json:"target_type,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
}

// LogAction records a manual audit log entry (used by business modules).
func (s *Service) LogAction(ctx context.Context, params LogActionParams) error {
	if !s.config.Enabled {
		return nil
	}

	// Extract operator and IP from context
	operatorID := GetOperatorIDFromContext(ctx)
	ipAddress := GetIPAddressFromContext(ctx)

	entry := &storage.AuditEntry{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		OperatorID: operatorID,
		Action:     params.Action,
		TargetID:   params.TargetID,
		TargetType: params.TargetType,
		Changes:    params.Changes,
		IPAddress:  ipAddress,
	}

	return s.Log(ctx, entry)
}

// Query retrieves audit logs based on filter.
//
//nolint:gocritic // hugeParam: AuditFilter is acceptable at 80 bytes
func (s *Service) Query(ctx context.Context, filter storage.AuditFilter) (*storage.AuditQueryResult, error) {
	return s.storage.Query(ctx, filter)
}

// Shutdown gracefully shuts down the audit service.
// Waits for all queued logs to be processed before returning.
func (s *Service) Shutdown(timeout time.Duration) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	log.Println("[INFO] Shutting down audit service...")

	// Signal workers to stop accepting new work
	s.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[INFO] Audit service shutdown complete")
	case <-time.After(timeout):
		log.Println("[WARN] Audit service shutdown timeout - some logs may be lost")
	}

	// Close storage connections
	if err := s.storage.Close(); err != nil {
		log.Printf("[ERROR] Failed to close primary storage: %v", err)
	}
	if s.fallbackStorage != nil {
		if err := s.fallbackStorage.Close(); err != nil {
			log.Printf("[ERROR] Failed to close fallback storage: %v", err)
		}
	}

	return nil
}

// Context helper functions

type contextKey string

const (
	contextKeyOperatorID contextKey = "audit_operator_id"
	contextKeyIPAddress  contextKey = "audit_ip_address"
)

// SetOperatorIDInContext stores operator ID in context.
func SetOperatorIDInContext(ctx context.Context, operatorID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyOperatorID, operatorID)
}

// GetOperatorIDFromContext retrieves operator ID from context.
func GetOperatorIDFromContext(ctx context.Context) *uuid.UUID {
	if id, ok := ctx.Value(contextKeyOperatorID).(uuid.UUID); ok {
		return &id
	}
	return nil
}

// SetIPAddressInContext stores IP address in context.
func SetIPAddressInContext(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, contextKeyIPAddress, ipAddress)
}

// GetIPAddressFromContext retrieves IP address from context.
func GetIPAddressFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(contextKeyIPAddress).(string); ok {
		return ip
	}
	return ""
}

// sanitize removes sensitive information from the audit entry
func (s *Service) sanitize(entry *storage.AuditEntry) {
	if len(s.sensitiveFields) == 0 {
		return
	}

	if entry.Changes != nil {
		s.maskMap(entry.Changes)
	}
}

// maskMap recursively masks sensitive fields in a map
func (s *Service) maskMap(data map[string]interface{}) {
	for k, v := range data {
		// Check if key is sensitive
		if s.isSensitive(k) {
			data[k] = "******"
			continue
		}

		// Recursively mask maps and slices
		switch val := v.(type) {
		case map[string]interface{}:
			s.maskMap(val)
		case []interface{}:
			for _, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					s.maskMap(m)
				}
			}
		}
	}
}

// isSensitive checks if a field name is sensitive
func (s *Service) isSensitive(field string) bool {
	for _, sensitive := range s.sensitiveFields {
		if field == sensitive {
			return true
		}
	}
	return false
}
