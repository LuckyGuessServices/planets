package endpoints_test

import (
	"LuG/planets/api"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"

	"github.com/google/go-cmp/cmp"
)

// TestIndex Tests a typical request and response.
func TestIndex(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		request, errRequest := http.NewRequest("GET", "/api/", nil)
		if errRequest != nil {
			t.Fatal(errRequest)
		}
		response := httptest.NewRecorder()
		api.Router().ServeHTTP(response, request)

		if status := response.Code; status != http.StatusOK {
			t.Errorf("Expected HTTP status %v, got %v", http.StatusOK, status)
		}

		synctest.Wait()

		bodyActual := response.Body.Bytes()
		var bodyMapActual map[string]any
		if errUnmarshal := json.Unmarshal(bodyActual, &bodyMapActual); errUnmarshal != nil {
			t.Fatal("Failed to unmarshal JSON:", bodyActual, "\nError:", errUnmarshal)
		}

		bodyMapExpected := map[string]any{
			"service_name":              "LuG planets",
			"service_current_date_time": "2000-01-01T00:00:00+00:00",
		}
		if diff := cmp.Diff(bodyMapExpected, bodyMapActual); diff != "" {
			t.Errorf("JSON responses differ (-expected +actual):\n%s", diff)
		}
	})
}

// TestIndexNoSlash Tests a special case for the "index" route, when URI lacks a trailing slash.
func TestIndexNoSlashRedirect(t *testing.T) {
	request, errRequest := http.NewRequest("GET", "/api", nil)
	if errRequest != nil {
		t.Fatal(errRequest)
	}
	response := httptest.NewRecorder()
	api.Router().ServeHTTP(response, request)

	if status := response.Code; status != http.StatusMovedPermanently {
		t.Errorf("Expected HTTP status %v, got %v", http.StatusMovedPermanently, status)
	}

	responseBodyExpected := "<a href=\"/api/\">Moved Permanently</a>.\n\n"
	responseBodyActual := response.Body.String()
	if responseBodyActual != responseBodyExpected {
		t.Errorf("Expected body:\n%v\nGot:\n%v", responseBodyExpected, responseBodyActual)
	}
}
