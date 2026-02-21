package snapshot_test

import (
	"context"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const commonPlanetIndex int16 = 0

// Use the same date for all tests here. Eventually, it helps to ensure no data collision (db cleanup between tests).
var snapshotDate types.Date

func init() {
	var errParse error
	snapshotDate, errParse = types.NewDateParsed("2025-11-10")
	if errParse != nil {
		luglog.Panic("Failed to init 'snapshotDate': ", errParse)
	}
}

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// Ensures data is successfully stored to and queried from a table via ORM.
//
// See types.Date, storage.SnapshotRepository, RepositoryInterface.BySnapshotDate, RepositoryInterface.NewSnapshot
func TestQueryNew(t *testing.T) {
	// TODO Rework / Replace this test when an actual API endpoint is implemented.
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		ctx := context.Background()
		snapshotTable := storage.SnapshotRepository()

		// At first, we should not find any record here (we haven't added one yet).
		snapshots, errSelect := snapshotTable.BySnapshotDate(ctx, snapshotDate)
		require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", snapshotDate, errSelect)
		require.Empty(t, snapshots)

		// Then imagine the data is requested successfully from an external API.
		// Fill a new row with it:
		newSnapshot := snapshotTable.NewSnapshot()
		newSnapshot.SetSnapshotDate(snapshotDate)
		newSnapshot.SetPlanetIndex(commonPlanetIndex)
		newSnapshot.SetIsRetrograde(true)
		require.NoErrorf(t, newSnapshot.Save(ctx), "Failed to insert a new snapshot.")

		// Re-request the now newly added record again and compare its contents with the original inserted structure:
		snapshots, errSelect = snapshotTable.BySnapshotDate(ctx, snapshotDate)
		require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", snapshotDate, errSelect)
		require.Len(t, snapshots, 1)
		assert.NotSame(t, newSnapshot, snapshots[0])
		assert.Equal(t, newSnapshot, snapshots[0])
	})
}

// Asserts:
// 1. Scanning a null field.
// 2. Updating a nullable field with an actual value.
// 3. TxDB test: no extra records exist after the previous test.
//
// See types.Date, storage.SnapshotRepository, RepositoryInterface.BySnapshotDate, RepositoryInterface.NewSnapshot
func TestUpdateOld(t *testing.T) {
	// TODO Rework / Replace this test when an actual API endpoint is implemented.
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		ctx := context.Background()
		snapshotTable := storage.SnapshotRepository()

		// This should fail if there is a record from the previous test.
		newSnapshot := snapshotTable.NewSnapshot()
		newSnapshot.SetSnapshotDate(snapshotDate)
		newSnapshot.SetPlanetIndex(commonPlanetIndex)
		require.NoErrorf(t, newSnapshot.Save(ctx), "Failed to insert a new snapshot.")

		firstCollection, errFirstSelect := snapshotTable.BySnapshotDate(ctx, snapshotDate)
		require.NoErrorf(
			t,
			errFirstSelect,
			"(1) Failed to select snapshots by date '%s': %v",
			snapshotDate.String(),
			errFirstSelect,
		)
		require.Len(t, firstCollection, 1)

		firstSnapshot := firstCollection[0]
		// At first, the record's flag is NULL:
		assert.False(t, firstSnapshot.IsRetrogradeRaw().Valid)

		// Let's fix it and set an explicit value:
		firstSnapshot.SetIsRetrograde(true)
		// Let's ensure the flag is set to `true` now:
		assert.True(t, firstSnapshot.IsRetrogradeRaw().Valid)
		assert.True(t, firstSnapshot.IsRetrogradeRaw().V)
		// Update (not insert) the record:
		require.NoError(t, firstSnapshot.Save(ctx))

		// We request the same snapshots collection and should expect the same data set.
		secondCollection, errSecondSelect := snapshotTable.BySnapshotDate(ctx, snapshotDate)
		require.NoErrorf(
			t,
			errSecondSelect,
			"(2) Failed to select snapshots by date '%s': %v", snapshotDate.String(),
			errSecondSelect,
		)
		require.Len(t, secondCollection, 1)
		assert.NotSame(t, firstSnapshot, secondCollection[0])
		assert.Equal(t, firstSnapshot.ID(), secondCollection[0].ID())
		assert.Equal(t, firstSnapshot.IsRetrogradeRaw(), secondCollection[0].IsRetrogradeRaw())
	})
}
