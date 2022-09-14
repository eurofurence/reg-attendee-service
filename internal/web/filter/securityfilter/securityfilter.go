package securityfilter

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/go-http-utils/headers"
	"net/http"
	"net/url"
	"strings"
)

type SecurityFilter struct {
	wrappedFilter filter.Filter
	allowedGroups []config.FixedTokenEnum
}

// if allowedGroups is the empty list, no-one can call this endpoint
func Create(wrappedFilter filter.Filter, allowedGroups ...config.FixedTokenEnum) filter.Filter {
	return &SecurityFilter{wrappedFilter: wrappedFilter, allowedGroups: allowedGroups}
}

func (f *SecurityFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	err := f.checkAuthenticated(ctx, w, r)
	if err != nil {
		f.unauthenticatedError(ctx, w, r, err)
		return
	}

	err = f.checkAuthorized(ctx, w, r)
	if err != nil {
		f.unauthorizedError(ctx, w, r, err)
		return
	}

	f.wrappedFilter.Handle(ctx, w, r)
}

func (f *SecurityFilter) checkAuthenticated(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bearerTokenHeader := r.Header.Get(headers.Authorization)
	if bearerTokenHeader != "" && strings.HasPrefix(bearerTokenHeader, "Bearer ") {
		bearerToken := bearerTokenHeader[7:] + ""

		return f.checkTokenValidAndAddToContext(ctx, bearerToken)
	} else {
		logging.Ctx(ctx).Warn("invalid or missing authorization header, denying access, not authenticated")
		return errors.New("missing " + headers.Authorization + " header with bearer token")
	}
}

func (f *SecurityFilter) checkTokenValidAndAddToContext(ctx context.Context, bearerToken string) error {
	allGroups := config.AllAvailableFixedTokenGroups()
	if _, isValid := isTokenValidForOneOfTheGroups(bearerToken, allGroups); isValid {
		// success: authenticated
		ctxvalues.SetBearerToken(ctx, bearerToken)
		return nil
	} else {
		logging.Ctx(ctx).Warn("invalid bearer token, denying access")
		return errors.New("invalid bearer token")
	}
}

func (f *SecurityFilter) checkAuthorized(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	bearerToken := ctxvalues.BearerToken(ctx)
	if matchesGroup, isValid := isTokenValidForOneOfTheGroups(bearerToken, f.allowedGroups); isValid {
		// success: authorized
		ctxvalues.SetAuthorizedAsGroup(ctx, matchesGroup)
		return nil
	} else {
		logging.Ctx(ctx).Warn("unauthorized access attempt, denying access, not authorized")
		return errors.New("you are not unauthorized for this operation - the attempt has been logged")
	}
}

func isTokenValidForOneOfTheGroups(token string, groups []config.FixedTokenEnum) (config.FixedTokenEnum, bool) {
	for _, grp := range groups {
		expectedToken, err := config.FixedToken(grp)
		if err != nil {
			return -1, false
		}
		if expectedToken == token {
			return grp, true
		}
	}
	return -1, false
}

// 401 unauthorized means: invalid authentication (no token, or invalid token)
// 403 forbidden means: you don't have the necessary permissions

func (f *SecurityFilter) unauthenticatedError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	ctlutil.ErrorHandler(ctx, w, r, "auth.unauthorized", http.StatusUnauthorized, url.Values{"details": []string{err.Error()}})
}

func (f *SecurityFilter) unauthorizedError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	ctlutil.ErrorHandler(ctx, w, r, "auth.forbidden", http.StatusForbidden, url.Values{"details": []string{err.Error()}})
}
