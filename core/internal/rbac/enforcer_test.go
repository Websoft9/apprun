package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitEnforcerWithEmbeddedPolicy(t *testing.T) {
	// Reset enforcer before test
	resetEnforcerForTest()

	// Test initialization with embedded policy (no EntClient = in-memory mode)
	cfg := Config{}
	err := InitEnforcer(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, GetEnforcer())
	assert.False(t, IsUsingDatabaseStorage(), "Should not use database storage when EntClient is nil")

	// Verify embedded policies are loaded
	pCount, gCount, g2Count := GetPolicySummary()
	assert.Greater(t, pCount, 0, "Should have loaded policy rules from embedded CSV")
	t.Logf("Loaded policies: p=%d, g=%d, g2=%d", pCount, gCount, g2Count)
}

func TestCheckPermission(t *testing.T) {
	// Setup test enforcer with embedded policy
	setupTestEnforcerWithEmbeddedPolicy(t)

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

func TestSeedDefaultPoliciesIdempotent(t *testing.T) {
	// Setup test enforcer
	setupTestEnforcerWithEmbeddedPolicy(t)

	// Get initial policy count
	pCountBefore, _, _ := GetPolicySummary()

	// Call SeedDefaultPolicies multiple times (should be idempotent)
	err := SeedDefaultPolicies()
	assert.NoError(t, err)

	err = SeedDefaultPolicies()
	assert.NoError(t, err)

	// Policy count should remain the same
	pCountAfter, _, _ := GetPolicySummary()
	assert.Equal(t, pCountBefore, pCountAfter, "SeedDefaultPolicies should be idempotent")
}

func TestLoadEmbeddedPolicies(t *testing.T) {
	// Verify that embedded policy content can be parsed
	lines := splitLines(defaultPolicyContent)
	policyCount := 0

	for _, line := range lines {
		line = trimLine(line)
		if line == "" || line[0] == '#' {
			continue
		}

		ptype, rule, ok := parseCSVLine(line)
		if ok {
			assert.NotEmpty(t, ptype, "ptype should not be empty")
			assert.NotEmpty(t, rule, "rule should not be empty")
			policyCount++
		}
	}

	assert.Greater(t, policyCount, 0, "Should have parsed policies from embedded CSV")
	t.Logf("Parsed %d policy lines from embedded CSV", policyCount)
}

func TestHelperFunctions(t *testing.T) {
	t.Run("splitLines", func(t *testing.T) {
		content := "line1\nline2\nline3"
		lines := splitLines(content)
		assert.Len(t, lines, 3)
		assert.Equal(t, "line1", lines[0])
		assert.Equal(t, "line2", lines[1])
		assert.Equal(t, "line3", lines[2])
	})

	t.Run("trimLine", func(t *testing.T) {
		assert.Equal(t, "test", trimLine("  test  "))
		assert.Equal(t, "test", trimLine("\ttest\t"))
		assert.Equal(t, "test", trimLine("  \t  test  \t  "))
	})

	t.Run("parseCSVLine", func(t *testing.T) {
		ptype, rule, ok := parseCSVLine("p, role, resource, action")
		assert.True(t, ok)
		assert.Equal(t, "p", ptype)
		assert.Equal(t, []string{"role", "resource", "action"}, rule)

		ptype, rule, ok = parseCSVLine("g, user, role, domain")
		assert.True(t, ok)
		assert.Equal(t, "g", ptype)
		assert.Equal(t, []string{"user", "role", "domain"}, rule)

		_, _, ok = parseCSVLine("single")
		assert.False(t, ok)
	})
}

func setupTestEnforcerWithEmbeddedPolicy(t *testing.T) {
	// Reset and initialize enforcer with embedded policy (in-memory mode)
	resetEnforcerForTest()

	cfg := Config{} // No EntClient = in-memory mode with embedded policies
	err := InitEnforcer(cfg)
	require.NoError(t, err)
	require.NotNil(t, GetEnforcer())
}
