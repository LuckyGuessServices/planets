package endpoints

import (
	"LuG/planets/common"
	"net/http"
	"time"
)

func Index(_ *http.Request) any {
	return struct {
		ServiceName        string `json:"service_name"`
		ServiceCurrentTime string `json:"service_current_time"`
	}{
		ServiceName:        common.ServiceName,
		ServiceCurrentTime: time.Now().UTC().Format(common.ApiDateTimeTz),
	}
}
