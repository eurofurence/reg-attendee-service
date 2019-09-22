package ctxfilter

import (
	"context"
	"github.com/google/uuid"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxvalues"
	"net/http"
	"time"
)

type ContextFilter struct {
	timeout time.Duration
	wrappedFilter filter.Filter
}

func Create(timeout time.Duration, wrappedFilter filter.Filter) filter.Filter {
	return &ContextFilter{timeout: timeout, wrappedFilter: wrappedFilter}
}

const TraceIdHeader = "X-B3-TraceId"

func (f *ContextFilter) Handle(_ context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(context.Background(), f.timeout)
	defer cancel() // Cancel ctx as soon as Handle returns.

	ctx = ctxvalues.CreateContextWithValueMap(ctx)

	reqUuidStr := r.Header.Get(TraceIdHeader)
	if reqUuidStr == "" {
		reqUuid, err := uuid.NewRandom()
		if err == nil {
			reqUuidStr = reqUuid.String()
		} else {
			reqUuidStr ="uuid-generate-error"
		}
	}
	ctxvalues.SetRequestId(ctx, reqUuidStr)
	w.Header().Add(TraceIdHeader, reqUuidStr)

	f.wrappedFilter.Handle(ctx, w, r)
}
