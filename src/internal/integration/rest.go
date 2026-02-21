package integration

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/LuckyGuessServices/planets/internal/luglog"
)

// Adds a random milliseconds count up to this limit to a sleep duration before the next retry.
const sleepAddingLimitMs int = 500

type RESTClient struct {
	*http.Client
}

// Get makes a request.
//
// In case of request timeout, tries to request again after some time (more tries - longer time each try).
// Passing 0 or less as timeoutRetriesMax guarantees a single request try.
func (client *RESTClient) Get(
	ctx context.Context,
	requestURL *url.URL,
	timeoutRetriesMax int8,
) (response *http.Response, requestError error) {
	request, requestError := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if requestError != nil {
		return
	}

	// Request at least once.
	for tryIndex := int8(0); tryIndex == 0 || tryIndex < timeoutRetriesMax; tryIndex++ {
		if tryIndex > 0 {
			sleepAddingMs, _ := rand.Int(rand.Reader, big.NewInt(int64(sleepAddingLimitMs)))
			time.Sleep(
				time.Duration(tryIndex)*time.Second +
					time.Duration(sleepAddingMs.Int64())*time.Millisecond,
			)
		}

		response, requestError = client.Do(request)
		if netError, ok := errors.AsType[net.Error](requestError); ok && netError.Timeout() {
			if response != nil {
				if err := response.Body.Close(); err != nil {
					luglog.Print("Failed to close response body: ", err)
				}
			}

			continue
		}

		break
	}

	return
}

func NewRESTClient() *RESTClient {
	return &RESTClient{Client: &http.Client{}}
}
