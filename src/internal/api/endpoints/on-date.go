package endpoints

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LuckyGuessServices/planets/internal/api"
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot/storage"
	"github.com/LuckyGuessServices/planets/internal/general/types"
)

// OnDateRetryAfterSeconds reflects when it is reasonable to retry OnDate endpoint again.
const OnDateRetryAfterSeconds = 60

type onDateResponseSnapshotsData struct {
	PlanetIndex  int16 `json:"planet_index"`
	IsRetrograde bool  `json:"is_retrograde"`
}

func OnDate(request *http.Request) *api.Response {
	requestParameters := request.URL.Query()

	const parameterNameDate string = "date"
	if !requestParameters.Has(parameterNameDate) {
		return api.NewParameterErrorsResponse(map[string]string{
			parameterNameDate: api.ResponseParameterErrorRequired,
		})
	}

	// Strangely the internal parser sometimes works well with quotes and other times returns an error.
	// So let's trim possible quotes to be sure.
	dateString := strings.Trim(requestParameters.Get(parameterNameDate), `"`)
	var errDateParse error
	var dateValidated types.Date
	if len(dateString) > len(time.DateOnly) {
		errDateParse = errors.New("expected date only, got something else")
	} else {
		dateValidated, errDateParse = types.NewDateParsed(dateString)
	}
	if errDateParse != nil {
		return api.NewParameterErrorsResponse(map[string]string{
			parameterNameDate: api.ResponseParameterErrorInvalid,
		})
	}

	snapshotsTable := storage.SnapshotRepository()

	snapshots, errSnapshots := snapshotsTable.BySnapshotDate(request.Context(), dateValidated)
	if errSnapshots != nil {
		panic(fmt.Sprintf(`Failed to request snapshots by date "%s": %v`, dateValidated, errSnapshots))
	}

	snapshotsCount := len(snapshots)
	shouldReturnDataUnavailable := snapshotsCount > 0
	responseData := make([]*onDateResponseSnapshotsData, 0, snapshotsCount)
	for _, snapshot := range snapshots {
		if !snapshot.IsRetrogradeRaw().Valid {
			continue
		}

		shouldReturnDataUnavailable = false
		responseData = append(
			responseData,
			&onDateResponseSnapshotsData{
				PlanetIndex:  snapshot.PlanetIndex(),
				IsRetrograde: snapshot.IsRetrogradeRaw().V,
			},
		)
	}

	if snapshotsCount <= 0 {
		snapshotsTable.ObtainSnapshot(dateValidated)

		response := api.NewMessageResponse(
			"Request accepted, no data available yet. Retry after %d seconds." +
				strconv.Itoa(OnDateRetryAfterSeconds) + " seconds.",
		)
		response.HTTPStatusCode = http.StatusAccepted
		response.RetryAfter = OnDateRetryAfterSeconds

		return response
	}

	if shouldReturnDataUnavailable {
		return api.NewErrorResponse(api.ResponseErrorDataUnavailable)
	}

	response := api.NewResponse()
	response.Body = &struct {
		Snapshots []*onDateResponseSnapshotsData `json:"snapshots"`
	}{Snapshots: responseData}

	return response
}
