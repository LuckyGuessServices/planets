package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/integration"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// Tests cases when a response body may be longer than a specified limit.
//
// See integration.ReadLimitedBody
func TestResponseBodyLength(t *testing.T) {
	tests := []struct {
		name                 string
		bodyLimitBytes       int64
		expectedBody         []byte
		isLimitErrorExpected bool
	}{
		{
			name:                 "within-limit",
			bodyLimitBytes:       17,
			expectedBody:         []byte("Cats detected: 16"),
			isLimitErrorExpected: false,
		},
		{
			name:                 "limit-crossed",
			bodyLimitBytes:       16,
			expectedBody:         []byte("Cats detected: 1"),
			isLimitErrorExpected: true,
		},
		{
			name:                 "too-wide-limit",
			bodyLimitBytes:       100500,
			expectedBody:         []byte("Cats detected: 16"),
			isLimitErrorExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("Cats detected: 16")),
				}

				actualBody, errRead := integration.ReadLimitedBody(response, test.bodyLimitBytes)
				if test.isLimitErrorExpected {
					assert.ErrorIs(t, errRead, integration.ErrBodyTooLong)
				} else {
					require.NoError(t, errRead)
				}
				assert.Equal(t, string(test.expectedBody), string(actualBody))
			})
		})
	}
}
