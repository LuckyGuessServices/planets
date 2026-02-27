package api

import "net/http"

const (
	ResponseErrorInvalidParameters string = "invalid_parameters"
	ResponseErrorDataUnavailable   string = "data_unavailable"

	ResponseParameterErrorInvalid  string = "invalid"
	ResponseParameterErrorRequired string = "required"
)

func HTTPStatusCodeByErrorCode(errorCode string) int {
	switch errorCode {
	case ResponseErrorDataUnavailable:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
