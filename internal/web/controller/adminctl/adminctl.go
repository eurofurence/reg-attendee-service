package adminctl

import (
	"context"
	"encoding/json"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"time"
)

var attendeeService attendeesrv.AttendeeService

var identityRegexp *regexp.Regexp

func Create(server chi.Router, attendeeSrv attendeesrv.AttendeeService) {
	attendeeService = attendeeSrv

	server.Get("/api/rest/v1/attendees/{id}/admin", filter.HasGroupOrApiToken(config.OidcAdminGroup(), filter.WithTimeout(3*time.Second, getAdminInfoHandler)))
	server.Put("/api/rest/v1/attendees/{id}/admin", filter.HasGroupOrApiToken(config.OidcAdminGroup(), filter.WithTimeout(3*time.Second, writeAdminInfoHandler)))
	server.Post("/api/rest/v1/attendees/find", filter.LoggedInOrApiToken(filter.WithTimeout(60*time.Second, findAttendeesHandler)))
	server.Get("/api/rest/v1/attendees/identity/{identity}", filter.HasGroupOrApiToken(config.OidcAdminGroup(), filter.WithTimeout(3*time.Second, regsByIdentityHandler)))

	identityRegexp = regexp.MustCompile("^[a-zA-Z0-9]+$")
}

// --- handlers ---

func getAdminInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	attendee, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	adminInfo, err := attendeeService.GetAdminInfo(ctx, attendee.ID)
	if err != nil {
		adminInfoReadErrorHandler(ctx, w, r, err)
		return
	}

	dto := admin.AdminInfoDto{}
	mapAdminInfoToDto(adminInfo, &dto)
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func writeAdminInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	attendee, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	dto, err := parseBodyToAdminInfoDto(ctx, w, r)
	if err != nil {
		return
	}

	suppressMinorUpdateEmail := r.URL.Query().Get("suppressMinorUpdateEmail") == "yes"

	// this will also create a blank adminInfo with id filled in
	adminInfo, err := attendeeService.GetAdminInfo(ctx, attendee.ID)
	if err != nil {
		adminInfoReadErrorHandler(ctx, w, r, err)
		return
	}

	validationErrs := validate(ctx, dto, adminInfo)
	if len(validationErrs) != 0 {
		adminInfoValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}

	mapDtoToAdminInfo(dto, adminInfo)

	err = attendeeService.UpdateAdminInfo(ctx, attendee, adminInfo, suppressMinorUpdateEmail)
	if err != nil {
		adminInfoWriteErrorHandler(ctx, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func findAttendeesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitedAccess := true
	if filter.IsGroupOrApiTokenCond(r, config.OidcAdminGroup()) {
		limitedAccess = false
	} else {
		allowed, err := attendeeService.CanUseFindAttendee(ctx)
		if err != nil || !allowed {
			culprit := ctxvalues.Subject(ctx)
			ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this operation - the attempt has been logged", fmt.Sprintf("unauthorized access attempt to find endpoint by %s", culprit))
			return
		}
	}

	criteria, err := parseBodyToAttendeeSearchCriteria(ctx, w, r)
	if err != nil {
		return
	}

	if limitedAccess {
		criteria.FillFields = limitToAllowedFields(criteria.FillFields)
		for i := range criteria.MatchAny {
			criteria.MatchAny[i].Status = limitToAttendingStatus(criteria.MatchAny[i].Status)
		}
	}

	results, err := attendeeService.FindAttendees(ctx, criteria)
	if err != nil {
		searchReadErrorHandler(ctx, w, r, err)
		return
	}

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, results)
}

func limitToAllowedFields(desired []string) []string {
	allowed := []string{"id", "nickname", "first_name", "last_name", "country",
		"spoken_languages", "registration_language", "birthday", "pronouns", "tshirt_size",
		"flags", "options", "packages", "status",
		"total_dues", "payment_balance", "current_dues", "identity_subject", "avatar",
	}

	result := make([]string, 0)
	for _, d := range desired {
		for _, a := range allowed {
			if d == a {
				result = append(result, d)
			}
		}
	}

	if len(result) == 0 {
		return allowed
	} else {
		return result
	}
}

func limitToAttendingStatus(desired []status.Status) []status.Status {
	if len(desired) == 0 {
		return []status.Status{status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn}
	} else {
		result := make([]status.Status, 0)
		for _, s := range desired {
			if s == status.Approved || s == status.PartiallyPaid || s == status.Paid || s == status.CheckedIn {
				result = append(result, s)
			}
		}
		return result
	}
}

func regsByIdentityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	identity := chi.URLParam(r, "identity")
	if !identityRegexp.MatchString(identity) {
		regsByIdentityInvalidErrorHandler(ctx, w, r, identity)
		return
	}

	atts, err := attendeeService.IsOwnedByIdentity(ctx, identity)
	if err != nil {
		regsByIdentityErrorHandler(ctx, w, r, identity, err)
		return
	}
	if len(atts) == 0 {
		regsByIdentityNotFoundErrorHandler(ctx, w, r, identity)
		return
	}

	dto := attendee.AttendeeIdList{
		Ids: make([]uint, len(atts)),
	}
	for i, _ := range atts {
		dto.Ids[i] = atts[i].ID
	}
	sort.Slice(dto.Ids, func(i, j int) bool { return dto.Ids[i] < dto.Ids[j] })

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
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

func parseBodyToAdminInfoDto(ctx context.Context, w http.ResponseWriter, r *http.Request) (*admin.AdminInfoDto, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	dto := &admin.AdminInfoDto{}
	err := decoder.Decode(dto)
	if err != nil {
		adminInfoParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

func parseBodyToAttendeeSearchCriteria(ctx context.Context, w http.ResponseWriter, r *http.Request) (*attendee.AttendeeSearchCriteria, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	dto := &attendee.AttendeeSearchCriteria{}
	err := decoder.Decode(dto)
	if err != nil {
		searchCriteriaParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

// --- error handlers ---

func adminInfoReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("adminInfo could not be read for existing attendee: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "admin.read.error", http.StatusInternalServerError, url.Values{})
}

func adminInfoWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("adminInfo could not be written for existing attendee: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "admin.write.error", http.StatusInternalServerError, url.Values{})
}

func adminInfoParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("adminInfo body could not be parsed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "admin.parse.error", http.StatusBadRequest, url.Values{})
}

func adminInfoValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received adminInfo data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "admin.data.invalid", http.StatusBadRequest, errs)
}

func searchCriteriaParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("attendee search criteria body could not be parsed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "search.parse.error", http.StatusBadRequest, url.Values{})
}

func searchReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("attendee search failed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "search.read.error", http.StatusInternalServerError, url.Values{})
}

func regsByIdentityInvalidErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, identity string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid identity %s", url.QueryEscape(identity))
	ctlutil.ErrorHandler(ctx, w, r, "attendee.owned.invalid", http.StatusBadRequest, url.Values{"identity": []string{"identity id can only consist of A-Z, a-z, 0-9"}})
}

func regsByIdentityErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, identity string, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("could not read registrations for identity %s: %s", identity, err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.owned.error", http.StatusInternalServerError, url.Values{})
}

func regsByIdentityNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, identity string) {
	aulogging.Logger.Ctx(ctx).Debug().Printf("found no registrations owned by identity %s", identity)
	ctlutil.ErrorHandler(ctx, w, r, "attendee.owned.notfound", http.StatusNotFound, url.Values{})
}
