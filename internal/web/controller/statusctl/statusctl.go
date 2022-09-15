package statusctl

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/filterhelper"
	ctlutil2 "github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
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
	server.Get("/api/rest/v1/attendees/{id}/status", filterhelper.BuildHandler("3s", getStatusHandler, config.TokenForAdmin))
	server.Post("/api/rest/v1/attendees/{id}/status", filterhelper.BuildHandler("3s", postStatusHandler, config.TokenForAdmin))
	server.Get("/api/rest/v1/attendees/{id}/status-history", filterhelper.BuildHandler("3s", getStatusHistoryHandler, config.TokenForAdmin))
}

// --- handlers ---

func getStatusHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	_, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	// TODO ensure if user, can only get their own data

	// TODO get data from service instead of defaults
	// TODO mapAttendeeToDto(existingAttendee, &dto)
	dto := status.StatusDto{
		Status: "new",
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil2.WriteJson(ctx, w, dto)
}

func postStatusHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	_, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	// TODO implement me

	// TODO get data from service instead of defaults
	// TODO mapAttendeeToDto(existingAttendee, &dto)
	dto := status.StatusDto{
		Status: "new",
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil2.WriteJson(ctx, w, dto)
}

func getStatusHistoryHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	attendee, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}
	// TODO get data from service instead of defaults
	// TODO mapAttendeeToDto(existingAttendee, &dto)
	timestamp := time.Now().Format(time.RFC3339)
	dto := status.StatusHistoryDto{
		Id: fmt.Sprintf("%d", attendee.ID),
		StatusHistory: []status.StatusChange{{
			Timestamp: timestamp,
			Status:    "new",
		}},
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil2.WriteJson(ctx, w, dto)
}

// --- helpers ---

func attendeeByIdMustReturnOnError(ctx context.Context, w http.ResponseWriter, r *http.Request) (*entity.Attendee, error) {
	id, err := ctlutil2.AttendeeIdFromVars(ctx, w, r)
	if err != nil {
		return &entity.Attendee{}, err
	}
	attendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil2.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return &entity.Attendee{}, err
	}
	return attendee, nil
}
