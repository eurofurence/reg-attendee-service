package attendeectl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/eurofurence/reg-attendee-service/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"strconv"
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
	if config.OptionalInitialRegTokenConfigured() {
		server.Post("/api/rest/v1/attendees", filterhelper.BuildHandler("3s", newAttendeeHandler, config.TokenForAdmin, config.OptionalTokenForInitialReg))
	} else {
		server.Post("/api/rest/v1/attendees", filterhelper.BuildUnauthenticatedHandler("3s", newAttendeeHandler))
	}

	server.Get("/api/rest/v1/attendees/max-id", filterhelper.BuildUnauthenticatedHandler("3s", getAttendeeMaxIdHandler))
	server.Get("/api/rest/v1/attendees/{id}", filterhelper.BuildHandler("3s", getAttendeeHandler, config.TokenForAdmin, config.TokenForLoggedInUser))
	server.Put("/api/rest/v1/attendees/{id}", filterhelper.BuildHandler("3s", updateAttendeeHandler, config.TokenForAdmin, config.TokenForLoggedInUser))
}

func newAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
	// TODO react to duplicate by sending 409 instead
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	location := fmt.Sprintf("%s/%d", r.RequestURI, id)
	logging.Ctx(ctx).Info("sending Location " + location)
	w.Header().Set(headers.Location, location)
	ctlutil.WriteHeader(ctx, w, http.StatusCreated)
}

func getAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	existingAttendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	dto := attendee.AttendeeDto{}
	mapAttendeeToDto(existingAttendee, &dto)
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func updateAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

func getAttendeeMaxIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
	logging.Ctx(ctx).Warnf("received attendee data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "attendee.data.invalid", http.StatusBadRequest, errs)
}

func attendeeParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("attendee body could not be parsed: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "attendee.parse.error", http.StatusBadRequest, url.Values{})
}

func attendeeWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("attendee could not be written: %v", err)
	if err.Error() == "duplicate attendee data - you are already registered" {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.data.duplicate", http.StatusBadRequest, url.Values{"attendee": {"there is already an attendee with this information (looking at nickname, email, and zip code)"}})
	} else {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
	}
}

func attendeeMaxIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("could not determine max id: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "attendee.max_id.error", http.StatusInternalServerError, url.Values{})
}
