package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// TestSnapshotBunInsertHook ensures the ORM entity hooked fields are set properly when a record is inserted.
func TestSnapshotBunInsertHook(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		row := storage.SnapshotBun{}
		assert.Zero(t, row.ColumnID)
		assert.Zero(t, row.ColumnCreatedAt)

		require.NoError(t, row.Save(context.Background()), "Failed to insert a snapshot.")
		assert.NotZero(t, row.ColumnID)
		assert.Equal(t, "2000-01-01T00:00:00Z", row.ColumnCreatedAt.Format(time.RFC3339))
	})
}

// TestSnapshotBunUpdateHook ensures the ORM entity certain hooked fields are not changed
// after updating in a database.
func TestSnapshotBunUpdateHook(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		ctx := context.Background()

		row := storage.SnapshotBun{}
		require.NoError(t, row.Save(ctx), "Failed to insert a snapshot.")

		idBeforeUpdate := row.ColumnID
		createdAtBeforeUpdate := row.ColumnCreatedAt

		// Now let's update the record and ensure no certain fields are updated automatically.
		row.SetIsRetrograde(true)
		require.NoError(t, row.Save(ctx), "Failed to update a snapshot.")
		assert.Equal(t, idBeforeUpdate, row.ColumnID)
		assert.Equal(t, createdAtBeforeUpdate, row.ColumnCreatedAt)
	})
}
