package mock

import "net/http"

type Transport struct {
	RoundTripFunction func(request *http.Request) (*http.Response, error)
}

func (transport *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport.RoundTripFunction(request)
}

func NewTransport() *Transport {
	return &Transport{}
}
