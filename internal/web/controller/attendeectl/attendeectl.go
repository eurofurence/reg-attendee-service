package attendeectl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
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
	"sort"
	"strconv"
	"time"
)

var attendeeService attendeesrv.AttendeeService

func Create(server chi.Router, attendeeSrv attendeesrv.AttendeeService) {
	attendeeService = attendeeSrv

	if config.RequireLoginForReg() {
		server.Post("/api/rest/v1/attendees", filter.LoggedIn(filter.WithTimeout(3*time.Second, newAttendeeHandler)))
	} else {
		server.Post("/api/rest/v1/attendees", filter.WithTimeout(3*time.Second, newAttendeeHandler))
	}
	server.Get("/api/rest/v1/attendees", filter.LoggedIn(filter.WithTimeout(3*time.Second, myRegsHandler)))
	server.Get("/api/rest/v1/attendees/max-id", filter.WithTimeout(3*time.Second, getAttendeeMaxIdHandler))
	server.Get("/api/rest/v1/attendees/{id}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getAttendeeHandler)))
	server.Put("/api/rest/v1/attendees/{id}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, updateAttendeeHandler)))
	server.Get("/api/rest/v1/attendees/{id}/due-date", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getDueDateHandler)))
	server.Put("/api/rest/v1/attendees/{id}/due-date", filter.HasGroupOrApiToken(config.OidcAdminGroup(), filter.WithTimeout(3*time.Second, overrideDueDateHandler)))

	server.Get("/api/rest/v1/attendees/{id}/flags/{flag}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getFlagHandler)))
	server.Get("/api/rest/v1/attendees/{id}/options/{option}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getOptionHandler)))
	server.Get("/api/rest/v1/attendees/{id}/packages/{package}", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getPackageHandler)))
}

func newAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dto, err := parseBodyToAttendeeDto(ctx, w, r)
	if err != nil {
		return
	}
	validationErrs := validate(ctx, dto, &entity.Attendee{Flags: config.DefaultFlags(), Packages: config.DefaultPackages(), Options: config.DefaultOptions()}, "irrelevant")
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

	if err := filter.IsSubjectOrGroupOrApiToken(w, r, existingAttendee.Identity, config.OidcAdminGroup()); err != nil {
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

	suppressMinorUpdateEmail := r.URL.Query().Get("suppressMinorUpdateEmail") == "yes"

	if err := filter.IsSubjectOrGroupOrApiToken(w, r, attd.Identity, config.OidcAdminGroup()); err != nil {
		return
	}

	latestStatus, err := obtainAttendeeLatestStatusMustReturnOnError(ctx, w, r, attd)
	if err != nil {
		return
	}

	validationErrs := validate(ctx, dto, attd, latestStatus)
	if len(validationErrs) != 0 {
		attendeeValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	mapDtoToAttendee(dto, attd)
	err = attendeeService.UpdateAttendee(ctx, attd, suppressMinorUpdateEmail)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.Header().Add(headers.Location, r.URL.Path)
	w.WriteHeader(http.StatusOK)
}

func getDueDateHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := filter.IsSubjectOrGroupOrApiToken(w, r, existingAttendee.Identity, config.OidcAdminGroup()); err != nil {
		return
	}

	dto := attendee.DueDate{
		DueDate: existingAttendee.CacheDueDate,
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func overrideDueDateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	dto, err := parseBodyToDueDate(ctx, w, r)
	if err != nil {
		return
	}
	attd, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}

	validationErrs := validateDueDateChange(ctx, dto, attd)
	if len(validationErrs) != 0 {
		dueDateValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}

	attd.CacheDueDate = dto.DueDate

	// note: even if suppressMinorUpdateEmails were false, still no emails would trigger as
	//       there are no due date changes in the transactions, and the attendee is written before we
	//       check for such changes.
	err = attendeeService.UpdateAttendee(ctx, attd, true)
	if err != nil {
		attendeeWriteErrorHandler(ctx, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func myRegsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	atts, err := attendeeService.IsOwnerFor(ctx)
	if err != nil {
		myRegsErrorHandler(ctx, w, r, err)
		return
	}
	if len(atts) == 0 {
		myRegsNotFoundErrorHandler(ctx, w, r)
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

func getFlagHandler(w http.ResponseWriter, r *http.Request) {
	getChoiceHandler(w, r,
		"flag",
		config.Configuration().Choices.Flags,
		func(ctx context.Context, w http.ResponseWriter, r *http.Request, attendee *entity.Attendee, code string, choice config.ChoiceConfig) (string, error) {
			if choice.AdminOnly {
				adminInfo, err := attendeeService.GetAdminInfo(ctx, attendee.ID)
				if err != nil {
					choiceErrorHandler(ctx, w, r, "flag", code, err)
					return "", err
				}
				return adminInfo.Flags, nil
			} else {
				return attendee.Flags, nil
			}
		},
	)
}

func getOptionHandler(w http.ResponseWriter, r *http.Request) {
	getChoiceHandler(w, r,
		"option",
		config.Configuration().Choices.Options,
		func(_ context.Context, _ http.ResponseWriter, _ *http.Request, attendee *entity.Attendee, _ string, _ config.ChoiceConfig) (string, error) {
			return attendee.Options, nil
		},
	)
}

func getPackageHandler(w http.ResponseWriter, r *http.Request) {
	getChoiceHandler(w, r,
		"package",
		config.Configuration().Choices.Packages,
		func(_ context.Context, _ http.ResponseWriter, _ *http.Request, attendee *entity.Attendee, _ string, _ config.ChoiceConfig) (string, error) {
			return packagesFromEntity(attendee.Packages), nil
		},
	)
}

func getChoiceHandler(w http.ResponseWriter, r *http.Request, choiceType string,
	choiceConfigMap map[string]config.ChoiceConfig,
	commaSeparatedValueGetter func(ctx context.Context, w http.ResponseWriter, r *http.Request, attendee *entity.Attendee, code string, choice config.ChoiceConfig) (string, error),
) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}

	code, choice, err := choiceFromVars(ctx, w, r, choiceType, choiceConfigMap)
	if err != nil {
		return
	}

	requestedAttendee, err := attendeeService.GetAttendee(ctx, id)
	if err != nil {
		ctlutil.AttendeeNotFoundErrorHandler(ctx, w, r, id)
		return
	}

	err = choiceVisibilityCheckMustReturnOnError(ctx, w, r, requestedAttendee, choiceType, code, choice)
	if err != nil {
		return
	}

	value, err := commaSeparatedValueGetter(ctx, w, r, requestedAttendee, code, choice)
	if err != nil {
		return
	}

	dto := attendee.ChoiceState{
		Present: commaSeparatedContains(value, code),
	}

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func choiceVisibilityCheckMustReturnOnError(ctx context.Context, w http.ResponseWriter, r *http.Request, requestedAttendee *entity.Attendee, choiceType string, code string, choice config.ChoiceConfig) (err error) {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
		// admin rights, all flags visible
		return nil
	} else if ctxvalues.Subject(ctx) == requestedAttendee.Identity {
		// self
		if choiceType == "flag" {
			if choice.AdminOnly {
				if !sliceContains(choice.VisibleFor, "self") {
					choiceNotAccessibleHandler(ctx, w, r, choiceType, code)
					return errors.New("not accessible")
				}
			}
		}
		return nil
	} else {
		// by area
		allowed := false
		if len(choice.VisibleFor) > 0 {
			allowed, err = attendeeService.CanAccessAdditionalInfoArea(ctx, choice.VisibleFor...)
			if err != nil {
				choiceErrorHandler(ctx, w, r, choiceType, code, err)
				return errors.New("internal error")
			}
		}
		if !allowed {
			choiceNotAccessibleHandler(ctx, w, r, choiceType, code)
			return errors.New("not accessible")
		}
		return nil
	}
}

func idFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctlutil.InvalidAttendeeIdErrorHandler(ctx, w, r, idStr)
	}
	return uint(id), err
}

func choiceFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string, choiceConfig map[string]config.ChoiceConfig) (string, config.ChoiceConfig, error) {
	code := chi.URLParam(r, paramName)
	choice, ok := choiceConfig[code]
	if !ok {
		choiceNotFoundErrorHandler(ctx, w, r, paramName, code)
		return code, config.ChoiceConfig{}, fmt.Errorf("invalid %s %s requested", paramName, url.QueryEscape(code))
	}
	return code, choice, nil
}

func parseBodyToAttendeeDto(ctx context.Context, w http.ResponseWriter, r *http.Request) (*attendee.AttendeeDto, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	dto := &attendee.AttendeeDto{}
	err := decoder.Decode(dto)
	if err != nil {
		attendeeParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

func parseBodyToDueDate(ctx context.Context, w http.ResponseWriter, r *http.Request) (*attendee.DueDate, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	dto := &attendee.DueDate{}
	err := decoder.Decode(dto)
	if err != nil {
		dueDateParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

func obtainAttendeeLatestStatusMustReturnOnError(ctx context.Context, w http.ResponseWriter, r *http.Request, att *entity.Attendee) (status.Status, error) {
	history, err := attendeeService.GetFullStatusHistory(ctx, att)
	if err != nil {
		attendeeReadErrorHandler(ctx, w, r, err)
		return "unknown", err
	} else if len(history) == 0 {
		err := errors.New("got empty status change history")
		attendeeReadErrorHandler(ctx, w, r, err)
		return "unknown", err
	}

	latest := history[len(history)-1]
	return latest.Status, nil
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
	} else if err.Error() == "duplicate - must use a separate email address and identity account for each person" {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.user.duplicate", http.StatusConflict, url.Values{"user": {"you already have a registration - please use a separate email address and matching account per person"}})
	} else if err.Error() == "your changes would lead to duplicate attendee data - same nickname, zip, email" {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.user.duplicate", http.StatusConflict, url.Values{"attendee": {"your changes would lead to duplicate attendee data - same nickname, zip, email"}})
	} else {
		ctlutil.ErrorHandler(ctx, w, r, "attendee.write.error", http.StatusInternalServerError, url.Values{})
	}
	// TODO: distinguish attendee.payment.error -> bad gateway
}

func attendeeReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("attendee could not be read: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.read.error", http.StatusInternalServerError, url.Values{})
}

func attendeeMaxIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("could not determine max id: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.max_id.error", http.StatusInternalServerError, url.Values{})
}

func dueDateParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("due date body could not be parsed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "duedate.parse.error", http.StatusBadRequest, url.Values{})
}

func dueDateValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received due date data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "duedate.data.invalid", http.StatusBadRequest, errs)
}

func myRegsErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("could not read registrations for logged in subject: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.owned.error", http.StatusInternalServerError, url.Values{})
}

func myRegsNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	aulogging.Logger.Ctx(ctx).Debug().Printf("found no registrations owned by logged in subject")
	ctlutil.ErrorHandler(ctx, w, r, "attendee.owned.notfound", http.StatusNotFound, url.Values{})
}

func choiceNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string, code string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("found no %s %s in configuration", paramName, url.QueryEscape(code))
	ctlutil.ErrorHandler(ctx, w, r, "attendee.param.invalid", http.StatusBadRequest, url.Values{})
}

func choiceNotAccessibleHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string, code string) {
	culprit := ctxvalues.Subject(ctx)
	ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this operation - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for %s %s by %s", paramName, url.QueryEscape(code), culprit))
}

func choiceErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string, code string, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to check visibility for %s %s: %s", paramName, url.QueryEscape(code), err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "attendee.param.error", http.StatusInternalServerError, url.Values{})
}
