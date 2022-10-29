package middleware

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/go-http-utils/headers"
	"net/http"
)

func CorsHandling(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if config.IsCorsDisabled() {
			aulogging.Logger.Ctx(ctx).Info().Print("sending headers to disable CORS. This configuration is not intended for production use, only for local development!")
			w.Header().Set(headers.AccessControlAllowOrigin, config.CorsAllowOrigin())
			w.Header().Set(headers.AccessControlAllowMethods, "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set(headers.AccessControlAllowHeaders, "content-type")
			w.Header().Set(headers.AccessControlAllowCredentials, "true")
			w.Header().Set(headers.AccessControlExposeHeaders, "Location, "+TraceIdHeader)
		}

		if r.Method == http.MethodOptions {
			aulogging.Logger.Ctx(ctx).Debug().Print("received OPTIONS request. Responding with OK.")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
