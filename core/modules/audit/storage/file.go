package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStorage implements Storage interface using JSON file logging.
// This is a fallback storage when database is unavailable.
type FileStorage struct {
	logPath string
	mu      sync.Mutex
	file    *os.File
}

// NewFileStorage creates a new file storage instance.
func NewFileStorage(logPath string) (*FileStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileStorage{
		logPath: logPath,
		file:    file,
	}, nil
}

// Write appends an audit log entry to the file.
func (s *FileStorage) Write(ctx context.Context, entry *AuditEntry) error {
	if entry == nil {
		return fmt.Errorf("audit entry cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Write JSON line to file
	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// Sync to disk for durability
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync log file: %w", err)
	}

	return nil
}

// Query is not supported for file storage.
// Returns an error indicating database storage should be used for queries.
//
//nolint:gocritic // hugeParam: AuditFilter is acceptable at 80 bytes
func (s *FileStorage) Query(ctx context.Context, filter AuditFilter) (*AuditQueryResult, error) {
	return nil, fmt.Errorf("query not supported for file storage - use database storage for queries")
}

// Close closes the log file.
func (s *FileStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Rotate creates a new log file with timestamp suffix (for log rotation).
func (s *FileStorage) Rotate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close current file
	if s.file != nil {
		if err := s.file.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := fmt.Sprintf("%s.%s", s.logPath, timestamp)
	if err := os.Rename(s.logPath, rotatedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Open new file
	file, err := os.OpenFile(s.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	s.file = file
	return nil
}
