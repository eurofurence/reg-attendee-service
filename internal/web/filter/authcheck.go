package filter

import (
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"net/http"
)

func HasRoleOrApiToken(role string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsRole(ctx, role) {
			handler(w, r)
		} else {
			culprit := ctxvalues.Subject(ctx)
			if culprit != "" {
				ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized for this operation - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for role %s by %s", role, culprit))
			} else {
				ctlutil.UnauthenticatedError(ctx, w, r, "missing Authorization header with bearer token", "anonymous access attempt")
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
			ctlutil.UnauthenticatedError(ctx, w, r, "missing Authorization header with bearer token", "anonymous access attempt")
		}
	}
}

// IsSubjectOrRoleOrApiToken cannot be used as a filter because the subject needs to be loaded from the database first (part of the attendee admin data). Use in your handler functions.
//
// Do not forget to return from the handler if an error is returned!
func IsSubjectOrRoleOrApiToken(w http.ResponseWriter, r *http.Request, subject string, role string) error {
	ctx := r.Context()
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsRole(ctx, role) || ctxvalues.Subject(ctx) == subject {
		return nil
	} else {
		culprit := ctxvalues.Subject(ctx)
		ctlutil.UnauthorizedError(ctx, w, r, "you are not authorized to access this data - the attempt has been logged", fmt.Sprintf("unauthorized access attempt for %s by %s", subject, culprit))
		return errors.New("neither api token nor subject match - unauthorized")
	}
}
