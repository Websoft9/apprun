package rbac

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitEnforcer(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")

	// Write minimal test policy
	policyContent := `p, owner, *, *
p, viewer, config, read`

	err := os.WriteFile(policyPath, []byte(policyContent), 0644)
	require.NoError(t, err)

	// Test initialization (model is now embedded)
	cfg := Config{
		PolicyPath: policyPath,
	}

	err = InitEnforcer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, GetEnforcer())
}

func TestCheckPermission(t *testing.T) {
	// Setup test enforcer
	tmpDir := t.TempDir()
	setupTestEnforcer(t, tmpDir)

	// Add test roles
	err := AddUserRole(1, 1, "owner")
	require.NoError(t, err)
	err = AddUserRole(2, 1, "viewer")
	require.NoError(t, err)

	tests := []struct {
		name      string
		userID    int64
		projectID int64
		resource  string
		action    string
		want      bool
	}{
		{
			name:      "Owner has all permissions",
			userID:    1,
			projectID: 1,
			resource:  "config",
			action:    "create",
			want:      true,
		},
		{
			name:      "Viewer can read",
			userID:    2,
			projectID: 1,
			resource:  "config",
			action:    "read",
			want:      true,
		},
		{
			name:      "Viewer cannot create",
			userID:    2,
			projectID: 1,
			resource:  "config",
			action:    "create",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := CheckPermission(tt.userID, tt.projectID, tt.resource, tt.action)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, allowed)
		})
	}
}

func setupTestEnforcer(t *testing.T, tmpDir string) {
	policyPath := filepath.Join(tmpDir, "policy.csv")

	policyContent := `p, owner, *, *
p, viewer, config, read`

	err := os.WriteFile(policyPath, []byte(policyContent), 0644)
	require.NoError(t, err)

	// Reset and initialize enforcer for test (model is now embedded)
	resetEnforcerForTest()

	cfg := Config{
		PolicyPath: policyPath,
	}

	err = InitEnforcer(cfg)
	require.NoError(t, err)
	require.NotNil(t, GetEnforcer())
}
