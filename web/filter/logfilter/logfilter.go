package logfilter

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter"
	"net/http"
)

type LogFilter struct {
	wrappedFilter filter.Filter
}

func Create(wrappedFilter filter.Filter) filter.Filter {
	return &LogFilter{wrappedFilter: wrappedFilter}
}

func (f *LogFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// TODO implement request logging here
	f.wrappedFilter.Handle(ctx, w, r)
	// TODO implement response and timing logging here
	// TODO can we subscribe a special logging function for when the context timeout fires?
}
