package middleware

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		method := r.Method
		path := r.URL.EscapedPath()

		start := time.Now()
		aulogging.Logger.Ctx(ctx).Debug().Printf("received request %s %s", method, path)

		next.ServeHTTP(w, r)

		elapsed := time.Since(start)
		aulogging.Logger.Ctx(ctx).Info().Printf("request %s %s -> %s (%d ms)", method, path, ctxvalues.HttpStatus(ctx), elapsed.Nanoseconds()/1000000)
	}

	return http.HandlerFunc(fn)
}
