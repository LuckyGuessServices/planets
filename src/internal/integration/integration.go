package integration

import (
	"fmt"
	"io"
	"net/http"
)

func ReadLimitedBody(response *http.Response, limitBytes int64) (responseBody []byte, readError error) {
	responseBody, readError = io.ReadAll(io.LimitReader(response.Body, limitBytes+1))
	if readError != nil {
		return
	}
	if int64(len(responseBody)) > limitBytes {
		responseBody = responseBody[:limitBytes]
		readError = fmt.Errorf("%w; the part being read: '%s'", ErrBodyTooLong, responseBody)
	}

	return
}
