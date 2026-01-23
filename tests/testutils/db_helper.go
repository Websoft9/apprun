package testutils

import (
"database/sql"
"fmt"
"os"
"testing"

"github.com/stretchr/testify/require"
_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
// Returns *sql.DB for basic connectivity tests
func SetupTestDB(t *testing.T) *sql.DB {
t.Helper()

// Get test database URL from environment
dbURL := os.Getenv("TEST_DATABASE_URL")
if dbURL == "" {
dbURL = "postgresql://postgres:postgres@localhost:5432/apprun_test?sslmode=disable"
}

db, err := sql.Open("postgres", dbURL)
require.NoError(t, err, "failed to connect to test database")

err = db.Ping()
require.NoError(t, err, "failed to ping test database")

t.Cleanup(func() {
db.Close()
})

return db
}

// CreateTestDatabase creates a new test database (for parallel testing)
func CreateTestDatabase(dbName string) error {
adminURL := "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"

db, err := sql.Open("postgres", adminURL)
if err != nil {
return fmt.Errorf("failed to connect to postgres: %w", err)
}
defer db.Close()

// Drop existing test database
_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName))

// Create new test database
_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
if err != nil {
return fmt.Errorf("failed to create test database: %w", err)
}

return nil
}

// DropTestDatabase removes test database
func DropTestDatabase(dbName string) error {
adminURL := "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"

db, err := sql.Open("postgres", adminURL)
if err != nil {
return fmt.Errorf("failed to connect to postgres: %w", err)
}
defer db.Close()

_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName))
if err != nil {
return fmt.Errorf("failed to drop test database: %w", err)
}

return nil
}
