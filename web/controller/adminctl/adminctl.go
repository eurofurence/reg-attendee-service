package adminctl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/web/filter/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/eurofurence/reg-attendee-service/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var attendeeService attendeesrv.AttendeeService

// TODO we should not wire this up here
func init() {
	attendeeService = &attendeesrv.AttendeeServiceImplData{}
}

// use only for testing
func OverrideAttendeeService(overrideAttendeeServiceForTesting attendeesrv.AttendeeService) {
	attendeeService = overrideAttendeeServiceForTesting
}

func Create(server chi.Router) {
	server.Get("/api/rest/v1/attendees/{id:[1-9][0-9]*}/admin", filterhelper.BuildHandler("3s", getAdminInfoHandler, config.TokenForAdmin))
}

// --- handlers ---

func getAdminInfoHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	_, err = attendeeService.GetAttendee(ctx, id)
	if err != nil {
		attendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	// TODO get data from service instead of defaults
	// TODO mapAttendeeToDto(existingAttendee, &dto)
	timestamp := time.Now().Format(time.RFC3339)
	dto := admin.AdminInfoDto{
		Id: fmt.Sprintf("%d", id),
		StatusHistory: []admin.StatusChange{{
			Timestamp: timestamp,
			Status:    "new",
		}},
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	writeJson(ctx, w, dto)
}

// --- parameter parsers ---

func idFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		invalidIdErrorHandler(ctx, w, r, idStr)
	}
	return uint(id), err
}

// --- error handlers ---

func invalidIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id string) {
	logging.Ctx(ctx).Warnf("received invalid attendee id '%s'", id)
	errorHandler(ctx, w, r, "attendee.id.invalid", http.StatusBadRequest, url.Values{})
}

func attendeeNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id uint) {
	logging.Ctx(ctx).Warnf("attendee id %v not found", id)
	errorHandler(ctx, w, r, "attendee.id.notfound", http.StatusNotFound, url.Values{})
}

func errorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string, status int, details url.Values) {
	timestamp := time.Now().Format(time.RFC3339)
	response := admin.ErrorDto{Message: msg, Timestamp: timestamp, Details: details, RequestId: ctxvalues.RequestId(ctx)}
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	writeHeader(ctx, w, status)
	writeJson(ctx, w, response)
}

func writeJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		logging.Ctx(ctx).Warnf("error while encoding json response: %v", err)
	}
}

func writeHeader(ctx context.Context, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	ctxvalues.SetHttpStatus(ctx, status)
}
