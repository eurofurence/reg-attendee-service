package logfilter

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"net/http"
	"time"
)

type LogFilter struct {
	wrappedFilter filter.Filter
}

func Create(wrappedFilter filter.Filter) filter.Filter {
	return &LogFilter{wrappedFilter: wrappedFilter}
}

func (f *LogFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	aulogging.Logger.Ctx(ctx).Debug().Printf("received request %s %s", r.Method, r.URL.EscapedPath())

	f.wrappedFilter.Handle(ctx, w, r)

	elapsed := time.Since(start)
	aulogging.Logger.Ctx(ctx).Info().Printf("request %s %s -> %s (%d ms)", r.Method, r.URL.EscapedPath(), ctxvalues.HttpStatus(ctx), elapsed.Nanoseconds()/1000000)
}
