package endpoints_test

import (
	"net/http"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/helper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

func TestUnknownEndpoint(t *testing.T) {
	response := helper.QueryApiGet(t, "/unknown", nil).Result()
	defer func() { _ = response.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.Equal(t, "404 page not found\n", string(helper.ExtractAndCloseResponseBody(t, response)))
}
