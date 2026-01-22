package metrics

import (
	"context"
	"testing"

	"apprun/ent/enttest"
	"apprun/internal/rbac"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsPermission_PlatformAdmin verifies platform_admin has platform:metrics:read permission
func TestMetricsPermission_PlatformAdmin(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Initialize RBAC
	err := rbac.InitEnforcer(rbac.Config{EntClient: client})
	require.NoError(t, err, "Failed to initialize RBAC enforcer")

	// Seed default policies from embedded CSV
	err = rbac.SeedDefaultPolicies()
	require.NoError(t, err, "Failed to seed default policies")

	enforcer := rbac.GetEnforcer()
	require.NotNil(t, enforcer, "RBAC enforcer should not be nil")

	t.Run("platform:metrics policy exists", func(t *testing.T) {
		policies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:metrics")
		require.NotEmpty(t, policies, "platform:metrics policy should exist for platform_admin")
		assert.Equal(t, []string{"platform_admin", "platform:metrics", "read"}, policies[0])
		t.Log("✅ platform:metrics:read policy loaded")
	})

	t.Run("platform_admin role has platform:metrics:read", func(t *testing.T) {
		// Check using Casbin enforcer directly
		allowed, err := enforcer.Enforce("platform_admin", "platform", "platform:metrics", "read")
		require.NoError(t, err, "Permission check should not error")
		assert.True(t, allowed, "platform_admin should have platform:metrics:read permission")
		t.Log("✅ platform_admin has platform:metrics:read permission")
	})

	t.Run("Admin user has platform:metrics:read", func(t *testing.T) {
		// Create admin user
		adminUser, err := client.User.Create().
			SetEmail("admin@test.com").
			SetPasswordHash("hash").
			SetRole("platform_admin").
			SetStatus(1).
			Save(ctx)
		require.NoError(t, err)

		// Assign platform_admin role
		_, err = enforcer.AddGroupingPolicy(rbac.FormatUserKey(adminUser.ID), "platform_admin", rbac.FormatDomain(0))
		require.NoError(t, err)

		// Check permission for actual user
		allowed, err := rbac.CheckPermission(adminUser.ID, 0, "platform:metrics", "read")
		require.NoError(t, err)
		assert.True(t, allowed, "Admin user should have platform:metrics:read permission")
		t.Log("✅ Admin user has platform:metrics:read permission")
	})

	t.Run("Non-admin user does NOT have platform:metrics:read", func(t *testing.T) {
		// Create regular user
		regularUser, err := client.User.Create().
			SetEmail("user@test.com").
			SetPasswordHash("hash").
			SetRole("platform_user").
			SetStatus(1).
			Save(ctx)
		require.NoError(t, err)

		// Check permission for regular user
		allowed, err := rbac.CheckPermission(regularUser.ID, 0, "platform:metrics", "read")
		require.NoError(t, err)
		assert.False(t, allowed, "Regular user should NOT have platform:metrics:read permission")
		t.Log("✅ Regular user correctly denied platform:metrics:read")
	})
}

// TestMetricsPermission_Separation verifies metrics and audit permissions are independent
func TestMetricsPermission_Separation(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Initialize RBAC
	err := rbac.InitEnforcer(rbac.Config{EntClient: client})
	require.NoError(t, err)

	// Seed default policies from embedded CSV
	err = rbac.SeedDefaultPolicies()
	require.NoError(t, err)

	enforcer := rbac.GetEnforcer()

	t.Run("platform:metrics and platform:audit are separate policies", func(t *testing.T) {
		metricsPolicies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:metrics")
		auditPolicies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:audit")

		require.NotEmpty(t, metricsPolicies, "platform:metrics policy should exist")
		require.NotEmpty(t, auditPolicies, "platform:audit policy should exist")

		assert.NotEqual(t, metricsPolicies, auditPolicies, "Metrics and audit should be separate policies")
		t.Log("✅ Metrics and audit permissions are properly separated")
	})
}
