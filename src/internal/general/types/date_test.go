package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests converting time into "date only" variable (omitting hours, minutes, seconds, nanoseconds).
//
// See types.NewDate
func TestDateCreation(t *testing.T) {
	expectedTime := time.Date(2026, time.February, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		inputTime time.Time
	}{
		{name: "midnight", inputTime: time.Date(2026, time.February, 3, 0, 0, 0, 0, time.UTC)},
		{name: "various", inputTime: time.Date(2026, time.February, 3, 17, 23, 12, 586, time.UTC)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				actualDate := types.NewDate(test.inputTime)
				assert.Equal(t, expectedTime, actualDate.Time)
			})
		})
	}
}

// Tests parsing date from a string.
//
// See types.NewDateParsed, types.NewDate
func TestDateCreationFromString(t *testing.T) {
	expectedTime := time.Date(2026, time.February, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		dateString    string
		hasParseError bool
	}{
		{name: "RFC3339", dateString: "2026-02-03T17:25:00Z", hasParseError: false},
		{name: "date-only", dateString: "2026-02-03", hasParseError: false},
		{name: "Little-Endian", dateString: "03/02/2026", hasParseError: true}, // Ambiguous format.
		{name: "Middle-Endian", dateString: "02/03/2026", hasParseError: true}, // Ambiguous format.
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				parsedDate, errTimeParse := types.NewDateParsed(test.dateString)
				if test.hasParseError {
					assert.Error(t, errTimeParse)
				} else {
					require.NoError(t, errTimeParse)
					assert.Equal(t, expectedTime, parsedDate.Time)
				}
			})
		})
	}
}

// Tests fmt.Stringer implementation for Date.
//
// See Date.String()
func TestConvertingDateToString(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		assert.Equal(t, "2000-01-01", types.NewDate(time.Now().UTC()).String())
	})
}

// Tests types.Date JSON processing.
//
// See Date.MarshalJSON, Date.UnmarshalJSON
func TestMarshallingDate(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		requestJSON := []byte(`{"requested_date":"2026-02-21"}`)
		requestData := &struct {
			RequestedDate types.Date `json:"requested_date"`
		}{}
		require.NoError(t, json.Unmarshal(requestJSON, requestData))

		expectedDate := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedDate, requestData.RequestedDate.Time)

		newJSON, errMarshal := json.Marshal(requestData)
		require.NoErrorf(t, errMarshal, "Failed to marshal Date '%v' back into JSON: %v", requestData, errMarshal)
		assert.JSONEq(t, string(requestJSON), string(newJSON))
	})
}
