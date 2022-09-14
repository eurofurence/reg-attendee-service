package ctxfilter

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type ContextFilter struct {
	timeout       time.Duration
	wrappedFilter filter.Filter
}

func Create(timeout time.Duration, wrappedFilter filter.Filter) filter.Filter {
	return &ContextFilter{timeout: timeout, wrappedFilter: wrappedFilter}
}

const TraceIdHeader = "X-B3-TraceId"

func (f *ContextFilter) Handle(ctxOrig context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctxOrig, f.timeout)
	defer cancel() // Cancel ctx as soon as Handle returns.

	reqUuidStr := r.Header.Get(TraceIdHeader)
	if reqUuidStr == "" {
		reqUuid, err := uuid.NewRandom()
		if err == nil {
			reqUuidStr = reqUuid.String()[:8]
		} else {
			// this should not normally ever happen, but continue with this fixed requestId
			reqUuidStr = "ffffffff"
		}
	}
	ctx = logging.CreateContextWithLoggerForRequestId(ctx, reqUuidStr)

	ctx = ctxvalues.CreateContextWithValueMap(ctx)
	ctxvalues.SetRequestId(ctx, reqUuidStr)
	w.Header().Add(TraceIdHeader, reqUuidStr)

	f.wrappedFilter.Handle(ctx, w, r)
}
