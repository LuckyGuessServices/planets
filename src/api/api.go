package api

import (
	"LuG/planets/api/endpoints"
	"encoding/json"
	"log"
	"net/http"
)

// Router A complete router for the API server.
func Router() *http.ServeMux {
	router := http.NewServeMux()

	// Add sub-routers:
	router.Handle("/api/", http.StripPrefix("/api", RouterGeneral()))

	return router
}

// RouterGeneral Router for general (for instance, not "admin") API endpoints.
func RouterGeneral() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", decorateHandlerJSON(endpoints.Index))

	return router
}

func decorateHandlerJSON(handler func(*http.Request) any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := handler(r)

		responseDataJSON, errMarshal := json.Marshal(responseData)
		if errMarshal != nil {
			w.WriteHeader(http.StatusInternalServerError)

			logHandlerError(errMarshal, "route '"+r.RequestURI+"', marshalling JSON response")

			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeResponse(responseDataJSON, w, r)
	}
}

func writeResponse(responseData []byte, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, errResponse := w.Write(responseData); errResponse != nil {
		logHandlerError(errResponse, "route '"+r.RequestURI+"', writing response")

		return
	}
}

func logHandlerError(err error, errPrefix string) {
	log.Print(errPrefix+" ---> ", err)
}
