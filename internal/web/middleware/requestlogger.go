package middleware

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		method := r.Method
		path := r.URL.EscapedPath()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()
		aulogging.Logger.Ctx(ctx).Debug().Printf("received request %s %s", method, path)

		defer func() {
			elapsed := time.Since(start)
			aulogging.Logger.Ctx(ctx).Info().Printf("request %s %s -> %d (%d ms)", method, path, ww.Status(), elapsed.Nanoseconds()/1000000)
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
