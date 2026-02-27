package endpoints_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/api"
	"github.com/LuckyGuessServices/planets/internal/api/endpoints"
	"github.com/LuckyGuessServices/planets/internal/domain/planets"
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot"
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/integration/mercury"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/helper"
	"github.com/LuckyGuessServices/planets/internal/test_tools/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commonRequestedDateString string = "2025-11-10"
	apiPathOnDate             string = "/on-date"
)

var commonRequestedDate = helper.NewDateParsedOrPanic(commonRequestedDateString)

// Tests a successful request, when the necessary data is already present in a local storage.
//
// See endpoints.OnDate
func TestOnDateSuccess(t *testing.T) {
	tests := []struct {
		name      string
		dateValue string
	}{
		{name: "no-quotes", dateValue: commonRequestedDateString},
		{name: "with-quotes", dateValue: `"` + commonRequestedDateString + `"`},
		{name: "no-dashes", dateValue: `"` + strings.ReplaceAll(commonRequestedDateString, "-", "") + `"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				// Place necessary data in a local storage:
				ctxInsert := context.Background()
				newSnapshot := storage.SnapshotRepository().NewSnapshot()
				newSnapshot.SetSnapshotDate(commonRequestedDate)
				newSnapshot.SetPlanetIndex(planets.Mercury().Index())
				newSnapshot.SetIsRetrograde(true)
				require.NoErrorf(t, newSnapshot.Save(ctxInsert), "Failed to insert a new snapshot.")

				// Then make an API request and assert the response.

				requestParameters := &url.Values{}
				requestParameters.Set("date", test.dateValue)
				response := helper.QueryApiGet(t, apiPathOnDate, requestParameters).Result()
				defer func() { _ = response.Body.Close() }()

				assert.Equal(t, http.StatusOK, response.StatusCode)

				expectedBody := map[string]any{
					"snapshots": []map[string]any{
						{
							"planet_index":  planets.Mercury().Index(),
							"is_retrograde": true,
						},
					},
				}
				helper.RequireJSONResponse(t, expectedBody, helper.ExtractAndCloseResponseBody(t, response))
			})
		})
	}
}

// Tests cases, when the necessary data is not found in local storage and requested from an external source.
// Firstly, local API will advise to re-request the same query later.
// With the next request local API will return data (if an external request was successful) or an error.
//
// See endpoints.OnDate
func TestOnDateRequestExternalWhenMissingLocal(t *testing.T) {
	transport := mock.NewTransport()

	tests := []struct {
		name                    string
		isExpectedDataValid     bool
		externalResponse        *http.Response
		expectedLocalStatusCode int
		expectedLocalBody       map[string]any
	}{
		{
			name:                "success",
			isExpectedDataValid: true,
			externalResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"is_retrograde":true}`)),
				Header:     make(http.Header),
			},
			expectedLocalStatusCode: http.StatusOK,
			expectedLocalBody: map[string]any{
				"snapshots": []map[string]any{
					{
						"planet_index":  planets.Mercury().Index(),
						"is_retrograde": true,
					},
				},
			},
		},
		{
			name:                "failure",
			isExpectedDataValid: false,
			externalResponse: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`Internal Server Error`)),
				Header:     make(http.Header),
			},
			expectedLocalStatusCode: http.StatusNotFound,
			expectedLocalBody: map[string]any{
				"error": api.ResponseErrorDataUnavailable,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				mercury.SetMockClientForTests("", transport)

				externalRequestCount := 0
				transport.RoundTripFunction = func(request *http.Request) (*http.Response, error) {
					externalRequestCount++

					return test.externalResponse, nil
				}

				// Ensure that initially the necessary data is missing in a local storage.
				ctxLocalStorageQuery := context.Background()
				localData, errLocalData := storage.SnapshotRepository().BySnapshotDate(
					ctxLocalStorageQuery,
					commonRequestedDate,
				)
				require.NoError(t, errLocalData)
				require.Empty(t, localData)

				// Make an API request:
				requestParameters := &url.Values{}
				requestParameters.Set("date", commonRequestedDateString)
				response := helper.QueryApiGet(t, apiPathOnDate, requestParameters).Result()
				defer func() { _ = response.Body.Close() }()

				assert.Equal(t, http.StatusAccepted, response.StatusCode)

				// For now, there is no data, and an external request is scheduled.
				// Assert particular headers:
				responseHeaders := response.Header
				retryAfter, isSetRetryAfter := responseHeaders["Retry-After"]
				require.Truef(t, isSetRetryAfter, `"Retry-After" header is missing.`)
				assert.Lenf(t, retryAfter, 1, `"Retry-After" header must be set once.`)
				assert.Equal(t, strconv.Itoa(endpoints.OnDateRetryAfterSeconds), retryAfter[0])

				// The response should reflect the key header:
				helper.RequireJSONResponse(
					t,
					map[string]any{
						"message": "Request accepted, no data available yet. Retry after %d seconds." +
							strconv.Itoa(endpoints.OnDateRetryAfterSeconds) + " seconds.",
					},
					helper.ExtractAndCloseResponseBody(t, response),
				)

				// Then a request to an external source is made in background.
				assert.Equalf(t, 1, externalRequestCount, "The external API should be requested exactly once.")

				// After that, a new record must always be added to a local storage:
				localData, errLocalData = storage.SnapshotRepository().BySnapshotDate(
					ctxLocalStorageQuery,
					commonRequestedDate,
				)
				require.NoError(t, errLocalData)
				require.Len(t, localData, 1)
				addedSnapshot := localData[0]
				assert.Equal(t, planets.Mercury().Index(), addedSnapshot.PlanetIndex())
				// ... But its validity depends on external request success:
				if test.isExpectedDataValid {
					assert.Truef(t, addedSnapshot.IsRetrogradeRaw().Valid, `"is_retrograde" field must not be NULL`)
					assert.True(t, addedSnapshot.IsRetrogradeRaw().V)
				} else {
					assert.Falsef(t, addedSnapshot.IsRetrogradeRaw().Valid, `"is_retrograde" field must be NULL`)
				}

				// Next, we make the same request to our API again:
				response = helper.QueryApiGet(t, apiPathOnDate, requestParameters).Result()
				defer func() { _ = response.Body.Close() }()

				// From now on, the result depends on external request success.
				assert.Equal(t, test.expectedLocalStatusCode, response.StatusCode)
				helper.RequireJSONResponse(t, test.expectedLocalBody, helper.ExtractAndCloseResponseBody(t, response))
			})
		})
	}
}

