package integration_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/LuckyGuessServices/planets/internal/integration"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/LuckyGuessServices/planets/internal/test_tools/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests cases when a response is received on time or timeouts, and then client retries the same request.
//
// See RESTClient.Get
func TestRetryIfRequestTimeouts(t *testing.T) {
	var triesMade int8
	transport := mock.NewTransport()
	transport.RoundTripFunction = func(request *http.Request) (*http.Response, error) {
		triesMade++
		waitFor := time.Hour
		if triesMade > 2 {
			waitFor = 0
		}

		select {
		case <-request.Context().Done():
			return nil, request.Context().Err()
		case <-time.After(waitFor):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("Some response.")),
				Header:     make(http.Header),
			}, nil
		}
	}

	client := integration.NewRESTClient()
	client.Transport = transport
	client.Timeout = time.Minute

	tests := []struct {
		name                      string
		retriesMax                int8
		isTimeoutErrorExpected    bool
		expectedCurrentTimeString string
	}{
		{
			name:                      "enough-tries-for-ok",
			retriesMax:                3,
			isTimeoutErrorExpected:    false,
			expectedCurrentTimeString: "2000-01-01T00:02:03Z",
		},
		{
			name:                      "timeout-single-try",
			retriesMax:                1,
			isTimeoutErrorExpected:    true,
			expectedCurrentTimeString: "2000-01-01T00:01:00Z",
		},
		{
			name:                      "timeout-more-tries",
			retriesMax:                2,
			isTimeoutErrorExpected:    true,
			expectedCurrentTimeString: "2000-01-01T00:02:01Z",
		},
		{
			name:                      "zero-tries-as-single-try",
			retriesMax:                0,
			isTimeoutErrorExpected:    true,
			expectedCurrentTimeString: "2000-01-01T00:01:00Z",
		},
		{
			name:                      "negative-tries-as-single-try",
			retriesMax:                -1,
			isTimeoutErrorExpected:    true,
			expectedCurrentTimeString: "2000-01-01T00:01:00Z",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				// Reset the tries counter for each subtest:
				triesMade = 0

				response, errRequest := client.Get(context.Background(), &url.URL{}, test.retriesMax)
				// A response body (even if empty) must be always closed eventually.
				// If not done, the request watching goroutine will wait endlessly until a context may be canceled.
				if response != nil {
					require.NoError(t, response.Body.Close())
				}

				if test.isTimeoutErrorExpected {
					require.Error(t, errRequest)
					errNet, isErrorOfTypeNet := errors.AsType[net.Error](errRequest)
					require.True(t, isErrorOfTypeNet, errRequest)
					assert.Truef(t, errNet.Timeout(), errNet.Error())
				} else {
					require.NoError(t, errRequest)
				}

				expectedCurrentTime, errTimeParse := time.Parse(time.RFC3339, test.expectedCurrentTimeString)
				require.NoError(t, errTimeParse)

				actualCurrentTime := time.Now().UTC()
				if test.retriesMax > 1 {
					assert.LessOrEqual(t, expectedCurrentTime, actualCurrentTime)
					assert.Greater(t, expectedCurrentTime.Add(time.Second), actualCurrentTime)
				} else {
					assert.Equal(t, expectedCurrentTime, actualCurrentTime)
				}
			})
		})
	}
}
