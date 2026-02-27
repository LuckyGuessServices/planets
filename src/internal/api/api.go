package api

type Response struct {
	Body           any
	HTTPStatusCode int

	// If > 0, sets "Retry-After" header.
	//
	// Useful for cases like when data is not available immediately, but will be available later.
	RetryAfter int
}

func NewResponse() *Response {
	return &Response{}
}

func NewErrorResponse(errorCode string) *Response {
	return &Response{
		Body: &jsonErrorCode{
			ErrorCode: errorCode,
		},
		HTTPStatusCode: HTTPStatusCodeByErrorCode(errorCode),
	}
}

func NewParameterErrorsResponse(parameterErrorCodes map[string]string) *Response {
	errorCode := ResponseErrorInvalidParameters

	return &Response{
		Body: &struct {
			*jsonErrorCode
			*jsonParameterErrorCodes
		}{
			&jsonErrorCode{ErrorCode: errorCode},
			&jsonParameterErrorCodes{ParameterErrorCodes: parameterErrorCodes},
		},
		HTTPStatusCode: HTTPStatusCodeByErrorCode(errorCode),
	}
}

// NewMessageResponse prepares a JSON-ready response body with the provided message as its only element.
func NewMessageResponse(message string) *Response {
	return &Response{
		Body: &struct {
			*jsonMessage
		}{
			&jsonMessage{Message: message},
		},
	}
}

type jsonErrorCode struct {
	ErrorCode string `json:"error"`
}

type jsonParameterErrorCodes struct {
	ParameterErrorCodes map[string]string `json:"parameter_errors"`
}

type jsonMessage struct {
	Message string `json:"message"`
}
