package filter

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/system"
	"net/http"
	"time"
)

func parseTimeout(timeout string) time.Duration {
	parsedDuration, err := time.ParseDuration(timeout)
	if err != nil {
		aulogging.Logger.NoCtx().Error().WithErr(err).Printf("invalid timeout duration '%s', try something like '800ms' or '4s': %s", timeout, err.Error())
		system.Exit(1)
	}
	return parsedDuration
}

func WithTimeout(timeout string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), parseTimeout(timeout))
		defer cancel()

		handler(w, r.WithContext(ctx))
	}
}
