package attendeectl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/eurofurence/reg-attendee-service/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/web/filter/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/eurofurence/reg-attendee-service/web/util/media"
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

func RestDispatcher(router *mux.Router) {
	if config.OptionalInitialRegTokenConfigured() {
		router.HandleFunc("/v1/attendees", filterhelper.BuildHandler("3s", newAttendeeHandler, config.TokenForAdmin, config.OptionalTokenForInitialReg)).Methods(http.MethodPut, http.MethodOptions)
	} else {
		router.HandleFunc("/v1/attendees", filterhelper.BuildUnauthenticatedHandler("3s", newAttendeeHandler)).Methods(http.MethodPut, http.MethodOptions)
	}
	router.HandleFunc("/v1/attendees/max-id", filterhelper.BuildUnauthenticatedHandler("3s", getAttendeeMaxIdHandler)).Methods(http.MethodGet)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", filterhelper.BuildHandler("3s", getAttendeeHandler, config.TokenForAdmin, config.TokenForLoggedInUser)).Methods(http.MethodGet)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", filterhelper.BuildHandler("3s", updateAttendeeHandler, config.TokenForAdmin, config.TokenForLoggedInUser)).Methods(http.MethodPost, http.MethodOptions)
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
	err = mapDtoToAttendee(dto, newAttendee)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
		return
	}
	id, err := attendeeService.RegisterNewAttendee(ctx, newAttendee)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	location := fmt.Sprintf("%s/%d", r.RequestURI, id)
	logging.Ctx(ctx).Info("sending Location " + location)
	w.Header().Set(headers.Location, location)
	writeHeader(ctx, w, http.StatusCreated)
}

func getAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	existingAttendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		attendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	dto := attendee.AttendeeDto{}
	mapAttendeeToDto(existingAttendee, &dto)
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	writeJson(ctx, w, dto)
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
	attendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		attendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	validationErrs := validate(ctx, dto, attendee)
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	err = mapDtoToAttendee(dto, attendee)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
		return
	}
	err = attendeeService.UpdateAttendee(ctx, attendee)
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
	writeJson(ctx, w, dto)
}

func idFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		invalidIdErrorHandler(ctx, w, r, vars["id"])
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
	errorHandler(ctx, w, r, "attendee.data.invalid", http.StatusBadRequest, errs)
}

func invalidIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id string) {
	logging.Ctx(ctx).Warnf("received invalid attendee id '%s'", id)
	errorHandler(ctx, w, r, "attendee.id.invalid", http.StatusBadRequest, url.Values{})
}

func attendeeNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id uint) {
	logging.Ctx(ctx).Warnf("attendee id %v not found", id)
	errorHandler(ctx, w, r, "attendee.id.notfound", http.StatusNotFound, url.Values{})
}

func attendeeParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("attendee body could not be parsed: %v", err)
	errorHandler(ctx, w, r, "attendee.parse.error", http.StatusBadRequest, url.Values{})
}

func attendeeWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("attendee could not be written: %v", err)
	if err.Error() == "duplicate attendee data - you are already registered" {
		errorHandler(ctx, w, r, "attendee.data.duplicate", http.StatusBadRequest, url.Values{"attendee": {"there is already an attendee with this information (looking at nickname, email, and zip code)"}})
	} else {
		errorHandler(ctx, w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
	}
}

func attendeeMaxIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("could not determine max id: %v", err)
	errorHandler(ctx, w, r, "attendee.max_id.error", http.StatusInternalServerError, url.Values{})
}

func errorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string, status int, details url.Values) {
	timestamp := time.Now().Format(time.RFC3339)
	response := attendee.ErrorDto{Message: msg, Timestamp: timestamp, Details: details, RequestId: ctxvalues.RequestId(ctx)}
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
