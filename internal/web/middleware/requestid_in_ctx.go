package middleware

import (
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/google/uuid"
	"net/http"
)

const TraceIdHeader = "X-Request-Id"

func AddRequestIdToContextAndResponse(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reqIdStr := r.Header.Get(TraceIdHeader)
		if reqIdStr == "" || len(reqIdStr) > 8 {
			reqUuid, err := uuid.NewRandom()
			if err == nil {
				reqIdStr = reqUuid.String()[:8]
			} else {
				// this should not normally ever happen, but continue with this fixed requestId rather than none
				reqIdStr = "ffffffff"
			}
		}

		ctx = ctxvalues.CreateContextWithValueMap(ctx)
		ctxvalues.SetRequestId(ctx, reqIdStr)
		w.Header().Add(TraceIdHeader, reqIdStr)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
