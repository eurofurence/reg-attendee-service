package packagectl

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/counts"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
)

var attendeeService attendeesrv.AttendeeService

func Create(server chi.Router, attendeeSrv attendeesrv.AttendeeService) {
	attendeeService = attendeeSrv

	server.Get("/api/rest/v1/packages/{package}/limit", filter.LoggedInOrApiToken(filter.WithTimeout(3*time.Second, getPackageLimit)))
	server.Post("/api/rest/v1/packages/{package}/limit", filter.HasGroupOrApiToken(config.OidcAdminGroup(), filter.WithTimeout(30*time.Second, recalcPackageLimit)))
}

func getPackageLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code, choice, err := packageFromVars(ctx, w, r)
	if err != nil {
		return
	}

	if choice.Limit == 0 {
		packageUnlimitedErrorHandler(ctx, w, r, code)
		return
	}

	count, err := attendeeService.GetLimitBookings(ctx, code)
	if err != nil {
		otherErrorHandler(ctx, w, r, code, err)
		return
	}

	dto := counts.PackageCount{
		Pending:   count.Pending,
		Attending: count.Attending,
		Limit:     choice.Limit,
	}
	w.Header().Add(headers.ContentType, media.ContentTypeApplicationJson)
	ctlutil.WriteJson(ctx, w, dto)
}

func recalcPackageLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code, choice, err := packageFromVars(ctx, w, r)
	if err != nil {
		return
	}

	if choice.Limit == 0 {
		packageUnlimitedErrorHandler(ctx, w, r, code)
		return
	}

	if err := attendeeService.RecalculateLimit(ctx, code); err != nil {
		otherErrorHandler(ctx, w, r, code, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func packageFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (string, config.ChoiceConfig, error) {
	code := chi.URLParam(r, "package")
	choice, ok := config.Configuration().Choices.Packages[code]
	if !ok {
		packageNotFoundErrorHandler(ctx, w, r, code)
		return code, config.ChoiceConfig{}, fmt.Errorf("invalid package %s requested", url.QueryEscape(code))
	}
	return code, choice, nil
}

func packageNotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, code string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("found no package %s in configuration", url.QueryEscape(code))
	ctlutil.ErrorHandler(ctx, w, r, "package.param.notfound", http.StatusNotFound, url.Values{})
}

func packageUnlimitedErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, code string) {
	aulogging.Logger.Ctx(ctx).Warn().Printf("package %s is unlimited", url.QueryEscape(code))
	ctlutil.ErrorHandler(ctx, w, r, "package.param.unlimited", http.StatusBadRequest, url.Values{"details": []string{"this package is unlimited, we do not track allocations for unlimited packages"}})
}

func otherErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, code string, err error) {
	aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("failed to check limits for package %s: %s", url.QueryEscape(code), err.Error())
	ctlutil.ErrorHandler(ctx, w, r, "package.read.error", http.StatusInternalServerError, url.Values{})
}
