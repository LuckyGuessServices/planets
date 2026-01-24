package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/api/router"
	"github.com/google/go-cmp/cmp"
)

// QueryApi makes a request to the service's API and returns the recorded response.
//
// apiURLSuffixPath is a request path going after [router.PrefixGeneral].
func QueryApi(t *testing.T, method, apiURLSuffixPath string) *httptest.ResponseRecorder {
	t.Helper()

	request, errRequest := http.NewRequest(method, router.PrefixGeneral+apiURLSuffixPath, nil)
	if errRequest != nil {
		t.Fatal("Failed to create a HTTP request:", errRequest)
	}

	response := httptest.NewRecorder()
	router.Create().ServeHTTP(response, request)

	return response
}

// RequireJSONResponse prepares the expected map (make a copy if you intend to use the original expected map later),
// then compares it with JSON response body decoded into another map.
func RequireJSONResponse(t *testing.T, mutableExpected map[string]any, responseBody *bytes.Buffer) {
	t.Helper()

	// Prepare the map of expected values to respect JSON numbers unmarshalling politic.
	for key, storedValue := range mutableExpected {
		switch value := storedValue.(type) {
		case int:
			mutableExpected[key] = float64(value)
		case int8:
			mutableExpected[key] = float64(value)
		case int16:
			mutableExpected[key] = float64(value)
		case int32:
			mutableExpected[key] = float64(value)
		case int64:
			mutableExpected[key] = strconv.FormatInt(value, 10)
		case uint64:
			mutableExpected[key] = strconv.FormatUint(value, 10)
		default:
			// No conversion needed.
		}
	}

	responseBodyBytes := responseBody.Bytes()
	var actual map[string]any
	if err := json.Unmarshal(responseBodyBytes, &actual); err != nil {
		t.Fatalf(
			"Failed to parse response body as JSON.\nError: %v\nRaw body (within quotes): '%s'",
			err,
			responseBodyBytes,
		)
	}

	if diff := cmp.Diff(mutableExpected, actual); diff != "" {
		t.Errorf("JSON mismatch (-expected +actual):\n%s", diff)
	}
}
