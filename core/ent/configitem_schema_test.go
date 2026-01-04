package ent_test

import (
	"context"
	"testing"
	"time"

	"apprun/ent/configitem"
	"apprun/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigitem_UniqueKeyConstraint tests key unique constraint
func TestConfigitem_UniqueKeyConstraint(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Create first config item
	_, err := client.Configitem.Create().
		SetKey("test.key").
		SetValue("value1").
		Save(ctx)
	require.NoError(t, err)

	// Try to create duplicate key (should fail)
	_, err = client.Configitem.Create().
		SetKey("test.key").
		SetValue("value2").
		Save(ctx)
	assert.Error(t, err, "duplicate key should fail")
}

// TestConfigitem_StatusEnum tests status enum validation
func TestConfigitem_StatusEnum(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Test active status
	item1, err := client.Configitem.Create().
		SetKey("test.active").
		SetValue("value").
		SetStatus(configitem.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, configitem.StatusActive, item1.Status)

	// Test inactive status
	item2, err := client.Configitem.Create().
		SetKey("test.inactive").
		SetValue("value").
		SetStatus(configitem.StatusInactive).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, configitem.StatusInactive, item2.Status)

	// Test default status (should be active)
	item3, err := client.Configitem.Create().
		SetKey("test.default").
		SetValue("value").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, configitem.StatusActive, item3.Status)
}

// TestConfigitem_AutoTimestamps tests created_at and updated_at
func TestConfigitem_AutoTimestamps(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()
	beforeCreate := time.Now()

	// Create item
	item, err := client.Configitem.Create().
		SetKey("test.timestamps").
		SetValue("value1").
		Save(ctx)
	require.NoError(t, err)

	afterCreate := time.Now()

	// Verify created_at is set
	assert.True(t, item.CreatedAt.After(beforeCreate) || item.CreatedAt.Equal(beforeCreate))
	assert.True(t, item.CreatedAt.Before(afterCreate) || item.CreatedAt.Equal(afterCreate))

	// Verify updated_at is set
	assert.True(t, item.UpdatedAt.After(beforeCreate) || item.UpdatedAt.Equal(beforeCreate))
	assert.True(t, item.UpdatedAt.Before(afterCreate) || item.UpdatedAt.Equal(afterCreate))

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()

	updated, err := client.Configitem.UpdateOne(item).
		SetValue("value2").
		Save(ctx)
	require.NoError(t, err)

	// Verify created_at unchanged
	assert.WithinDuration(t, item.CreatedAt, updated.CreatedAt, time.Millisecond)

	// Verify updated_at changed
	assert.True(t, updated.UpdatedAt.After(item.UpdatedAt) || updated.UpdatedAt.Equal(beforeUpdate))
} // TestConfigitem_StatusIndex tests status index for filtering
func TestConfigitem_StatusIndex(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Create test data
	_, err := client.Configitem.Create().
		SetKey("active1").
		SetValue("v1").
		SetStatus(configitem.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Configitem.Create().
		SetKey("active2").
		SetValue("v2").
		SetStatus(configitem.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Configitem.Create().
		SetKey("inactive1").
		SetValue("v3").
		SetStatus(configitem.StatusInactive).
		Save(ctx)
	require.NoError(t, err)

	// Query active items
	activeItems, err := client.Configitem.Query().
		Where(configitem.StatusEQ(configitem.StatusActive)).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, activeItems, 2)

	// Query inactive items
	inactiveItems, err := client.Configitem.Query().
		Where(configitem.StatusEQ(configitem.StatusInactive)).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, inactiveItems, 1)
}

// TestConfigitem_SoftDelete tests soft delete behavior
func TestConfigitem_SoftDelete(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx := context.Background()

	// Create item
	item, err := client.Configitem.Create().
		SetKey("test.soft.delete").
		SetValue("value").
		Save(ctx)
	require.NoError(t, err)

	// Soft delete (set to inactive)
	err = client.Configitem.UpdateOne(item).
		SetStatus(configitem.StatusInactive).
		Exec(ctx)
	require.NoError(t, err)

	// Item still exists in database
	count, err := client.Configitem.Query().
		Where(configitem.KeyEQ("test.soft.delete")).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// But not in active items
	activeCount, err := client.Configitem.Query().
		Where(
			configitem.KeyEQ("test.soft.delete"),
			configitem.StatusEQ(configitem.StatusActive),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, activeCount)
}
