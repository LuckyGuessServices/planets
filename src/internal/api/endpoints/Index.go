package endpoints

import (
	"net/http"
	"time"

	"github.com/LuckyGuessServices/planets/internal/api"
	"github.com/LuckyGuessServices/planets/internal/general"
)

func Index(_ *http.Request) *api.Response {
	response := api.NewResponse()
	response.Body = &struct {
		ServiceName        string `json:"service_name"`
		ServiceCurrentTime string `json:"service_current_time"`
	}{
		ServiceName:        general.ServiceName,
		ServiceCurrentTime: time.Now().UTC().Format(general.ApiDateTimeTz),
	}

	return response
}
