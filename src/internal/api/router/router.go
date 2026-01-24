package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/LuckyGuessServices/planets/internal/api"
	"github.com/LuckyGuessServices/planets/internal/api/endpoints"
	"github.com/LuckyGuessServices/planets/internal/paths/general"
)

const (
	PrefixGeneral string = "/api"
)

// Create returns a new complete (composed) router instance for the API server.
func Create() *http.ServeMux {
	router := http.NewServeMux()

	// Add sub-routers:
	router.Handle(PrefixGeneral+"/", createGeneral())

	return router
}

// createGeneral returns a sub-router for general API endpoints.
//
// An opposite example of non-general endpoints:
// "admin"-related endpoints should be available by a different sub-router.
func createGeneral() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET "+PrefixGeneral+"/{$}", decorateHandlerJSON(endpoints.Index))

	return router
}

func decorateHandlerJSON(handler func(*http.Request) *api.Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if panicVal := recover(); panicVal != nil {
				w.WriteHeader(http.StatusInternalServerError)
				panicPackage, panicFile := general.PanicSource()
				panicString := fmt.Sprintf(
					"%#v; paniced function: %s\npanicked path: %s",
					panicVal,
					panicPackage,
					panicFile,
				)
				newHandlerErrorLogger(errors.New(panicString), r).Prefix("panic").Log()
			}
		}()

		response := handler(r)

		if response.HTTPStatusCode > 0 {
			w.WriteHeader(response.HTTPStatusCode)
		}

		if response.Body != nil {
			responseDataJSON, errMarshal := json.Marshal(response.Body)
			if errMarshal != nil {
				w.WriteHeader(http.StatusInternalServerError)
				newHandlerErrorLogger(errMarshal, r).
					Prefix("marshaling response into JSON").
					Extra(fmt.Sprintf("raw body: %#v", response.Body)).
					Log()

				return
			}

			w.Header().Set("Content-Type", "application/json")
			writeResponse(responseDataJSON, w, r)
		}
	}
}

func writeResponse(responseData []byte, w http.ResponseWriter, r *http.Request) {
	if _, errResponse := w.Write(responseData); errResponse != nil {
		newHandlerErrorLogger(errResponse, r).Prefix("writing response").Log()

		return
	}
}
