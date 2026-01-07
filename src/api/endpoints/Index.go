package endpoints

import (
	"LuG/planets/common"
	"net/http"
	"time"
)

func Index(_ *http.Request) any {
	return struct {
		ServiceName            string `json:"service_name"`
		ServiceCurrentDateTime string `json:"service_current_date_time"`
	}{
		ServiceName:            common.ServiceName,
		ServiceCurrentDateTime: time.Now().UTC().Format(common.ApiDateTimeTz),
	}
}
