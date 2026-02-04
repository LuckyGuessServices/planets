package endpoints_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/api/router"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/helper"
	"github.com/stretchr/testify/assert"
)

// TestIndex tests a typical request and response.
func TestIndex(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		response := helper.QueryApi(t, "GET", "/")

		assert.Equal(t, http.StatusOK, response.Code)

		bodyExpected := map[string]any{
			"service_name":         "lug-planets",
			"service_current_time": "2000-01-01T00:00:00Z",
		}
		helper.RequireJSONResponse(t, bodyExpected, response.Body)
	})
}

// TestIndexNoSlash Tests a special case for the "index" route, when URI lacks a trailing slash.
func TestIndexNoSlashRedirect(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		response := helper.QueryApi(t, "GET", "")

		assert.Equal(t, http.StatusMovedPermanently, response.Code)
		assert.Equal(
			t,
			fmt.Sprintf("<a href=\"%s/\">Moved Permanently</a>.\n\n", router.PrefixGeneral),
			response.Body.String(),
		)
	})
}
