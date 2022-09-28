package middleware

import (
	"crypto/rsa"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/go-http-utils/headers"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
)

// --- getting the values from the request ---

func fromCookie(r *http.Request) string {
	cookieName := config.OidcTokenCookieName()
	if cookieName == "" {
		// ok if not configured, don't accept cookies then
		return ""
	}

	authCookie, _ := r.Cookie(cookieName)
	if authCookie == nil {
		// missing cookie is not considered an error, either
		return ""
	}

	return authCookie.Value
}

func fromAuthHeader(r *http.Request) string {
	return r.Header.Get(headers.Authorization)
}

func fromAuthHeaderOrCookie(r *http.Request) string {
	h := fromAuthHeader(r)
	if h == "" {
		return fromCookie(r)
	} else {
		return h
	}
}

func fromApiTokenHeader(r *http.Request) string {
	return r.Header.Get("X-Api-Token")
}

// --- middleware validating the values and adding to context values ---

func keyFuncForKey(rsaPublicKey *rsa.PublicKey) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		return rsaPublicKey, nil
	}
}

// TODO example - no idea if this matches the idp claims structure - compare to room service!

type GlobalClaims struct {
	Name  string   `json:"name"`
	EMail string   `json:"email"`
	Roles []string `json:"roles"`
}

type CustomClaims struct {
	Global GlobalClaims `json:"global"`
}

type AllClaims struct {
	jwt.RegisteredClaims
	CustomClaims
}

func TokenValidator(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// try api token first
		apiTokenValue := fromApiTokenHeader(r)
		if apiTokenValue != "" {
			// ignore jwt if set (may still need to pass it through to other service)
			if apiTokenValue == config.FixedApiToken() {
				ctxvalues.SetApiToken(ctx, apiTokenValue)
				next.ServeHTTP(w, r)
			} else {
				ctlutil.UnauthenticatedError(ctx, w, r, "invalid api token", "request supplied invalid api token, denying")
			}
			return
		}

		// now try bearer token
		bearerTokenValue := fromAuthHeaderOrCookie(r)
		if bearerTokenValue != "" {
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(bearerTokenValue, bearerPrefix) {
				ctlutil.UnauthenticatedError(ctx, w, r, "value of Authorization header did not start with 'Bearer '", "request supplied malformed bearer token, denying")
				return
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(bearerTokenValue, bearerPrefix))

			errorMessage := ""
			for _, key := range config.OidcKeySet() {
				claims := AllClaims{}
				token, err := jwt.ParseWithClaims(tokenString, &claims, keyFuncForKey(key), jwt.WithValidMethods([]string{"RS256", "RS512"}))
				if err == nil && token.Valid {
					parsedClaims, ok := token.Claims.(*AllClaims)
					if ok {
						// TODO this is probably not the exact token structure
						ctxvalues.SetBearerToken(ctx, bearerTokenValue)
						ctxvalues.SetEmail(ctx, parsedClaims.Global.EMail)
						ctxvalues.SetName(ctx, parsedClaims.Global.Name)
						ctxvalues.SetSubject(ctx, parsedClaims.Subject)
						for _, role := range parsedClaims.Global.Roles {
							ctxvalues.SetAuthorizedAsRole(ctx, role)
						}

						next.ServeHTTP(w, r)
						return
					}
					errorMessage = "empty claims substructure"
				} else if err != nil {
					errorMessage = err.Error()
				} else {
					errorMessage = "token parsed but invalid"
				}
			}
			ctlutil.UnauthenticatedError(ctx, w, r, "invalid bearer token", errorMessage)
			return
		}

		// not supplying either is a valid use case, there are endpoints that allow anonymous access
		next.ServeHTTP(w, r)
		return
	}
	return http.HandlerFunc(fn)
}

// --- accessors see ctxvalues ---
