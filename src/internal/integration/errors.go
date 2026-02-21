package integration

import "errors"

var (
	ErrBodyTooLong                  = errors.New("too long response body")
	ErrDataMissing                  = errors.New("response body lacks required data")
	ErrUnexpectedResponseStatusCode = errors.New("unexpected response status code")
)
