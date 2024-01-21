// Package selfclient provides a very simple self API client.
package selfclient

import (
	"context"
	"errors"
	"fmt"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/go-http-utils/headers"
	"net/http"
	"strings"
	"time"
)

var (
	client       aurestclientapi.Client
	baseUrl      string
	idToken      string
	accessToken  string
	cookieDomain string
)

var (
	UnauthorizedError = errors.New("got unauthorized from userinfo endpoint")
	DownstreamError   = errors.New("downstream unavailable - see log for details")
)

func requestManipulator(ctx context.Context, r *http.Request) {
	r.AddCookie(&http.Cookie{
		Name:     config.OidcIdTokenCookieName(),
		Value:    idToken,
		Domain:   cookieDomain,
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	r.AddCookie(&http.Cookie{
		Name:     config.OidcAccessTokenCookieName(),
		Value:    accessToken,
		Domain:   cookieDomain,
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func Setup() error {
	httpClient, err := auresthttpclient.New(0, nil, requestManipulator)
	if err != nil {
		return err
	}

	requestLoggingClient := aurestlogging.New(httpClient)

	client = requestLoggingClient

	baseUrl = config.GeneratorBaseUrl()
	idToken = config.GeneratorIdToken()
	accessToken = config.GeneratorAccessToken()
	cookieDomain = config.GeneratorCookieDomain()

	if baseUrl == "" || idToken == "" || accessToken == "" || cookieDomain == "" {
		return errors.New("missing parameters for setting up selfclient - cannot run generator")
	}

	return nil
}

func errByStatus(err error, status int) error {
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized {
		return UnauthorizedError
	}
	if status >= 300 {
		return DownstreamError
	}
	return nil
}

func SendRegistration(ctx context.Context, dto *attendee.AttendeeDto) (string, error) {
	url := fmt.Sprintf("%s/api/rest/v1/attendees", baseUrl)
	response := aurestclientapi.ParsedResponse{}
	err := client.Perform(ctx, http.MethodPost, url, dto, &response)

	// parse location header
	loc := ""
	if val, ok := response.Header[headers.Location]; ok {
		loc = val[0]
	}
	id := strings.TrimPrefix(loc, url+"/")

	return id, errByStatus(err, response.Status)
}
