package statusctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
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

	latest, err := obtainAttendeeLatestStatusMustReturnOnError(ctx, w, r, att)
	if err != nil {
		return
	}

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
	dto, err := parseBodyToStatusChangeDto(ctx, w, r)
	if err != nil {
		return
	}
	latestStatusChange, err := obtainAttendeeLatestStatusMustReturnOnError(ctx, w, r, att)
	if err != nil {
		return
	}

	validationErrs := validate(ctx, latestStatusChange.Status, dto)
	if len(validationErrs) != 0 {
		statusChangeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}

	if err = attendeeService.StatusChangeAllowed(ctx, latestStatusChange.Status, dto.Status); err != nil {
		statusChangeForbiddenErrorHandler(ctx, w, r, err)
		return
	}

	if err = attendeeService.StatusChangePossible(ctx, att, latestStatusChange.Status, dto.Status); err != nil {
		if errors.Is(err, paymentservice.DownstreamError) || errors.Is(err, mailservice.DownstreamError) {
			statusChangeDownstreamError(ctx, w, r, err)
			return
		}
		statusChangeUnavailableErrorHandler(ctx, w, r, err)
		return
	}

	err = attendeeService.UpdateDuesAndDoStatusChangeIfNeeded(ctx, att, latestStatusChange.Status, dto.Status, dto.Comment)
	if err != nil {
		if errors.Is(err, paymentservice.DownstreamError) || errors.Is(err, mailservice.DownstreamError) {
			statusChangeDownstreamError(ctx, w, r, err)
		} else {
			statusWriteErrorHandler(ctx, w, r, err)
		}
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

	mappedHistory := make([]status.StatusChangeDto, 0)
	for _, h := range history {
		mappedHistory = append(mappedHistory, status.StatusChangeDto{
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

func statusWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("could not obtain status history: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "status.write.error", http.StatusInternalServerError, url.Values{})
}

func statusParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("status change body could not be parsed: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "status.parse.error", http.StatusBadRequest, url.Values{})
}

func statusChangeValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	logging.Ctx(ctx).Warnf("received status change data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "status.data.invalid", http.StatusBadRequest, errs)
}

func statusChangeForbiddenErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	// TODO log user so we can figure out who tried it
	logging.Ctx(ctx).Warnf("forbidden status change attempted: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "auth.forbidden", http.StatusForbidden, url.Values{"details": []string{err.Error()}})
}

func statusChangeUnavailableErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("unavailable status change attempted: %v", err)
	message := "status.data.invalid"
	if errors.Is(err, attendeesrv.SameStatusError) {
		message = "status.unchanged.invalid"
	} else if errors.Is(err, attendeesrv.InsufficientPaymentError) {
		message = "status.unpaid.dues"
	} else if errors.Is(err, attendeesrv.HasPaymentBalanceError) {
		message = "status.has.paid"
	} else if errors.Is(err, attendeesrv.CannotDeleteError) {
		message = "status.cannot.delete"
	} else if errors.Is(err, attendeesrv.GoToApprovedFirst) {
		message = "status.use.approved"
	}
	ctlutil.ErrorHandler(ctx, w, r, message, http.StatusConflict, url.Values{"details": []string{err.Error()}})
}

func statusChangeDownstreamError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("downstream error during status change: %v", err)
	message := "unknown"
	if errors.Is(err, paymentservice.DownstreamError) {
		message = "status.payment.error"
	} else if errors.Is(err, mailservice.DownstreamError) {
		message = "status.mail.error"
	}
	ctlutil.ErrorHandler(ctx, w, r, message, http.StatusBadGateway, url.Values{"details": []string{err.Error()}})
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

func obtainAttendeeLatestStatusMustReturnOnError(ctx context.Context, w http.ResponseWriter, r *http.Request, att *entity.Attendee) (entity.StatusChange, error) {
	history, err := attendeeService.GetFullStatusHistory(ctx, att)
	if err != nil {
		statusReadErrorHandler(ctx, w, r, err)
		return entity.StatusChange{}, err
	} else if len(history) == 0 {
		err := errors.New("got empty status change history")
		statusReadErrorHandler(ctx, w, r, err)
		return entity.StatusChange{}, err
	}

	latest := history[len(history)-1]
	return latest, nil
}

func parseBodyToStatusChangeDto(ctx context.Context, w http.ResponseWriter, r *http.Request) (*status.StatusChangeDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := &status.StatusChangeDto{}
	err := decoder.Decode(dto)
	if err != nil {
		statusParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}
