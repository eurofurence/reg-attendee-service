package ctxfilter

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
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

func (f *ContextFilter) Handle(ctxOrig context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctxOrig, f.timeout)
	defer cancel() // Cancel ctx as soon as Handle returns.

	f.wrappedFilter.Handle(ctx, w, r)
}
