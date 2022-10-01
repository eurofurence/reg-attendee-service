package filter

import (
	"context"
	"net/http"
	"time"
)

func WithTimeout(timeout time.Duration, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		handler(w, r.WithContext(ctx))
	}
}
