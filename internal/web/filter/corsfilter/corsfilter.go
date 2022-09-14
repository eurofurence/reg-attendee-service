package corsfilter

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxfilter"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"github.com/go-http-utils/headers"
	"net/http"
)

type CorsFilter struct {
	wrappedFilter filter.Filter
}

func Create(wrappedFilter filter.Filter) filter.Filter {
	return &CorsFilter{wrappedFilter: wrappedFilter}
}

func (f *CorsFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if config.IsCorsDisabled() {
		logging.Ctx(ctx).Warn("sending headers to disable CORS. This configuration is not intended for production use, only for local development!")
		w.Header().Set(headers.AccessControlAllowOrigin, "*")
		w.Header().Set(headers.AccessControlAllowMethods, "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set(headers.AccessControlAllowHeaders, "content-type")
		w.Header().Set(headers.AccessControlExposeHeaders, "Location, "+ctxfilter.TraceIdHeader)
	}

	if r.Method == http.MethodOptions {
		logging.Ctx(ctx).Info("received OPTIONS request. Responding with OK.")

		status := http.StatusOK
		w.WriteHeader(status)
		ctxvalues.SetHttpStatus(ctx, status)

		return
	}

	f.wrappedFilter.Handle(ctx, w, r)
}
