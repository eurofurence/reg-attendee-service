package securityfilter

import (
	"context"
	"errors"
	"github.com/go-http-utils/headers"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxvalues"
	"net/http"
	"strings"
)

type SecurityFilter struct {
	wrappedFilter filter.Filter
}

func Create(wrappedFilter filter.Filter) filter.Filter {
	return &SecurityFilter{wrappedFilter: wrappedFilter}
}

func (f *SecurityFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	err := f.checkAuthenticated(ctx, w, r)
	if err != nil {
		f.unauthenticatedError(ctx, w, r)
		return
	}

	// this is the place to add unauthorized check (provide a bool-valued function for declarative-like security?)

	f.wrappedFilter.Handle(ctx, w, r)
}

func (f *SecurityFilter) checkAuthenticated(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bearerTokenHeader := r.Header.Get(headers.Authorization)
	if bearerTokenHeader != "" && strings.HasPrefix(bearerTokenHeader, "Bearer ") {
		bearerToken := bearerTokenHeader[7:] + ""

		err := f.checkToken(ctx, bearerToken)
		if err != nil {
			return err
		} else {
			ctxvalues.SetBearerToken(ctx, bearerToken)
			return nil
		}
	} else {
		logging.Ctx(ctx).Warn("invalid or missing authorization header, denying access")
		return errors.New("missing " + headers.Authorization + " header with bearer token")
	}
}

func (f *SecurityFilter) checkToken(ctx context.Context, bearerToken string) error {
	if (bearerToken != config.FixedToken()) {
		logging.Ctx(ctx).Warn("invalid bearer token, denying access")
		return errors.New("invalid bearer token")
	} else {
		return nil
	}
}

// 401 unauthorized means: invalid authentication (no token, or invalid token)
// 403 forbidden means: you don't have the necessary permissions

func (f *SecurityFilter) unauthenticatedError(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// for now we just have a fixed token, so only this error can happen
	w.WriteHeader(http.StatusUnauthorized)
	ctxvalues.SetHttpStatus(ctx, http.StatusUnauthorized)
}
