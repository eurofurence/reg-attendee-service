package attendeectl

import (
	"context"
	"encoding/json"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var attendeeService attendeesrv.AttendeeService

func init() {
	attendeeService = &attendeesrv.AttendeeServiceImplData{}
}

// use only for testing
func OverrideAttendeeService(overrideAttendeeServiceForTesting attendeesrv.AttendeeService) {
	attendeeService = overrideAttendeeServiceForTesting
}

func Create(server chi.Router) {
	if config.RequireLoginForReg() {
		server.Post("/api/rest/v1/attendees", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, newAttendeeHandler)))
	} else {
		server.Post("/api/rest/v1/attendees", filter.WithTimeout(3*time.Second, newAttendeeHandler))
	}
	server.Get("/api/rest/v1/attendees/max-id", filter.WithTimeout(3*time.Second, getAttendeeMaxIdHandler))
	server.Get("/api/rest/v1/attendees/{id}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getAttendeeHandler)))
	server.Put("/api/rest/v1/attendees/{id}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, updateAttendeeHandler)))
}

func newAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dto, err := parseBodyToAttendeeDto(ctx, w, r)
	if err != nil {
		return
	}
	validationErrs := validate(ctx, dto, &entity.Attendee{Flags: config.DefaultFlags(), Packages: config.DefaultPackages(), Options: config.DefaultOptions()})
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	newAttendee := attendeeService.NewAttendee(ctx)
	mapDtoToAttendee(dto, newAttendee)
	id, err := attendeeService.RegisterNewAttendee(ctx, newAttendee)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	location := fmt.Sprintf("%s/%d", r.RequestURI, id)
	aulogging.Logger.Ctx(ctx).Info().Printf("sending Location %s", location)
	w.Header().Set(headers.Location, location)
	w.WriteHeader(http.StatusCreated)
}

func getAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	existingAttendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}

	if err := filter.IsSubjectOrRoleOrApiToken(w, r, existingAttendee.Identity, config.OidcAdminRole()); err != nil {
		return
	}

	dto := attendee.AttendeeDto{}
	mapAttendeeToDto(existingAttendee, &dto)
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func updateAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	dto, err := parseBodyToAttendeeDto(ctx, w, r)
	if err != nil {
		return
	}
	attd, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}

	if err := filter.IsSubjectOrRoleOrApiToken(w, r, attd.Identity, config.OidcAdminRole()); err != nil {
		return
	}

	validationErrs := validate(ctx, dto, attd)
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	mapDtoToAttendee(dto, attd)
	err = attendeeService.UpdateAttendee(ctx, attd)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.Header().Add(headers.Location, r.RequestURI)
}

func getAttendeeMaxIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	max, err := attendeeService.GetAttendeeMaxId(ctx)
	if err != nil {
		attendeeMaxIdErrorHandler(ctx, w, r, err)
		return
	}
	dto := attendee.AttendeeMaxIdDto{}
	dto.MaxId = max
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func idFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctlutil.InvalidAttendeeIdErrorHandler(ctx, w, r, idStr)
	}
	return uint(id), err
}

func parseBodyToAttendeeDto(ctx context.Context, w http.ResponseWriter, r *http.Request) (*attendee.AttendeeDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := &attendee.AttendeeDto{}
	err := decoder.Decode(dto)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

func attendeeValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received attendee data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "attendee.data.invalid", http.StatusBadRequest, errs)
}

func attendeeParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("attendee body could not be parsed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.parse.error", http.StatusBadRequest, url.Values{})
}

func attendeeWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("attendee could not be written: %s", err.Error())
	if err.Error() == "duplicate attendee data - you are already registered" {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.data.duplicate", http.StatusConflict, url.Values{"attendee": {"there is already an attendee with this information (looking at nickname, email, and zip code)"}})
	} else {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
	}
	// TODO: distinguish attendee.payment.error -> bad gateway
}

func attendeeMaxIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("could not determine max id: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.max_id.error", http.StatusInternalServerError, url.Values{})
}
