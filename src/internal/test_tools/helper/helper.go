package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/api/router"
	"github.com/google/go-cmp/cmp"
)

// QueryApiGet makes a GET request to the service's API and returns the recorded response.
//
// endpointPath is a request path going after [router.PrefixGeneral] (e.g., "/something" instead of "/api/something").
func QueryApiGet(t *testing.T, endpointPath string, queryParameters *url.Values) *httptest.ResponseRecorder {
	t.Helper()

	requestURL, errURLParse := url.Parse(router.PrefixGeneral)
	if errURLParse != nil {
		t.Fatalf("Failed to parse API base URL '%s': %v", router.PrefixGeneral, errURLParse)
	}
	requestURL = requestURL.JoinPath(endpointPath)
	if queryParameters != nil {
		requestURL.RawQuery = queryParameters.Encode()
	}

	request, errRequest := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	if errRequest != nil {
		t.Fatal("Failed to create a HTTP request:", errRequest)
	}

	response := httptest.NewRecorder()
	router.Create().ServeHTTP(response, request)

	return response
}

// ExtractAndCloseResponseBody extracts response body as a slice of bytes and closes http.Response.Body stream.
func ExtractAndCloseResponseBody(t *testing.T, response *http.Response) []byte {
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Fatal("Failed to close response body: ", err)
		}
	}()

	readBody, errRead := io.ReadAll(response.Body)
	if errRead != nil {
		t.Fatal("Failed to read response body: ", errRead)
	}

	return readBody
}

// RequireJSONResponse compares the expected map with JSON response body normalized to another map.
func RequireJSONResponse(t *testing.T, expectedData map[string]any, responseBody []byte) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	decoder.UseNumber()
	var actualData map[string]any
	if err := decoder.Decode(&actualData); err != nil {
		t.Fatalf(
			"Failed to parse response body as JSON.\nError: %v\nRaw body (within quotes): '%s'",
			err,
			responseBody,
		)
	}

	if diff := cmp.Diff(expectedData, actualData, cmpJSONNormalizer); diff != "" {
		t.Errorf("JSON mismatch (-expected +actual):\n%s", diff)
	}
}

func shouldNormalizeJSONValue(value any) bool {
	if value == nil {
		return false
	}
	if _, isJSONNumber := value.(json.Number); isJSONNumber {
		return false
	}

	reflectionType := reflect.TypeOf(value)
	switch reflectionType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Slice:
		// If it's a slice, but not of "[]any" kind.
		return reflectionType.Elem().Kind() != reflect.Interface
	default:
		return false
	}
}

var cmpJSONNormalizer = cmp.FilterValues(
	// Normalize only particular values.
	// Prevent normalizing the same values twice.
	func(expectedValue, actualValue any) bool {
		return shouldNormalizeJSONValue(expectedValue) || shouldNormalizeJSONValue(actualValue)
	},
	cmp.Transformer("NormalizeForJSON", func(inputValue any) any {
		valueReflection := reflect.ValueOf(inputValue)
		if !valueReflection.IsValid() {
			return nil
		}

		switch valueReflection.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return json.Number(strconv.FormatInt(valueReflection.Int(), 10))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return json.Number(strconv.FormatUint(valueReflection.Uint(), 10))

		case reflect.Slice:
			// Handle the slice type mismatch ( []map -> []any ).
			genericSlice := make([]any, valueReflection.Len())
			for index := 0; index < valueReflection.Len(); index++ {
				genericSlice[index] = valueReflection.Index(index).Interface()
			}

			return genericSlice

		default:
			return inputValue
		}
	}),
)
