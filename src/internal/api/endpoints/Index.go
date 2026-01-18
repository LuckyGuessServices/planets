package endpoints

import (
	"net/http"
	"time"

	"github.com/LuckyGuessServices/planets/internal/general"
)

func Index(_ *http.Request) any {
	return struct {
		ServiceName        string `json:"service_name"`
		ServiceCurrentTime string `json:"service_current_time"`
	}{
		ServiceName:        general.ServiceName,
		ServiceCurrentTime: time.Now().UTC().Format(general.ApiDateTimeTz),
	}
}
