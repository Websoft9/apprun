package middleware_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"apprun/ent/enttest"
	"apprun/internal/rbac"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStory554Policies verifies Story 5.5.4 platform admin policies are loaded
func TestStory554Policies(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Initialize RBAC
	err := rbac.InitEnforcer(rbac.Config{EntClient: client})
	require.NoError(t, err)

	err = rbac.SeedDefaultPolicies()
	require.NoError(t, err)

	enforcer := rbac.GetEnforcer()

	t.Run("platform:user policy exists", func(t *testing.T) {
		policies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:user")
		assert.NotEmpty(t, policies)
		assert.Equal(t, []string{"platform_admin", "platform:user", "manage"}, policies[0])
		t.Log("✅ platform:user:manage policy loaded")
	})

	t.Run("platform:audit policy exists", func(t *testing.T) {
		policies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:audit")
		assert.NotEmpty(t, policies)
		assert.Equal(t, []string{"platform_admin", "platform:audit", "read"}, policies[0])
		t.Log("✅ platform:audit:read policy loaded")
	})

	t.Run("platform:rbac policy exists", func(t *testing.T) {
		policies := enforcer.GetFilteredPolicy(0, "platform_admin", "platform:rbac")
		assert.NotEmpty(t, policies)
		assert.Equal(t, []string{"platform_admin", "platform:rbac", "manage"}, policies[0])
		t.Log("✅ platform:rbac:manage policy loaded")
	})

	t.Run("wildcard policy preserved", func(t *testing.T) {
		policies := enforcer.GetFilteredPolicy(0, "platform_admin", "*")
		assert.NotEmpty(t, policies)
		assert.Equal(t, []string{"platform_admin", "*", "*"}, policies[0])
		t.Log("✅ Wildcard policy preserved for backward compatibility")
	})
}

// TestStory554PermissionEnforcement verifies permission checks work correctly
func TestStory554PermissionEnforcement(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Initialize RBAC
	err := rbac.InitEnforcer(rbac.Config{EntClient: client})
	require.NoError(t, err)

	err = rbac.SeedDefaultPolicies()
	require.NoError(t, err)

	// Create admin user
	adminUser, err := client.User.Create().
		SetEmail("admin@test.com").
		SetUsername("admin").
		SetPasswordHash("hashed").
		SetRole("platform_admin").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// Assign platform_admin role
	enforcer := rbac.GetEnforcer()
	_, err = enforcer.AddGroupingPolicy("u:"+fmt.Sprint(adminUser.ID), "platform_admin")
	require.NoError(t, err)

	t.Run("Admin has platform:user:manage", func(t *testing.T) {
		allowed, err := rbac.CheckPermission(adminUser.ID, 0, "platform:user", "manage")
		require.NoError(t, err)
		assert.True(t, allowed)
		t.Log("✅ Admin has platform:user:manage permission")
	})

	t.Run("Admin has platform:audit:read", func(t *testing.T) {
		allowed, err := rbac.CheckPermission(adminUser.ID, 0, "platform:audit", "read")
		require.NoError(t, err)
		assert.True(t, allowed)
		t.Log("✅ Admin has platform:audit:read permission")
	})

	t.Run("Admin has platform:rbac:manage", func(t *testing.T) {
		allowed, err := rbac.CheckPermission(adminUser.ID, 0, "platform:rbac", "manage")
		require.NoError(t, err)
		assert.True(t, allowed)
		t.Log("✅ Admin has platform:rbac:manage permission")
	})

	t.Run("Admin has wildcard access", func(t *testing.T) {
		allowed, err := rbac.CheckPermission(adminUser.ID, 0, "any_resource", "any_action")
		require.NoError(t, err)
		assert.True(t, allowed)
		t.Log("✅ Wildcard access works")
	})

	// Create regular user
	regularUser, err := client.User.Create().
		SetEmail("user@test.com").
		SetUsername("user").
		SetPasswordHash("hashed").
		SetRole("user").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	t.Run("Regular user does NOT have platform:user:manage", func(t *testing.T) {
		allowed, err := rbac.CheckPermission(regularUser.ID, 0, "platform:user", "manage")
		require.NoError(t, err)
		assert.False(t, allowed)
		t.Log("✅ Regular user correctly denied access")
	})
}

// TestStory554DatabaseSync verifies policies sync to database
func TestStory554DatabaseSync(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Initialize RBAC
	err := rbac.InitEnforcer(rbac.Config{EntClient: client})
	require.NoError(t, err)

	err = rbac.SeedDefaultPolicies()
	require.NoError(t, err)

	t.Run("Policies in database", func(t *testing.T) {
		count, err := client.CasbinRule.Query().Count(ctx)
		require.NoError(t, err)
		assert.Greater(t, count, 40, "should have loaded all policies including Story 5.5.4")
		t.Logf("✅ Database contains %d policy rules", count)
	})

	t.Run("Idempotent seeding", func(t *testing.T) {
		countBefore, _ := client.CasbinRule.Query().Count(ctx)

		// Seed again
		err := rbac.SeedDefaultPolicies()
		require.NoError(t, err)

		countAfter, _ := client.CasbinRule.Query().Count(ctx)
		assert.Equal(t, countBefore, countAfter)
		t.Log("✅ Policy seeding is idempotent")
	})
}
