package rbac

import (
	"context"
	"testing"

	"apprun/ent/enttest"

	"github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewEntAdapter(t *testing.T) {
	t.Run("nil client returns error", func(t *testing.T) {
		adapter, err := NewEntAdapter(nil)
		assert.Error(t, err)
		assert.Nil(t, adapter)
	})

	t.Run("valid client creates adapter", func(t *testing.T) {
		client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
		defer client.Close()

		adapter, err := NewEntAdapter(client)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
	})
}

func TestEntAdapterCRUD(t *testing.T) {
	// Create in-memory SQLite client for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	adapter, err := NewEntAdapter(client)
	require.NoError(t, err)

	// Test AddPolicy
	t.Run("AddPolicy", func(t *testing.T) {
		err := adapter.AddPolicy("p", "p", []string{"admin", "data", "read"})
		assert.NoError(t, err)

		// Verify policy exists in database
		ctx := context.Background()
		rules, err := client.CasbinRule.Query().All(ctx)
		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "p", rules[0].Ptype)
		assert.Equal(t, "admin", rules[0].V0)
		assert.Equal(t, "data", rules[0].V1)
		assert.Equal(t, "read", rules[0].V2)
	})

	// Test RemovePolicy
	t.Run("RemovePolicy", func(t *testing.T) {
		err := adapter.RemovePolicy("p", "p", []string{"admin", "data", "read"})
		assert.NoError(t, err)

		// Verify policy removed from database
		ctx := context.Background()
		rules, err := client.CasbinRule.Query().All(ctx)
		assert.NoError(t, err)
		assert.Len(t, rules, 0)
	})

	// Test AddPolicy with grouping
	t.Run("AddGroupingPolicy", func(t *testing.T) {
		err := adapter.AddPolicy("g", "g", []string{"user:1", "admin", "project:1"})
		assert.NoError(t, err)

		ctx := context.Background()
		rules, err := client.CasbinRule.Query().All(ctx)
		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "g", rules[0].Ptype)
	})
}

func TestEntAdapterLoadPolicy(t *testing.T) {
	// Create in-memory SQLite client for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Pre-populate database with policies
	ctx := context.Background()
	_, err := client.CasbinRule.Create().
		SetPtype("p").
		SetV0("admin").
		SetV1("data").
		SetV2("*").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.CasbinRule.Create().
		SetPtype("p").
		SetV0("viewer").
		SetV1("config").
		SetV2("read").
		Save(ctx)
	require.NoError(t, err)

	adapter, err := NewEntAdapter(client)
	require.NoError(t, err)

	// Test LoadPolicy into Casbin model
	t.Run("LoadPolicy loads rules from database", func(t *testing.T) {
		// Use the model from embedded content
		m, err := model.NewModelFromString(modelContent)
		require.NoError(t, err)

		err = adapter.LoadPolicy(m)
		assert.NoError(t, err)

		// Check that policies were loaded into the model
		policies := m["p"]["p"].Policy
		assert.Len(t, policies, 2)
	})
}

func TestEntAdapterRemoveFilteredPolicy(t *testing.T) {
	// Create in-memory SQLite client for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Pre-populate database with policies
	ctx := context.Background()
	_, err := client.CasbinRule.Create().
		SetPtype("p").
		SetV0("admin").
		SetV1("data").
		SetV2("read").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.CasbinRule.Create().
		SetPtype("p").
		SetV0("admin").
		SetV1("config").
		SetV2("read").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.CasbinRule.Create().
		SetPtype("p").
		SetV0("viewer").
		SetV1("data").
		SetV2("read").
		Save(ctx)
	require.NoError(t, err)

	adapter, err := NewEntAdapter(client)
	require.NoError(t, err)

	// Remove all policies for "admin"
	t.Run("RemoveFilteredPolicy by subject", func(t *testing.T) {
		err := adapter.RemoveFilteredPolicy("p", "p", 0, "admin")
		assert.NoError(t, err)

		// Verify only viewer policy remains
		rules, err := client.CasbinRule.Query().All(ctx)
		assert.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "viewer", rules[0].V0)
	})
}

func TestGetAllPolicy(t *testing.T) {
	// Create in-memory SQLite client for testing
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// Pre-populate database
	ctx := context.Background()
	_, err := client.CasbinRule.Create().
		SetPtype("p").
		SetV0("admin").
		SetV1("*").
		SetV2("*").
		Save(ctx)
	require.NoError(t, err)

	adapter, err := NewEntAdapter(client)
	require.NoError(t, err)

	t.Run("GetAllPolicy returns all rules", func(t *testing.T) {
		policies, err := adapter.GetAllPolicy()
		assert.NoError(t, err)
		assert.Len(t, policies, 1)
		assert.Equal(t, "p", policies[0][0])     // ptype
		assert.Equal(t, "admin", policies[0][1]) // v0
	})
}

func TestParseCSVLine(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		wantRule []string
		wantOk   bool
	}{
		{
			input:    "p, admin, data, read",
			wantType: "p",
			wantRule: []string{"admin", "data", "read"},
			wantOk:   true,
		},
		{
			input:    "g, user:1, admin, project:1",
			wantType: "g",
			wantRule: []string{"user:1", "admin", "project:1"},
			wantOk:   true,
		},
		{
			input:    "g2, user:1, platform_admin",
			wantType: "g2",
			wantRule: []string{"user:1", "platform_admin"},
			wantOk:   true,
		},
		{
			input:  "invalid",
			wantOk: false,
		},
		{
			input:  "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ptype, rule, ok := parseCSVLine(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantType, ptype)
				assert.Equal(t, tt.wantRule, rule)
			}
		})
	}
}
