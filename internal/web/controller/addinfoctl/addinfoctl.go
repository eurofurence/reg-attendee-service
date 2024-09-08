package addinfoctl

import (
	"context"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/addinfo"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
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

	server.Get("/api/rest/v1/additional-info/{area}", filter.LoggedInOrApiToken(filter.WithTimeout(60*time.Second, getAllAdditionalInfoHandler)))

	areaRegexp = regexp.MustCompile("^[a-z]+$")
}

func getAdditionalInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r, false)
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
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r, true)
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
	ctx, id, area, err := ctxIdAreaAllowedAndExists_MustReturn(w, r, true)
	if err != nil {
		return
	}

	oldValue, err := attendeeService.GetAdditionalInfo(ctx, id, area)
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
		return
	}
	if oldValue == "" {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.notfound.error", http.StatusNotFound, url.Values{})
		return
	}

	err = attendeeService.WriteAdditionalInfo(ctx, id, area, "")
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.write.error", http.StatusInternalServerError, url.Values{})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAllAdditionalInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, area, err := ctxFullAreaReadAllowedAndExists_MustReturn(w, r)
	if err != nil {
		return
	}

	values, err := attendeeService.GetFullAdditionalInfoArea(ctx, area)
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
		return
	}

	result := addinfo.AdditionalInfoFullArea{
		Area:   area,
		Values: values,
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, &result)
}

func ctxIdAreaAllowedAndExists_MustReturn(w http.ResponseWriter, r *http.Request, wantWriteAccess bool) (context.Context, uint, string, error) {
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
		allowed, err = attendeeService.CanAccessOwnAdditionalInfoArea(ctx, id, wantWriteAccess, area)
		if err != nil {
			ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
			return ctx, id, area, err
		}
		if !allowed {
			culprit := ctxvalues.Subject(ctx)
			ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this additional info area - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for add info area %s by %s", area, culprit))
			return ctx, id, area, errors.New("forbidden")
		}
	}

	_, err = attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return ctx, id, area, err
	}

	return ctx, id, area, nil
}

func ctxFullAreaReadAllowedAndExists_MustReturn(w http.ResponseWriter, r *http.Request) (context.Context, string, error) {
	ctx := r.Context()

	area, err := areaFromVarsValidated_MustReturn(ctx, w, r)
	if err != nil {
		return ctx, area, err
	}

	allowed, err := attendeeService.CanAccessAdditionalInfoArea(ctx, area)
	if err != nil {
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.read.error", http.StatusInternalServerError, url.Values{})
		return ctx, area, err
	}
	if !allowed {
		culprit := ctxvalues.Subject(ctx)
		ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this additional info area - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for add info area %s by %s", area, culprit))
		return ctx, area, errors.New("forbidden")
	}

	return ctx, area, nil
}

func idAndAreaFromVarsValidated_MustReturn(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, string, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctlutil.InvalidAttendeeIdErrorHandler(ctx, w, r, idStr)
		return uint(id), "", err
	}
	area, err := areaFromVarsValidated_MustReturn(ctx, w, r)
	return uint(id), area, err
}

func areaFromVarsValidated_MustReturn(ctx context.Context, w http.ResponseWriter, r *http.Request) (string, error) {
	area := chi.URLParam(r, "area")
	if !areaRegexp.MatchString(area) {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid additional info area '%s'", area)
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.area.invalid", http.StatusBadRequest, url.Values{"area": []string{"must match [a-z]+"}})
		return area, errors.New("invalid additional info area")
	}
	if area == "overdue" {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid additional info area '%s'", area)
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.area.invalid", http.StatusBadRequest, url.Values{"area": []string{"the special value 'overdue' is used internally and is forbidden here"}})
		return area, errors.New("invalid additional info area")
	}
	if validation.NotInAllowedValues(config.AdditionalInfoFieldNames(), area) {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received additional info area '%s' not listed in configuration", area)
		ctlutil.ErrorHandler(ctx, w, r, "addinfo.area.unlisted", http.StatusBadRequest, url.Values{"area": []string{"areas must be enabled in configuration"}})
		return area, errors.New("unlisted additional info area")
	}

	return area, nil
}
