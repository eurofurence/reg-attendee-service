package addinfoctl

import (
	"context"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var attendeeService attendeesrv.AttendeeService

var areaRegexp *regexp.Regexp

func Create(server chi.Router, attendeeSrv attendeesrv.AttendeeService) {
	attendeeService = attendeeSrv

	server.Get("/api/rest/v1/attendees/{id}/additional-info/{area}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getAdditionalInfoHandler)))
	server.Post("/api/rest/v1/attendees/{id}/additional-info/{area}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, writeAdditionalInfoHandler)))
	server.Delete("/api/rest/v1/attendees/{id}/additional-info/{area}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, deleteAdditionalInfoHandler)))

	areaRegexp = regexp.MustCompile("^[a-z]+$")
}

func getAdditionalInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r)
	if err != nil {
		return
	}

	value, err := attendeeService.GetAdditionalInfo(ctx, id, area)
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
		return
	}
	if value == "" {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.notfound.error", http.StatusNotFound, url.Values{})
		return
	}

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	_, err = io.WriteString(w, value)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().Printf("failed to send full additional info value as body for id %d area %s - maybe interrupted or timeout, and cannot notify recipient: %s", id, area, err.Error())
	}
}

func writeAdditionalInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r)
	if err != nil {
		return
	}

	value, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.receive.error", http.StatusInternalServerError, url.Values{})
		return
	}
	if len(value) == 0 {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.use.delete", http.StatusBadRequest, url.Values{})
		return
	}

	err = attendeeService.WriteAdditionalInfo(ctx, id, area, string(value))
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.write.error", http.StatusInternalServerError, url.Values{})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteAdditionalInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r)
	if err != nil {
		return
	}

	err = attendeeService.WriteAdditionalInfo(ctx, id, area, "")
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.write.error", http.StatusInternalServerError, url.Values{})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ctxIdAreaAllowedAndExists_MustReturn(w http.ResponseWriter, r *http.Request) (context.Context, uint, string, error) {
	ctx := r.Context()

	id, area, err := idAndAreaFromVarsValidated_MustReturn(ctx, w, r)
	if err != nil {
		return ctx, id, area, err
	}

	allowed, err := attendeeService.CanAccessAdditionalInfoArea(ctx, area)
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
		return ctx, id, area, err
	}
	if !allowed {
		culprit := ctxvalues.Subject(ctx)
		ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this additional info area - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for add info area %s by %s", area, culprit))
		return ctx, id, area, errors.New("forbidden")
	}

	_, err = attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return ctx, id, area, err
	}

	return ctx, id, area, nil
}

func idAndAreaFromVarsValidated_MustReturn(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, string, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctlutil.InvalidAttendeeIdErrorHandler(ctx, w, r, idStr)
		return uint(id), "", err
	}
	area := chi.URLParam(r, "area")
	if !areaRegexp.MatchString(area) {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid additional info area '%s'", area)
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.area.invalid", http.StatusBadRequest, url.Values{"area": []string{"must match [a-z]+"}})
		return uint(id), area, errors.New("invalid additional info area")
	}
	if area == "overdue" {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid additional info area '%s'", area)
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.area.invalid", http.StatusBadRequest, url.Values{"area": []string{"the special value 'overdue' is used internally and is forbidden here"}})
		return uint(id), area, errors.New("invalid additional info area")
	}
	return uint(id), area, nil
}
