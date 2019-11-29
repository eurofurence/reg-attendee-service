package countdownctl

import (
	"context"
	"encoding/json"
	"github.com/eurofurence/reg-attendee-service/api/v1/countdown"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/gorilla/mux"
	"math"
	"net/http"
	"time"
)

func RestDispatcher(router *mux.Router) {
	router.HandleFunc("/v1/countdown", filterhelper.BuildUnauthenticatedHandler("1s", countdownHandler)).Methods(http.MethodGet)
}

const isoDateTimeFormat = "2006-01-02T15:04:05-07:00"

func countdownHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	current := time.Now()
	target := config.RegistrationStartTime()
	secondsToGo := target.Sub(current).Seconds()
	if secondsToGo < 0 {
		secondsToGo = 0
	}

	responseDto := countdown.CountdownResultDto{}
	responseDto.TargetTimeIsoDateTime = target.Format(isoDateTimeFormat)
	responseDto.CurrentTimeIsoDateTime = current.Format(isoDateTimeFormat)
	responseDto.CountdownSeconds = int64(math.Round(secondsToGo))

	writeJson(ctx, w, responseDto)
}

func writeJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		logging.Ctx(ctx).Warnf("error while encoding json response: %v", err)
	}
}