// Tests cases, when local data exists, but a part of it is corrupted or missing.
// Ensures there will be no external requests. Asserts different responses.
//
// See endpoints.OnDate
func TestOnDateLocalDataDifferentStates(t *testing.T) {
	transport := mock.NewTransport()

	type snapshotData struct {
		PlanetIndex  int16
		IsRetrograde sql.Null[bool]
	}

	const someOtherPlanetIndex int16 = 0

	tests := []struct {
		name               string
		initialSnapshots   []snapshotData
		expectedStatusCode int
		expectedBody       map[string]any
	}{
		{
			name: "all-defined",
			initialSnapshots: []snapshotData{
				{PlanetIndex: planets.Mercury().Index(), IsRetrograde: sql.Null[bool]{V: true, Valid: true}},
				{PlanetIndex: someOtherPlanetIndex, IsRetrograde: sql.Null[bool]{V: true, Valid: true}},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: map[string]any{
				"snapshots": []map[string]any{
					{
						"planet_index":  someOtherPlanetIndex,
						"is_retrograde": true,
					},
					{
						"planet_index":  planets.Mercury().Index(),
						"is_retrograde": true,
					},
				},
			},
		},
		{
			name: "mercury-null",
			initialSnapshots: []snapshotData{
				{PlanetIndex: planets.Mercury().Index(), IsRetrograde: sql.Null[bool]{V: false, Valid: false}},
				{PlanetIndex: someOtherPlanetIndex, IsRetrograde: sql.Null[bool]{V: true, Valid: true}},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: map[string]any{
				"snapshots": []map[string]any{
					{
						"planet_index":  someOtherPlanetIndex,
						"is_retrograde": true,
					},
				},
			},
		},
		{
			name: "mercury-missing",
			initialSnapshots: []snapshotData{
				// No record about Mercury.
				{PlanetIndex: someOtherPlanetIndex, IsRetrograde: sql.Null[bool]{V: true, Valid: true}},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: map[string]any{
				"snapshots": []map[string]any{
					{
						"planet_index":  someOtherPlanetIndex,
						"is_retrograde": true,
					},
				},
			},
		},
		{
			name: "all-nulls",
			initialSnapshots: []snapshotData{
				{PlanetIndex: planets.Mercury().Index(), IsRetrograde: sql.Null[bool]{V: false, Valid: false}},
				{PlanetIndex: someOtherPlanetIndex, IsRetrograde: sql.Null[bool]{V: false, Valid: false}},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody: map[string]any{
				"error": api.ResponseErrorDataUnavailable,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				mercury.SetMockClientForTests("", transport)

				externalRequestCount := 0
				transport.RoundTripFunction = func(request *http.Request) (*http.Response, error) {
					externalRequestCount++

					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader(`Internal Server Error`)),
						Header:     make(http.Header),
					}, nil
				}

				// Add initial records:
				snapshotsRepo := storage.SnapshotRepository()
				var record *snapshot.Snapshot
				for _, data := range test.initialSnapshots {
					record = snapshotsRepo.NewSnapshot()
					record.SetSnapshotDate(commonRequestedDate)
					record.SetPlanetIndex(data.PlanetIndex)
					if data.IsRetrograde.Valid {
						record.SetIsRetrograde(data.IsRetrograde.V)
					}
					require.NoError(t, record.Save(context.Background()))
				}

				// Make an API request and assert the response.

				requestParameters := &url.Values{}
				requestParameters.Set("date", commonRequestedDateString)
				response := helper.QueryApiGet(t, apiPathOnDate, requestParameters).Result()
				defer func() { _ = response.Body.Close() }()

				assert.Equal(t, test.expectedStatusCode, response.StatusCode)

				helper.RequireJSONResponse(t, test.expectedBody, helper.ExtractAndCloseResponseBody(t, response))

				assert.Equalf(t, 0, externalRequestCount, "There should be no external API requests.")
			})
		})
	}
}

