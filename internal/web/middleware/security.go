package middleware

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/authservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctlutil"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-http-utils/headers"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
)

// --- getting the values from the request ---

func fromCookie(r *http.Request, cookieName string) string {
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
	headerValue := r.Header.Get(headers.Authorization)

	if !strings.HasPrefix(headerValue, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(headerValue, "Bearer ")
}

func fromApiTokenHeader(r *http.Request) string {
	return r.Header.Get(media.HeaderXApiKey)
}

// --- validating the individual pieces ---

// important - if any of these return an error, you must abort processing via "return" and log the error message

func checkApiToken_MustReturnOnError(ctx context.Context, apiTokenValue string) (success bool, err error) {
	if apiTokenValue != "" {
		// ignore jwt if set (may still need to pass it through to other service)
		if apiTokenValue == config.FixedApiToken() {
			ctxvalues.SetApiToken(ctx, apiTokenValue)
			return true, nil
		} else {
			return false, errors.New("request failed presented api token check, denying")
		}
	}
	return false, nil
}

func audienceMatchesOrNotConfigured(userInfo authservice.UserInfoResponse) bool {
	if config.OidcAllowedAudience() != "" {
		if len(userInfo.Audiences) != 1 || userInfo.Audiences[0] != config.OidcAllowedAudience() {
			return false
		}
	}
	return true
}

func checkAccessToken_MustReturnOnError(ctx context.Context, accessTokenValue string, verifyAudience bool) (success bool, err error) {
	if accessTokenValue != "" {
		if authservice.Get().IsEnabled() {
			ctxvalues.SetAccessToken(ctx, accessTokenValue) // need this for userinfo call

			userInfo, err := authservice.Get().UserInfo(ctx)
			if err != nil {
				return false, fmt.Errorf("request failed access token check, denying: %s", err.Error())
			}

			audienceOk := audienceMatchesOrNotConfigured(userInfo)
			if verifyAudience && !audienceOk {
				return false, errors.New("token audience does not match")
			}

			ctxvalues.SetName(ctx, userInfo.Name)
			ctxvalues.SetSubject(ctx, userInfo.Subject)
			ctxvalues.SetEmail(ctx, userInfo.Email)
			ctxvalues.SetEmailVerified(ctx, userInfo.EmailVerified)

			// rebuild groups list just in case (removes groups the user doesn't actually have)
			ctxvalues.ClearAuthorizedGroups(ctx)
			for _, group := range userInfo.Groups {
				ctxvalues.SetAuthorizedAsGroup(ctx, group)
			}

			return true, nil
		} else {
			return false, errors.New("request failed access token check, denying: no userinfo endpoint configured")
		}
	}
	return false, nil
}

func keyFuncForKey(rsaPublicKey *rsa.PublicKey) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		return rsaPublicKey, nil
	}
}

type CustomClaims struct {
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups,omitempty"`
	Name          string   `json:"name"`
	Avatar        string   `json:"avatar"`
}

type AllClaims struct {
	jwt.RegisteredClaims
	CustomClaims
}

func checkIdToken_MustReturnOnError(ctx context.Context, idTokenValue string) (success bool, err error) {
	if idTokenValue != "" {
		tokenString := strings.TrimSpace(idTokenValue)

		errorMessage := ""
		for _, key := range config.OidcKeySet() {
			claims := AllClaims{}
			token, err := jwt.ParseWithClaims(tokenString, &claims, keyFuncForKey(key), jwt.WithValidMethods([]string{"RS256", "RS512"}))
			if err == nil && token.Valid {
				parsedClaims, ok := token.Claims.(*AllClaims)
				if ok {
					if config.OidcAllowedAudience() != "" {
						if len(parsedClaims.Audience) != 1 || parsedClaims.Audience[0] != config.OidcAllowedAudience() {
							return false, errors.New("token audience does not match")
						}
					}

					if config.OidcAllowedIssuer() != "" {
						if parsedClaims.Issuer != config.OidcAllowedIssuer() {
							return false, errors.New("token issuer does not match")
						}
					}

					ctxvalues.SetIdToken(ctx, idTokenValue)
					ctxvalues.SetEmail(ctx, parsedClaims.Email)
					ctxvalues.SetEmailVerified(ctx, parsedClaims.EmailVerified)
					ctxvalues.SetName(ctx, parsedClaims.Name)
					ctxvalues.SetSubject(ctx, parsedClaims.Subject)
					ctxvalues.SetAvatar(ctx, parsedClaims.Avatar)
					for _, group := range parsedClaims.Groups {
						ctxvalues.SetAuthorizedAsGroup(ctx, group)
					}

					return true, nil
				}
				errorMessage = "empty claims substructure"
			} else if err != nil {
				errorMessage = err.Error()
			} else {
				errorMessage = "token parsed but invalid"
			}
		}
		return false, errors.New(errorMessage)
	}
	return false, nil
}

