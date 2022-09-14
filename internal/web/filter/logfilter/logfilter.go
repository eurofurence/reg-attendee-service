package logfilter

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
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
	logging.Ctx(ctx).Infof("received %s %s", r.Method, r.URL.EscapedPath())

	f.wrappedFilter.Handle(ctx, w, r)

	elapsed := time.Since(start)
	logging.Ctx(ctx).Infof("finished %s %s in %d ms -> %s", r.Method, r.URL.EscapedPath(),
		elapsed.Nanoseconds()/1000000, ctxvalues.HttpStatus(ctx))

	/*
		this will get called upon cancel(), but request processing does not react to the event and actually get aborted, so it's useless as it is

		would probably need to run processing the request in a goroutine and have main wait on this channel to see the timeout event ???

		go func() {
			for {
				_, ok := <- ctx.Done()
				if !ok {
					println("GOT done")
					return
				}
				println("GOT notdone")
			}
		}()
	*/
}
