package filter

import (
	"context"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"net/http"
)

// checkInternalAdminRequestHeader is a temporary safety measure until we have 2FA for admins.
//
// enforce extra internal request header for admin requests (header blocked for external requests)
//
// TODO: remove this workaround
func checkInternalAdminRequestHeaderForGroup(ctx context.Context, r *http.Request, group string) bool {
	if group == config.OidcAdminGroup() {
		adminRequestHeaderValue := r.Header.Get("X-Admin-Request")
		if adminRequestHeaderValue != "available" {
			aulogging.Logger.Ctx(ctx).Warn().Print("X-Admin-Request header was not set correctly!")
			return false
		}
	}
	return true
}

func HasGroupOrApiToken(group string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctxvalues.HasApiToken(ctx) || (ctxvalues.IsAuthorizedAsGroup(ctx, group) && checkInternalAdminRequestHeaderForGroup(ctx, r, group)) {
			handler(w, r)
		} else {
			culprit := ctxvalues.Subject(ctx)
			if culprit != "" {
				ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this operation - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for group %s by %s", group, culprit))
			} else {
				ctlutil.UnauthenticatedError(ctx, w, r, "you must be logged in for this operation", "anonymous access attempt")
			}
		}
	}
}

func LoggedInOrApiToken(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctxvalues.HasApiToken(ctx) || ctxvalues.Subject(ctx) != "" {
			handler(w, r)
		} else {
			ctlutil.UnauthenticatedError(ctx, w, r, "you must be logged in for this operation", "anonymous access attempt")
		}
	}
}

func LoggedIn(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctxvalues.Subject(ctx) != "" {
			handler(w, r)
		} else {
			ctlutil.UnauthenticatedError(ctx, w, r, "you must be logged in for this operation", "anonymous access attempt")
		}
	}
}

// IsSubjectOrGroupOrApiToken cannot be used as a filter because the subject needs to be loaded from the database first (part of the attendee admin data). Use in your handler functions.
//
// Do not forget to return from the handler if an error is returned!
func IsSubjectOrGroupOrApiToken(w http.ResponseWriter, r *http.Request, subject string, group string) error {
	ctx := r.Context()
	if ctxvalues.HasApiToken(ctx) || ctxvalues.Subject(ctx) == subject || (ctxvalues.IsAuthorizedAsGroup(ctx, group) && checkInternalAdminRequestHeaderForGroup(ctx, r, group)) {
		return nil
	} else {
		culprit := ctxvalues.Subject(ctx)
		ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized to access this data - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for %s by %s", subject, culprit))
		return errors.New("neither api token nor subject match - unauthorized")
	}
}

func IsGroupOrApiTokenCond(r *http.Request, group string) bool {
	ctx := r.Context()
	return ctxvalues.HasApiToken(ctx) || (ctxvalues.IsAuthorizedAsGroup(ctx, group) && checkInternalAdminRequestHeaderForGroup(ctx, r, group))
}
