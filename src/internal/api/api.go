package api

type Response struct {
	Body           any
	HTTPStatusCode int
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
	errorCode := ErrInvalidParameters

	return &Response{
		Body: &struct {
			jsonErrorCode
			jsonParameterErrorCodes
		}{
			jsonErrorCode{ErrorCode: errorCode},
			jsonParameterErrorCodes{ParameterErrorCodes: parameterErrorCodes},
		},
		HTTPStatusCode: HTTPStatusCodeByErrorCode(errorCode),
	}
}

type jsonErrorCode struct {
	ErrorCode string `json:"error"`
}

type jsonParameterErrorCodes struct {
	ParameterErrorCodes map[string]string `json:"parameter_errors"`
}
