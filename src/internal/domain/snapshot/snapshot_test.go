package snapshot_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use the same date for all tests here. Eventually, it helps to ensure no data collision (db cleanup between tests).
const (
	commonSnapshotDateAsString = "2025-11-09"
	commonPlanetIndex          = 0
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// TestSnapshot ensures data is successfully stored to and queried from a table via ORM.
func TestQueryNew(t *testing.T) {
	// TODO Rework / Replace this test when an actual API endpoint is implemented.
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		// Imagine the date is read from API request into a handler request structure.
		requestJSON := []byte(fmt.Sprintf(`{"requested_date":"%s"}`, commonSnapshotDateAsString))
		requestData := &struct {
			RequestedDate types.Date `json:"requested_date"`
		}{}
		require.NoError(t, json.Unmarshal(requestJSON, requestData))

		// Ensure Date.MarshalJSON does it magic correctly.
		newJSON, errMarshal := json.Marshal(requestData)
		require.NoErrorf(t, errMarshal, "Failed to marshal Date '%v' back into JSON: %v", requestData, errMarshal)
		assert.JSONEq(t, string(requestJSON), string(newJSON))

		// Test Date.String() correctly formats the underlying date-time.
		requestedDate := requestData.RequestedDate
		require.Equal(t, commonSnapshotDateAsString, requestedDate.String())

		ctx := context.Background()
		snapshotTable := storage.SnapshotRepository()

		// At first, we should not find any record here (we haven't added one yet).
		snapshots, errSelect := snapshotTable.BySnapshotDate(ctx, requestedDate)
		require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", requestedDate, errSelect)
		require.Empty(t, snapshots)

		// Then imagine the data is requested successfully from an external API.
		// Fill a new row with it:
		newSnapshot := snapshotTable.NewSnapshot()
		newSnapshot.SetSnapshotDate(requestedDate)
		newSnapshot.SetPlanetIndex(commonPlanetIndex)
		newSnapshot.SetIsRetrograde(true)
		require.NoError(t, newSnapshot.Save(ctx), "Failed to insert a new snapshot.")

		// Re-request the now newly added record again and compare its contents with the original inserted structure:
		snapshots, errSelect = snapshotTable.BySnapshotDate(ctx, requestedDate)
		require.NoErrorf(t, errSelect, "Failed to select snapshots by date '%s': %v", requestedDate, errSelect)
		require.Len(t, snapshots, 1)
		assert.NotSame(t, newSnapshot, snapshots[0])
		assert.Equal(t, newSnapshot, snapshots[0])
	})
}

// TestUpdateOld asserts:
// 1. Scanning a null field.
// 2. Updating a nullable field with an actual value.
// 3. TxDB test: no extra records exist after the previous test.
func TestUpdateOld(t *testing.T) {
	// TODO Rework / Replace this test when an actual API endpoint is implemented.
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		requestedTimeObj, errTimeParse := time.Parse(time.DateOnly, commonSnapshotDateAsString)
		require.NoError(t, errTimeParse)
		requestedDate := types.NewDate(requestedTimeObj)

		ctx := context.Background()
		snapshotTable := storage.SnapshotRepository()

		// This should fail if there is a record from the previous test.
		newSnapshot := snapshotTable.NewSnapshot()
		newSnapshot.SetSnapshotDate(requestedDate)
		newSnapshot.SetPlanetIndex(commonPlanetIndex)
		require.NoError(t, newSnapshot.Save(ctx), "Failed to insert a new snapshot.")

		firstCollection, errFirstSelect := snapshotTable.BySnapshotDate(ctx, requestedDate)
		require.NoErrorf(
			t,
			errFirstSelect,
			"(1) Failed to select snapshots by date '%s': %v",
			commonSnapshotDateAsString,
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
		secondCollection, errSecondSelect := snapshotTable.BySnapshotDate(ctx, requestedDate)
		require.NoErrorf(
			t,
			errSecondSelect,
			"(2) Failed to select snapshots by date '%s': %v", commonSnapshotDateAsString,
			errSecondSelect,
		)
		require.Len(t, secondCollection, 1)
		assert.NotSame(t, firstSnapshot, secondCollection[0])
		assert.Equal(t, firstSnapshot.ID(), secondCollection[0].ID())
		assert.Equal(t, firstSnapshot.IsRetrogradeRaw(), secondCollection[0].IsRetrogradeRaw())
	})
}
