package ctlutil

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/errorapi"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"time"
)

// --- common error handlers ---

// note, remember to bail out after calling these

func InvalidAttendeeIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid attendee id '%s'", id)
	ErrorHandler(ctx, w, r, "attendee.id.invalid", http.StatusBadRequest, url.Values{})
}

func AttendeeNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id uint) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("attendee id %d not found", id)
	ErrorHandler(ctx, w, r, "attendee.id.notfound", http.StatusNotFound, url.Values{})
}

func ErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string, status int, details url.Values) {
	timestamp := time.Now().Format(time.RFC3339)
	response := errorapi.ErrorDto{Message: msg, Timestamp: timestamp, Details: details, RequestId: ctxvalues.RequestId(ctx)}
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	WriteHeader(ctx, w, status)
	WriteJson(ctx, w, response)
}
