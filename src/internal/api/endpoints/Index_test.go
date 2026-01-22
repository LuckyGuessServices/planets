package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/api"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/helper"
	"github.com/stretchr/testify/assert"
)

// TestIndex Tests a typical request and response.
func TestIndex(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		request := helper.NewHTTPRequest(t, "GET", "/api/", nil)

		response := httptest.NewRecorder()
		api.Router().ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)

		bodyMapExpected := map[string]any{
			"service_name":         "LuG planets",
			"service_current_time": "2000-01-01T00:00:00+00:00",
		}
		helper.RequireJSONResponse(t, bodyMapExpected, response.Body)
	})
}

// TestIndexNoSlash Tests a special case for the "index" route, when URI lacks a trailing slash.
func TestIndexNoSlashRedirect(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		request := helper.NewHTTPRequest(t, "GET", "/api", nil)

		response := httptest.NewRecorder()
		api.Router().ServeHTTP(response, request)

		assert.Equal(t, http.StatusMovedPermanently, response.Code)
		assert.Equal(t, "<a href=\"/api/\">Moved Permanently</a>.\n\n", response.Body.String())
	})
}
