package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMigration(t *testing.T) {
	// Create temp directory for migrations
	tempDir := t.TempDir()

	ctx := context.Background()

	// Test: Generate first migration
	path, err := GenerateMigration(ctx, "create_users", tempDir)
	require.NoError(t, err)
	assert.Contains(t, path, "001_create_users.sql")

	// Verify file exists
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Migration: create_users")
	assert.Contains(t, string(content), "Version: 001")

	// Test: Generate second migration
	path2, err := GenerateMigration(ctx, "add_email_column", tempDir)
	require.NoError(t, err)
	assert.Contains(t, path2, "002_add_email_column.sql")
}

func TestGenerateMigration_EmptyName(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	_, err := GenerateMigration(ctx, "", tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration name is required")
}

func TestNewMigrator(t *testing.T) {
	// Test with default directory
	m := NewMigrator(nil, "")
	assert.Equal(t, "migrations", m.migrationsDir)

	// Test with custom directory
	m2 := NewMigrator(nil, "custom/migrations")
	assert.Equal(t, "custom/migrations", m2.migrationsDir)
}

func TestMigrationStatus(t *testing.T) {
	// Create temp directory with test migrations
	tempDir := t.TempDir()

	// Create test migration file
	migrationFile := filepath.Join(tempDir, "001_test.sql")
	err := os.WriteFile(migrationFile, []byte("-- test migration"), 0644)
	require.NoError(t, err)

	m := NewMigrator(nil, tempDir)
	ctx := context.Background()

	status, err := m.Status(ctx)
	require.NoError(t, err)

	assert.False(t, status.IsCurrent)
	assert.Len(t, status.Pending, 1)
	assert.Equal(t, "001_test.sql", status.Pending[0])
}

func TestAutoMigrateIfEnabled_Disabled(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		Driver:      "postgres",
		Host:        "localhost",
		Port:        5432,
		User:        "test",
		Password:    "test12345678",
		DBName:      "test",
		AutoMigrate: false, // Disabled
	}

	// Should return nil immediately when disabled
	err := AutoMigrateIfEnabled(ctx, cfg)
	assert.NoError(t, err)
}

// Note: Integration tests requiring real database should be in a separate file
// with build tag: //go:build integration
