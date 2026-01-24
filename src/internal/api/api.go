package api

type Response struct {
	HTTPStatusCode int
	Body           any
}

func NewResponse() *Response {
	return &Response{}
}

func NewErrorResponse(errorCode string) *Response {
	return &Response{
		HTTPStatusCode: HTTPStatusCodeByErrorCode(errorCode),
		Body: &jsonErrorCode{
			ErrorCode: errorCode,
		},
	}
}

func NewParameterErrorsResponse(parameterErrorCodes map[string]string) *Response {
	errorCode := ErrInvalidParameters

	return &Response{
		HTTPStatusCode: HTTPStatusCodeByErrorCode(errorCode),
		Body: &struct {
			jsonErrorCode
			jsonParameterErrorCodes
		}{
			jsonErrorCode{ErrorCode: errorCode},
			jsonParameterErrorCodes{ParameterErrorCodes: parameterErrorCodes},
		},
	}
}

type jsonErrorCode struct {
	ErrorCode string `json:"error"`
}

type jsonParameterErrorCodes struct {
	ParameterErrorCodes map[string]string `json:"parameter_errors"`
}
