package attendeectl

import (
	"encoding/json"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"rexis/rexis-go-attendee/api/v1/attendee"
	"rexis/rexis-go-attendee/internal/service/attendeesrv"
	"rexis/rexis-go-attendee/web/util/media"
	"strconv"
	"time"
)

var attendeeService attendeesrv.AttendeeService

func init() {
	attendeeService = &attendeesrv.AttendeeServiceImplData{}
}

func RestDispatcher(router *mux.Router) {
	router.HandleFunc("/v1/attendees", newAttendeeHandler).Methods(http.MethodPut)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", getAttendeeHandler).Methods(http.MethodGet)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", updateAttendeeHandler).Methods(http.MethodPost)
}

func newAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	dto, err := parseBodyToAttendeeDto(w, r)
	if err != nil {
		return
	}
	validationErrs := validate(dto, "")
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(w, r, validationErrs)
		return
	}
	entity := attendeeService.NewAttendee()
	err = mapDtoToAttendee(dto, entity)
	if err != nil {
		attendeeParseErrorHandler(w, r, err)
		return
	}
	id, err := attendeeService.RegisterNewAttendee(entity)
	if err != nil {
		attendeeWriteErrorHandler(w, r, err)
		return
	}
	w.Header().Set(headers.Location, fmt.Sprintf("%s/%d", r.RequestURI, id))
	w.WriteHeader(http.StatusCreated)
}

func getAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(w, r)
	if err != nil {
		return
	}
	entity, err := attendeeService.GetAttendee(id)
	if err != nil {
		attendeeNotFoundErrorHandler(w, r, id)
		return
	}
	dto := attendee.AttendeeDto{}
	mapAttendeeToDto(entity, &dto)
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	writeJson(w, dto)
}

func updateAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromVars(w, r)
	if err != nil {
		return
	}
	dto, err := parseBodyToAttendeeDto(w, r)
	if err != nil {
		return
	}
	validationErrs := validate(dto, fmt.Sprint(id))
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(w, r, validationErrs)
		return
	}
	entity, err := attendeeService.GetAttendee(id)
	if err != nil {
		attendeeNotFoundErrorHandler(w, r, id)
		return
	}
	err = mapDtoToAttendee(dto, entity)
	if err != nil {
		attendeeParseErrorHandler(w, r, err)
		return
	}
	err = attendeeService.UpdateAttendee(entity)
	if err != nil {
		attendeeWriteErrorHandler(w, r, err)
		return
	}
	w.Header().Add(headers.Location, r.RequestURI)
}

func idFromVars(w http.ResponseWriter, r *http.Request) (uint, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		invalidIdErrorHandler(w, r, vars["id"])
	}
	return uint(id), err
}

func parseBodyToAttendeeDto(w http.ResponseWriter, r *http.Request) (*attendee.AttendeeDto, error) {
	decoder := json.NewDecoder(r.Body)
	dto := &attendee.AttendeeDto{}
	err := decoder.Decode(dto)
	if err != nil {
		attendeeParseErrorHandler(w, r, err)
	}
	return dto, err
}

func attendeeValidationErrorHandler(w http.ResponseWriter, r *http.Request, errs url.Values) {
	log.Printf("received attendee data with validation errors: %v", errs)
	errorHandler(w, r, "attendee.data.invalid", http.StatusBadRequest, errs)
}

func invalidIdErrorHandler(w http.ResponseWriter, r *http.Request, id string) {
	log.Printf("received invalid attendee id '%s'", id)
	errorHandler(w, r, "attendee.id.invalid", http.StatusBadRequest, url.Values{})
}

func attendeeNotFoundErrorHandler(w http.ResponseWriter, r *http.Request, id uint) {
	log.Printf("attendee id %v not found", id)
	errorHandler(w, r, "attendee.id.notfound", http.StatusNotFound, url.Values{})
}

func attendeeParseErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("attendee body could not be parsed: %v", err)
	errorHandler(w, r, "attendee.parse.error", http.StatusBadRequest, url.Values{})
}

func attendeeWriteErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("attendee could not be written: %v", err)
	errorHandler(w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
}

func errorHandler(w http.ResponseWriter, r *http.Request, msg string, status int, details url.Values) {
	timestamp := time.Now().Format(time.RFC3339)
	response := attendee.ErrorDto{Message: msg, Timestamp: timestamp}
	// TODO include requestid
	w.Header().Set(headers.ContentType, media.ContentTypeApplicationJson)
	w.WriteHeader(status)
	writeJson(w, response)
}

func writeJson(w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		log.Printf("error while encoding json response: %v", err)
	}
}
