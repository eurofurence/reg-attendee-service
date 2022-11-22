package banctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
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
	server.Get("/api/rest/v1/bans", filter.HasRoleOrApiToken(config.OidcAdminRole(), filter.WithTimeout(3*time.Second, allBansHandler)))
	server.Post("/api/rest/v1/bans", filter.HasRoleOrApiToken(config.OidcAdminRole(), filter.WithTimeout(3*time.Second, newBanHandler)))
	server.Get("/api/rest/v1/bans/{id}", filter.HasRoleOrApiToken(config.OidcAdminRole(), filter.WithTimeout(3*time.Second, getBanHandler)))
	server.Put("/api/rest/v1/bans/{id}", filter.HasRoleOrApiToken(config.OidcAdminRole(), filter.WithTimeout(3*time.Second, updateBanHandler)))
	server.Delete("/api/rest/v1/bans/{id}", filter.HasRoleOrApiToken(config.OidcAdminRole(), filter.WithTimeout(3*time.Second, deleteBanHandler)))
}

func allBansHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bansInDb, err := attendeeService.GetAllBans(ctx)
	if err != nil {
		banReadErrorHandler(ctx, w, r, err)
		return
	}

	response := bans.BanRuleList{
		Bans: make([]bans.BanRule, len(bansInDb)),
	}
	for i, b := range bansInDb {
		banDto := bans.BanRule{}
		mapBanToDto(b, &banDto)
		response.Bans[i] = banDto
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, response)
}

func newBanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dto, err := parseBodyToBanDto(ctx, w, r)
	if err != nil {
		return
	}
	validationErrs := validate(ctx, dto, 0)
	if len(validationErrs) != 0 {
		banValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}
	newBan := attendeeService.NewBan(ctx)
	mapDtoToBan(dto, newBan)
	id, err := attendeeService.CreateBan(ctx, newBan)
	if err != nil {
		banWriteErrorHandler(ctx, w, r, err)
		return
	}
	location := fmt.Sprintf("%s/%d", r.RequestURI, id)
	aulogging.Logger.Ctx(ctx).Info().Printf("sending Location %s", location)
	w.Header().Set(headers.Location, location)
	w.WriteHeader(http.StatusCreated)
}

func getBanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}

	existingBan, err := attendeeService.GetBan(ctx, id)
	if err != nil {
		banNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	response := bans.BanRule{}
	mapBanToDto(existingBan, &response)

	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, response)
}

func updateBanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}
	dto, err := parseBodyToBanDto(ctx, w, r)
	if err != nil {
		return
	}
	validationErrs := validate(ctx, dto, id)
	if len(validationErrs) != 0 {
		banValidationErrorHandler(ctx, w, r, validationErrs)
		return
	}

	existingBan, err := attendeeService.GetBan(ctx, id)
	if err != nil {
		banNotFoundErrorHandler(ctx, w, r, id)
		return
	}
	mapDtoToBan(dto, existingBan)

	err = attendeeService.UpdateBan(ctx, existingBan)
	if err != nil {
		banWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteBanHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := idFromVars(ctx, w, r)
	if err != nil {
		return
	}

	existingBan, err := attendeeService.GetBan(ctx, id)
	if err != nil {
		banNotFoundErrorHandler(ctx, w, r, id)
		return
	}

	err = attendeeService.DeleteBan(ctx, existingBan)
	if err != nil {
		banWriteErrorHandler(ctx, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func idFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		invalidBanIdErrorHandler(ctx, w, r, idStr)
	}
	return uint(id), err
}

func parseBodyToBanDto(ctx context.Context, w http.ResponseWriter, r *http.Request) (*bans.BanRule, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	dto := &bans.BanRule{}
	err := decoder.Decode(dto)
	if err != nil {
		banParseErrorHandler(ctx, w, r, err)
	}
	return dto, err
}

func invalidBanIdErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received invalid ban id '%s'", url.QueryEscape(id))
	ctlutil.ErrorHandler(ctx, w, r, "ban.id.invalid", http.StatusBadRequest, url.Values{})
}

func banValidationErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, errs url.Values) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("received ban rule data with validation errors: %v", errs)
	ctlutil.ErrorHandler(ctx, w, r, "ban.data.invalid", http.StatusBadRequest, errs)
}

func banParseErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("ban rule body could not be parsed: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "ban.parse.error", http.StatusBadRequest, url.Values{})
}

func banNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, id uint) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("ban id %d not found", id)
	ctlutil.ErrorHandler(ctx, w, r, "ban.id.notfound", http.StatusNotFound, url.Values{})
}

func banWriteErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("ban rule could not be written: %s", err.Error())
	if errors.Is(err, attendeesrv.DuplicateBanError) {
		ctlutil.ErrorHandler(ctx, w, r, "ban.data.duplicate", http.StatusConflict, url.Values{"ban": {"there is already another ban rule with the same patterns"}})
	} else {
		ctlutil.ErrorHandler(ctx, w, r, "ban.write.error", http.StatusInternalServerError, url.Values{})
	}
}

func banReadErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("ban rule(s) could not be read: %s", err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "ban.read.error", http.StatusInternalServerError, url.Values{})
}
