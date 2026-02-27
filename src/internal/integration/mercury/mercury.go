package mercury

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/integration"
	"github.com/LuckyGuessServices/planets/internal/luglog"
)

const (
	// See docs at https://mercuryretrogradeapi.com/about.html
	apiBaseURL string = "https://mercuryretrogradeapi.com"

	requestTimeout         = 1 * time.Second
	timeoutRetriesMax int8 = 3

	// We should not allow reading input data endlessly. Also, a normal response should contain 23 bytes at max.
	// Let's have some room for additional bytes, in case some other data is added in the future,
	// but a total response length does not become abnormal.
	responseBodyLimitBytes int64 = 256
)

type APIClient struct {
	restClient *integration.RESTClient

	baseURL string
}

// newClient creates a new instance. This function should be called manually mainly for tests.
// On production, you should call Client instead.
//
// Pass "" as baseURL, if not going to specify httptest.Server URL.
//
// Pass nil as transport to invoke http.DefaultTransport, if not going to mock it.
func newClient(baseURL string, transport http.RoundTripper) (client *APIClient) {
	client = &APIClient{
		restClient: integration.NewRESTClient(),
		baseURL:    baseURL,
	}
	client.restClient.Timeout = requestTimeout
	client.restClient.Transport = transport

	return
}

var (
	apiClientMutex sync.Mutex
	apiClient      *APIClient
)

type brokenTransport struct{}

func (transport *brokenTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	errorMessage := fmt.Sprintf(
		`An API client must be mocked in "%s" environment. Init a mock client instance with "%s".`,
		env.Config().ApplicationEnvironment(),
		"SetMockClientForTests",
	)

	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader(errorMessage)),
		Header:     make(http.Header),
	}, nil
}

// SetMockClientForTests replaces (or fills if empty) current internal client instance with a mock.
// Then calling Client returns the mock client.
//
// Pass "" as baseURL, if not going to specify httptest.Server URL.
//
// Pass nil as transport to invoke http.DefaultTransport, if not going to mock it.
func SetMockClientForTests(baseURL string, transport http.RoundTripper) {
	apiClientMutex.Lock()
	defer apiClientMutex.Unlock()

	env.PanicIfEnvNotTest()

	apiClient = newClient(baseURL, transport)
}

// UnsetClientForTests clears current internal client instance.
//
// Always call it as a deferred function inside tests. This way you ensure the next test will not reuse
// the same client with the same mocked transport.
func UnsetClientForTests() {
	apiClientMutex.Lock()
	defer apiClientMutex.Unlock()

	env.PanicIfEnvNotTest()

	apiClient = nil
}

// Client consequently returns the same API client instance initialized during the first function call.
//
// Call SetMockClientForTests to init a mock instance.
func Client() *APIClient {
	apiClientMutex.Lock()
	defer apiClientMutex.Unlock()

	if apiClient != nil {
		return apiClient
	}

	if env.IsTest() {
		apiClient = newClient("", &brokenTransport{})
	} else {
		apiClient = newClient(apiBaseURL, nil)
	}

	return apiClient
}

func (client *APIClient) ByDate(ctx context.Context, date types.Date) (isRetrograde bool, processingError error) {
	const requestPath string = "/"
	requestURLParameters := &url.Values{}
	requestURLParameters.Set("date", date.String())

	response, processingError := client.makeGetRequest(ctx, requestPath, requestURLParameters)
	defer func() { closeResponseBody(response) }()
	if processingError != nil {
		return
	}

	responseBody, processingError := integration.ReadLimitedBody(response, responseBodyLimitBytes)
	if processingError != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		processingError = fmt.Errorf(
			"%w: %d; raw response body (within quotes): '%s'",
			integration.ErrUnexpectedResponseStatusCode,
			response.StatusCode,
			responseBody,
		)

		return
	}

	responseData := &struct {
		IsRetrograde types.Nullable[bool] `json:"is_retrograde"`
	}{}
	processingError = json.Unmarshal(responseBody, responseData)
	if processingError != nil {
		processingError = fmt.Errorf("%w; raw contents (within quotes): '%s'", processingError, responseBody)

		return
	}

	if !responseData.IsRetrograde.Valid {
		processingError = fmt.Errorf("%w; raw contents (within quotes): '%s'", integration.ErrDataMissing, responseBody)

		return
	}
	isRetrograde = responseData.IsRetrograde.V

	return
}

func (client *APIClient) makeGetRequest(
	ctx context.Context,
	endpointPath string,
	queryParameters *url.Values,
) (response *http.Response, requestError error) {
	requestURL, requestError := url.Parse(client.baseURL)
	if requestError != nil {
		return
	}

	requestURL = requestURL.JoinPath(endpointPath)
	requestURL.RawQuery = queryParameters.Encode()

	response, requestError = client.restClient.Get(ctx, requestURL, timeoutRetriesMax)

	return
}

func closeResponseBody(response *http.Response) {
	if response == nil {
		return
	}

	if err := response.Body.Close(); err != nil {
		luglog.Print("Failed to close Mercury API response body: ", err)
	}
}
