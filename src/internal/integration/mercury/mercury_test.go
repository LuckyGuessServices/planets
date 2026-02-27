package mercury_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/integration"
	"github.com/LuckyGuessServices/planets/internal/integration/mercury"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/helper"
	"github.com/LuckyGuessServices/planets/internal/test_tools/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var requestedDate = helper.NewDateParsedOrPanic("2025-11-10")

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// Tests the default API client transport in case a transport mock is not set.
//
// See mercury.Client, mercury.SetMockClientForTests
func TestClientMock(t *testing.T) {
	tests := []struct {
		name           string
		isClientMocked bool
	}{
		{name: "mocked", isClientMocked: true},
		{name: "default", isClientMocked: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				if test.isClientMocked {
					transport := mock.NewTransport()
					transport.RoundTripFunction = func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{"is_retrograde":true}`)),
							Header:     make(http.Header),
						}, nil
					}
					mercury.SetMockClientForTests("", transport)
				}

				// Here any method might be used. The point is to distinguish a mocked answer from a not mocked one.
				returnedValue, requestError := mercury.Client().ByDate(context.Background(), requestedDate)
				if test.isClientMocked {
					require.NoError(t, requestError)
					assert.True(t, returnedValue)
				} else {
					require.Error(t, requestError)
					assert.Equal(
						t,
						`unexpected response status code: 500; raw response body (within quotes): `+
							`'An API client must be mocked in "app_test" environment. `+
							`Init a mock client instance with "SetMockClientForTests".'`,
						requestError.Error(),
					)
				}
			})
		})
	}
}

// Asserts successful responses.
//
// See APIClient.ByDate
func TestByDateSuccess(t *testing.T) {
	transport := mock.NewTransport()

	tests := []struct {
		name          string
		mockValue     string
		expectedValue bool
	}{
		{name: "value-false", mockValue: "false", expectedValue: false},
		{name: "value-true", mockValue: "true", expectedValue: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				mercury.SetMockClientForTests("", transport)

				mockResponseBody := fmt.Sprintf(`{"junk_data":"junk","is_retrograde":%s}`, test.mockValue)
				transport.RoundTripFunction = func(_ *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(mockResponseBody)),
						Header:     make(http.Header),
					}, nil
				}

				actualValue, requestError := mercury.Client().ByDate(context.Background(), requestedDate)
				require.NoError(t, requestError)
				assert.Equal(t, test.expectedValue, actualValue)
			})
		})
	}
}

// Tests failure and an error message to find necessary data in JSON response.
//
// See APIClient.ByDate
func TestByDateDataMissing(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		transport := mock.NewTransport()
		mercury.SetMockClientForTests("", transport)

		transport.RoundTripFunction = func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"junk_data":"junk"}`)),
				Header:     make(http.Header),
			}, nil
		}

		_, requestError := mercury.Client().ByDate(context.Background(), requestedDate)
		require.Error(t, requestError)
		assert.ErrorIs(t, requestError, integration.ErrDataMissing)
		assert.Equal(
			t,
			`response body lacks required data; raw contents (within quotes): '{"junk_data":"junk"}'`,
			requestError.Error(),
		)
	})
}

// Ensures the correct error message (including a response body), when a response contains invalid status code.
//
// See APIClient.ByDate
func TestByDateInvalidStatusCode(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		transport := mock.NewTransport()
		mercury.SetMockClientForTests("", transport)

		transport.RoundTripFunction = func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`VERY IMPORTANT MESSAGE`)),
				Header:     make(http.Header),
			}, nil
		}

		_, requestError := mercury.Client().ByDate(context.Background(), requestedDate)
		require.Error(t, requestError)
		assert.ErrorIs(t, requestError, integration.ErrUnexpectedResponseStatusCode)
		assert.Equal(
			t,
			`unexpected response status code: 400; raw response body (within quotes): 'VERY IMPORTANT MESSAGE'`,
			requestError.Error(),
		)
	})
}

// Tests cases and asserts expected error strings when failing to Unmarshal a JSON response body.
//
// See APIClient.ByDate
func TestByDateInvalidResponseBody(t *testing.T) {
	transport := mock.NewTransport()
	mercury.SetMockClientForTests("", transport)

	tests := []struct {
		name                 string
		responseBodyContents string
		expectedErrorType    string
		expectedErrorString  string
	}{
		{
			name:                 "valid-json",
			responseBodyContents: `{"is_retrograde":true}`,
			expectedErrorType:    "json.SyntaxError",
			expectedErrorString:  ``,
		},
		{
			name:                 "invalid-value-type",
			responseBodyContents: `{"is_retrograde":"true"}`, // Value is string instead of bool.
			expectedErrorType:    "json.UnmarshalTypeError",
			expectedErrorString: `json: cannot unmarshal string into Go struct field .is_retrograde of type bool; ` +
				`raw contents (within quotes): '{"is_retrograde":"true"}'`,
		},
		{
			name:                 "invalid-json",
			responseBodyContents: `"is_retrograde":true`, // No surrounding curly brackets.
			expectedErrorType:    "json.SyntaxError",
			expectedErrorString: `invalid character ':' after top-level value; ` +
				`raw contents (within quotes): '"is_retrograde":true'`,
		},
		{
			name:                 "not-json",
			responseBodyContents: `Is retrograde: true`,
			expectedErrorType:    "json.SyntaxError",
			expectedErrorString: `invalid character 'I' looking for beginning of value; ` +
				`raw contents (within quotes): 'Is retrograde: true'`,
		},
		{
			name:                 "no-body",
			responseBodyContents: ``,
			expectedErrorType:    "json.SyntaxError",
			expectedErrorString:  `unexpected end of JSON input; raw contents (within quotes): ''`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				mercury.SetMockClientForTests("", transport)

				transport.RoundTripFunction = func(_ *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(test.responseBodyContents)),
						Header:     make(http.Header),
					}, nil
				}

				_, requestError := mercury.Client().ByDate(context.Background(), requestedDate)
				if test.expectedErrorString != "" {
					require.Error(t, requestError)

					var isJSONSyntaxError bool
					switch test.expectedErrorType {
					case "json.SyntaxError":
						_, isJSONSyntaxError = errors.AsType[*json.SyntaxError](requestError)
					case "json.UnmarshalTypeError":
						_, isJSONSyntaxError = errors.AsType[*json.UnmarshalTypeError](requestError)
					default:
						t.Fatalf("Unknown expected error type '%s'", test.expectedErrorType)
					}
					assert.Truef(t, isJSONSyntaxError, "%#v", requestError)

					assert.Equal(t, test.expectedErrorString, requestError.Error())
				} else {
					require.NoError(t, requestError)
				}
			})
		})
	}
}
