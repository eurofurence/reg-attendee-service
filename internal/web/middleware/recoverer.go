package middleware

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"net/http"
	"net/url"
	"runtime/debug"
)

func PanicRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr != nil && rvr != http.ErrAbortHandler {
				ctx := r.Context()
				stack := string(debug.Stack())
				aulogging.Logger.Ctx(ctx).Error().Print("recovered from PANIC: " + stack)
				ctlutil.ErrorHandler(ctx, w, r, "internal.error", http.StatusInternalServerError, url.Values{})
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