func allow(actualMethod string, actualUrlPath string, allowedMethod string, allowedUrlPath string) bool {
	return actualMethod == allowedMethod && actualUrlPath == allowedUrlPath
}

func allowPrefixSuffix(actualMethod string, actualUrlPath string, allowedMethod string, allowedPrefix string, allowedSuffix string) bool {
	return actualMethod == allowedMethod && strings.HasPrefix(actualUrlPath, allowedPrefix) && strings.HasSuffix(actualUrlPath, allowedSuffix)
}

func allowPrefixContains(actualMethod string, actualUrlPath string, allowedMethod string, allowedPrefix string, mustContain string) bool {
	return actualMethod == allowedMethod && strings.HasPrefix(actualUrlPath, allowedPrefix) && strings.Contains(actualUrlPath, mustContain)
}

func canSkipAccessTokenCheckWithCookieAuth(method string, urlPath string) bool {
	// positive list for request URLs and Methods where the access token check may be skipped
	// (performance critical + cannot be used for data extraction only!)
	return allow(method, urlPath, http.MethodGet, "/api/rest/v1/countdown") ||
		allow(method, urlPath, http.MethodGet, "/api/rest/v1/attendees") || // list my badge numbers
		allow(method, urlPath, http.MethodPost, "/api/rest/v1/attendees") // only receiving data
}

func canSkipAudienceCheckWithAccessToken(method string, urlPath string) bool {
	// positive list for request URLs and Methods where valid tokens for any audience are accepted,
	// meaning the identity provider signed them for another client, such as the dealer system.
	// (should not be set for endpoints that allow extraction of personal data!)
	return allow(method, urlPath, http.MethodGet, "/api/rest/v1/attendees") || // list my badge numbers
		allowPrefixSuffix(method, urlPath, http.MethodGet, "/api/rest/v1/attendees/", "/status") || // current status value only
		allowPrefixContains(method, urlPath, http.MethodGet, "/api/rest/v1/attendees/", "/flags/") || // yes/no only
		allowPrefixContains(method, urlPath, http.MethodGet, "/api/rest/v1/attendees/", "/options/") || // yes/no only
		allowPrefixContains(method, urlPath, http.MethodGet, "/api/rest/v1/attendees/", "/packages/") // yes/no only
}

// --- top level ---

func checkAllAuthentication_MustReturnOnError(ctx context.Context, method string, urlPath string, apiTokenHeaderValue string, authHeaderValue string, idTokenCookieValue string, accessTokenCookieValue string) (userFacingErrMsg string, err error) {
	// try api token first
	success, err := checkApiToken_MustReturnOnError(ctx, apiTokenHeaderValue)
	if err != nil {
		return "invalid api token", err
	}
	if success {
		return "", nil
	}

	// now try authorization header (gives only access token, so MUST use userinfo endpoint)
	success, err = checkAccessToken_MustReturnOnError(ctx, authHeaderValue, !canSkipAudienceCheckWithAccessToken(method, urlPath))
	if err != nil {
		return "invalid bearer token", err
	}
	if success {
		return "", nil
	}

	// now try cookie pair
	success, err = checkIdToken_MustReturnOnError(ctx, idTokenCookieValue)
	if err != nil {
		return "invalid id token in cookie", err
	}
	if success {
		if canSkipAccessTokenCheckWithCookieAuth(method, urlPath) {
			// ensure propagation even if not checked
			ctxvalues.SetAccessToken(ctx, accessTokenCookieValue)
			return "", nil
		}

		// TODO allow skipping non-working audience check for now
		success2, err := checkAccessToken_MustReturnOnError(ctx, accessTokenCookieValue, false)
		if err != nil {
			return "invalid or missing access token in cookie", err
		}
		if success2 {
			return "", nil
		}
	}

	// not supplying authorization is a valid use case, there are endpoints that allow anonymous access
	return "", nil
}

// --- middleware validating the values and adding to context values ---

func TokenValidator(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		apiTokenHeaderValue := fromApiTokenHeader(r)
		authHeaderValue := fromAuthHeader(r)
		idTokenCookieValue := fromCookie(r, config.OidcIdTokenCookieName())
		accessTokenCookieValue := fromCookie(r, config.OidcAccessTokenCookieName())

		userFacingErrMsg, err := checkAllAuthentication_MustReturnOnError(ctx, r.Method, r.URL.Path, apiTokenHeaderValue, authHeaderValue, idTokenCookieValue, accessTokenCookieValue)
		if err != nil {
			ctlutil.UnauthenticatedError(ctx, w, r, userFacingErrMsg, err.Error())
			return
		}

		next.ServeHTTP(w, r)
		return
	}
	return http.HandlerFunc(fn)
}

// --- accessors see ctxvalues ---
