package filterhelper

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxfilter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/handlefilter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/logfilter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/securityfilter"
	"net/http"
	"time"
)

func buildHandlerFunc(f filter.Filter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { f.Handle(context.TODO(), w, r) }
}

func parseTimeout(timeout string) time.Duration {
	parsedDuration, err := time.ParseDuration(timeout)
	if err != nil {
		logging.NoCtx().Fatalf("invalid timeout duration '%s', try something like '800ms' or '4s': %v", timeout, err)
	}
	return parsedDuration
}

func BuildUnauthenticatedNologgingHandler(timeout string, handler filter.ContextAwareHandler) func(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := parseTimeout(timeout)
	return buildHandlerFunc(
		ctxfilter.Create(timeoutDuration,
			handlefilter.Create(handler)))
}

func BuildUnauthenticatedHandler(timeout string, handler filter.ContextAwareHandler) func(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := parseTimeout(timeout)
	return buildHandlerFunc(
		ctxfilter.Create(timeoutDuration,
			logfilter.Create(
				handlefilter.Create(handler))))
}

func BuildHandler(timeout string, handler filter.ContextAwareHandler) func(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := parseTimeout(timeout)
	return buildHandlerFunc(
		ctxfilter.Create(timeoutDuration,
			logfilter.Create(
				securityfilter.Create(
					handlefilter.Create(handler)))))
}
