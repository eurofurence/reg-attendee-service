package attendeectl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/jumpy-squirrel/rexis-go-attendee/api/v1/attendee"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/service/attendeesrv"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/filterhelper"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/util/media"
	"log"
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
	router.HandleFunc("/v1/attendees", filterhelper.BuildUnauthenticatedHandler("3s", newAttendeeHandler)).Methods(http.MethodPut)
	// TODO authorization missing for these
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", filterhelper.BuildUnauthenticatedHandler("3s", getAttendeeHandler)).Methods(http.MethodGet)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", filterhelper.BuildUnauthenticatedHandler("3s", updateAttendeeHandler)).Methods(http.MethodPost)
}

func newAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	dto, err := parseBodyToAttendeeDto(ctx, w, r)
	if err != nil {
		return
	}
	validationErrs := validate(dto, "")
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	entity := attendeeService.NewAttendee(ctx)
	err = mapDtoToAttendee(dto, entity)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
		return
	}
	id, err := attendeeService.RegisterNewAttendee(ctx, entity)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.Header().Set(headers.Location, fmt.Sprintf("%s/%d", r.RequestURI, id))
	w.WriteHeader(http.StatusCreated)
}

func getAttendeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	entity, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		attendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	dto := attendee.AttendeeDto{}
	mapAttendeeToDto(entity, &dto)
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
	validationErrs := validate(dto, fmt.Sprint(id))
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	entity, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		attendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	err = mapDtoToAttendee(dto, entity)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
		return
	}
	err = attendeeService.UpdateAttendee(ctx, entity)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.Header().Add(headers.Location, r.RequestURI)
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
	log.Printf("received attendee data with validation errors: %v", errs)
	errorHandler(ctx, w, r, "attendee.data.invalid", http.StatusBadRequest, errs)
}

func invalidIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id string) {
	log.Printf("received invalid attendee id '%s'", id)
	errorHandler(ctx, w, r, "attendee.id.invalid", http.StatusBadRequest, url.Values{})
}

func attendeeNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id uint) {
	log.Printf("attendee id %v not found", id)
	errorHandler(ctx, w, r, "attendee.id.notfound", http.StatusNotFound, url.Values{})
}

func attendeeParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("attendee body could not be parsed: %v", err)
	errorHandler(ctx, w, r, "attendee.parse.error", http.StatusBadRequest, url.Values{})
}

func attendeeWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("attendee could not be written: %v", err)
	errorHandler(ctx, w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
}

func errorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string, status int, details url.Values) {
	timestamp := time.Now().Format(time.RFC3339)
	response := attendee.ErrorDto{Message: msg, Timestamp: timestamp}
	// TODO include requestid
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(status)
	writeJson(ctx, w, response)
}

func writeJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		log.Printf("error while encoding json response: %v", err)
	}
}
