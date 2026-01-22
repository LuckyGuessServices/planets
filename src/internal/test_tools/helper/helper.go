package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func NewHTTPRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()

	request, errRequest := http.NewRequest(method, url, body)
	if errRequest != nil {
		t.Fatal("Failed to create a HTTP request:", errRequest)
	}

	return request
}

func RequireJSONResponse(t *testing.T, expected map[string]any, responseBody *bytes.Buffer) {
	t.Helper()

	decoder := json.NewDecoder(responseBody)
	decoder.UseNumber()

	var actual map[string]any
	if err := decoder.Decode(&actual); err != nil {
		t.Fatalf("Failed to parse response body as JSON.\nError: %v\nRaw body (within quotes): '%s'", err, responseBody.Bytes())
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("JSON mismatch (-expected +actual):\n%s", diff)
	}
}
