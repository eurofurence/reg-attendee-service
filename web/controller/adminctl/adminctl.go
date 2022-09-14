package adminctl

import (
	"context"
	"encoding/json"
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
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
	server.Get("/api/rest/v1/attendees/{id}/admin", filterhelper.BuildHandler("3s", getAdminInfoHandler, config.TokenForAdmin))
	server.Put("/api/rest/v1/attendees/{id}/admin", filterhelper.BuildHandler("3s", writeAdminInfoHandler, config.TokenForAdmin))
}

// --- handlers ---

func getAdminInfoHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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

func writeAdminInfoHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	attendee, err := attendeeByIdMustReturnOnError(ctx, w, r)
	if err != nil {
		return
	}

	dto, err := parseBodyToAdminInfoDto(ctx, w, r)
	if err != nil {
		return
	}

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

	err = attendeeService.UpdateAdminInfo(ctx, adminInfo)
	if err != nil {
		adminInfoWriteErrorHandler(ctx, w, r, err)
		return
	}

	ctlutil.WriteHeader(ctx, w, http.StatusNoContent)
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
	dto := &admin.AdminInfoDto{}
	err := decoder.Decode(dto)
	if err != nil {
		adminInfoParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

// --- error handlers ---

func adminInfoReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("adminInfo could not be read for existing attendee: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "admin.read.error", http.StatusInternalServerError, url.Values{})
}

func adminInfoWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("adminInfo could not be written for existing attendee: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "admin.write.error", http.StatusInternalServerError, url.Values{})
}

func adminInfoParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	logging.Ctx(ctx).Warnf("adminInfo body could not be parsed: %v", err)
	ctlutil.ErrorHandler(ctx, w, r, "admin.parse.error", http.StatusBadRequest, url.Values{})
}

func adminInfoValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	logging.Ctx(ctx).Warnf("received adminInfo data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "admin.data.invalid", http.StatusBadRequest, errs)
}