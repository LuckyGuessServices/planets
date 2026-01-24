package api

import "net/http"

const (
	ErrInvalidParameters string = "invalid_parameters"
)

func HTTPStatusCodeByErrorCode(errorCode string) int {
	switch errorCode {
	default:
		return http.StatusBadRequest
	}
}