// Tests responses for various parameter invalid values.
//
// See endpoints.OnDate
func TestOnDateParameterValidation(t *testing.T) {
	tests := []struct {
		name                       string
		isDatePresent              bool
		dateValue                  string
		expectedParameterErrorCode string
	}{
		{
			name:                       "missing",
			isDatePresent:              false,
			dateValue:                  "",
			expectedParameterErrorCode: api.ResponseParameterErrorRequired,
		},
		{
			name:                       "empty-string",
			isDatePresent:              true,
			dateValue:                  "",
			expectedParameterErrorCode: api.ResponseParameterErrorInvalid,
		},
		{
			name:                       "parts-positions",
			isDatePresent:              true,
			dateValue:                  `10-11-2025`,
			expectedParameterErrorCode: api.ResponseParameterErrorInvalid,
		},
		{
			name:                       "has-time",
			isDatePresent:              true,
			dateValue:                  `2025-10-11T12:23:46+06:00`,
			expectedParameterErrorCode: api.ResponseParameterErrorInvalid,
		},
		{
			name:                       "not-date",
			isDatePresent:              true,
			dateValue:                  `z`,
			expectedParameterErrorCode: api.ResponseParameterErrorInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				requestParameters := &url.Values{}
				if test.isDatePresent {
					requestParameters.Set("date", test.dateValue)
				}

				response := helper.QueryApiGet(t, apiPathOnDate, requestParameters).Result()
				defer func() { _ = response.Body.Close() }()

				assert.Equal(t, http.StatusBadRequest, response.StatusCode)

				expectedBody := map[string]any{
					"error": api.ResponseErrorInvalidParameters,
					"parameter_errors": map[string]any{
						"date": test.expectedParameterErrorCode,
					},
				}
				helper.RequireJSONResponse(t, expectedBody, helper.ExtractAndCloseResponseBody(t, response))
			})
		})
	}
}
