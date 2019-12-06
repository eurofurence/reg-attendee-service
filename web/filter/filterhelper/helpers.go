package filterhelper

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/web/filter"
	"github.com/eurofurence/reg-attendee-service/web/filter/corsfilter"
	"github.com/eurofurence/reg-attendee-service/web/filter/ctxfilter"
	"github.com/eurofurence/reg-attendee-service/web/filter/handlefilter"
	"github.com/eurofurence/reg-attendee-service/web/filter/logfilter"
	"github.com/eurofurence/reg-attendee-service/web/filter/securityfilter"
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
			corsfilter.Create(
				handlefilter.Create(handler))))
}

func BuildUnauthenticatedHandler(timeout string, handler filter.ContextAwareHandler) func(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := parseTimeout(timeout)
	return buildHandlerFunc(
		ctxfilter.Create(timeoutDuration,
			logfilter.Create(
				corsfilter.Create(
					handlefilter.Create(handler)))))
}

func BuildHandler(timeout string, handler filter.ContextAwareHandler, allowedGroups ...config.FixedTokenEnum) func(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := parseTimeout(timeout)
	return buildHandlerFunc(
		ctxfilter.Create(timeoutDuration,
			logfilter.Create(
				corsfilter.Create(
					securityfilter.Create(
						handlefilter.Create(handler), allowedGroups...)))))
}
