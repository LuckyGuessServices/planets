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

// Tests a typical request and response.
//
// See endpoints.Index
func TestIndex(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		response := helper.QueryApiGet(t, "/", nil).Result()
		defer func() { _ = response.Body.Close() }()

		assert.Equal(t, http.StatusOK, response.StatusCode)

		bodyExpected := map[string]any{
			"service_name":         "lug-planets",
			"service_current_time": "2000-01-01T00:00:00Z",
		}
		helper.RequireJSONResponse(t, bodyExpected, helper.ExtractAndCloseResponseBody(t, response))
	})
}

// Tests a special case for the "index" route, when URI lacks a trailing slash.
//
// See endpoints.Index
func TestIndexNoSlashRedirect(t *testing.T) {
	test_tools.RunInSyncBubble(t, func(t *testing.T) {
		response := helper.QueryApiGet(t, "", nil).Result()
		defer func() { _ = response.Body.Close() }()

		assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode)
		assert.Equal(
			t,
			fmt.Sprintf("<a href=\"%s/\">Temporary Redirect</a>.\n\n", router.PrefixGeneral),
			string(helper.ExtractAndCloseResponseBody(t, response)),
		)
	})
}
