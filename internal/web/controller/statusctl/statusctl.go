package statusctl

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/filterhelper"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
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
	att, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	// TODO ensure if user, can only get their own data - once permission system is in
	// (right now regular users and staff are completely forbidden, but they'll need this)

	history, err := attendeeService.GetFullStatusHistory(ctx, att)
	if err != nil {
		statusReadErrorHandler(ctx, w, r, err)
		return
	} else if len(history) == 0 {
		statusReadErrorHandler(ctx, w, r, errors.New("got empty status change history"))
		return
	}

	latest := history[len(history)-1]
	dto := status.StatusDto{
		Status: latest.Status,
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func postStatusHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	att, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	err = attendeeService.RequestStatusChange(ctx, att, "approved", "")
	// TODO various error handlers
	if err != nil {

	} else {
		ctlutil.WriteHeader(ctx, w, http.StatusNoContent)
	}
}

func getStatusHistoryHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	att, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	history, err := attendeeService.GetFullStatusHistory(ctx, att)
	if err != nil {
		statusReadErrorHandler(ctx, w, r, err)
		return
	} else if len(history) == 0 {
		statusReadErrorHandler(ctx, w, r, errors.New("got empty status change history"))
		return
	}

	mappedHistory := make([]status.StatusChange, 0)
	for _, h := range history {
		mappedHistory = append(mappedHistory, status.StatusChange{
			Timestamp: h.CreatedAt.Format(time.RFC3339),
			Status:    h.Status,
			Comment:   h.Comments,
		})
	}
	dto := status.StatusHistoryDto{
		Id:            fmt.Sprintf("%d", att.ID),
		StatusHistory: mappedHistory,
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

// --- error handlers ---

func statusReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("could not obtain status history: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "status.read.error", http.StatusInternalServerError, url.Values{})
}

// --- helpers ---

func attendeeByIdMustReturnOnError(ctx context.Context, w http.ResponseWriter, r *http.Request) (*entity.Attendee, error) {
	id, err := ctlutil.AttendeeIdFromVars(ctx, w, r)
	if err != nil {
		return &entity.Attendee{}, err
	}
	attendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return &entity.Attendee{}, err
	}
	return attendee, nil
}
