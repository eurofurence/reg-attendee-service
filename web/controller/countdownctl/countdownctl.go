package countdownctl

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/api/v1/countdown"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/eurofurence/reg-attendee-service/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"math"
	"net/http"
	"time"
)

func Create(server chi.Router) {
	server.Get("/api/rest/v1/countdown", filterhelper.BuildUnauthenticatedHandler("1s", countdownHandler))
}

const isoDateTimeFormat = "2006-01-02T15:04:05-07:00"

func countdownHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	currentStr := r.URL.Query().Get("currentTime")
	if currentStr == "" {
		realCountdownHandler(ctx, w, r)
	} else {
		mockCountdownHandler(ctx, w, r, currentStr)
	}
}

func realCountdownHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	current := time.Now()
	commonCountdownHandler(ctx, w, r, current)
}

func mockCountdownHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, currentStr string) {
	// added for ease of testing0
	logging.Ctx(ctx).Info("used mock with currentTime=" + currentStr)
	current, err := time.Parse(config.StartTimeFormat, currentStr)
	if err != nil {
		// ignore unparseable date and use actual time instead (this is only for testing calls anyway)
		logging.Ctx(ctx).Warn("used mock with unparseable currentTime, using current time instead")
		current = time.Now()
	}
	commonCountdownHandler(ctx, w, r, current)
}

func commonCountdownHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, current time.Time) {
	target := config.RegistrationStartTime()
	secondsToGo := target.Sub(current).Seconds()
	if secondsToGo < 0 {
		secondsToGo = 0
	}

	responseDto := countdown.CountdownResultDto{}
	responseDto.TargetTimeIsoDateTime = target.Format(isoDateTimeFormat)
	responseDto.CurrentTimeIsoDateTime = current.Format(isoDateTimeFormat)
	responseDto.CountdownSeconds = int64(math.Round(secondsToGo))

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, responseDto)
}
