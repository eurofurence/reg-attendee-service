package securityfilter

import (
	"context"
	"github.com/go-http-utils/headers"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// setup

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
	tstSetupConfig()
}

func tstShutdown() {

}

func tstSetupConfig() {
	config.LoadTestingConfigurationFromPathOrAbort("../../../test/testconfig.yaml")
}

// tests

func TestMissingAuthHeader(t *testing.T) {
	docs.Description("a missing authorization header should lead to http status unauthorized")
	w := httptest.NewRecorder()
	r := tstMockGetRequest("omit")

	cut := Create(nil)
	cut.Handle(context.TODO(), w, r)

	response := w.Result()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected response status")
}

func TestEmptyAuthHeader(t *testing.T) {
	docs.Description("an empty authorization header should lead to http status unauthorized")
	w := httptest.NewRecorder()
	r := tstMockGetRequest("")

	cut := Create(nil)
	cut.Handle(context.TODO(), w, r)

	response := w.Result()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected response status")
}

func TestAuthHeaderNotBearer(t *testing.T) {
	docs.Description("authorization header without 'Bearer ' heading should lead to http status unauthorized")
	w := httptest.NewRecorder()
	r := tstMockGetRequest("Basic 128377389")

	cut := Create(nil)
	cut.Handle(context.TODO(), w, r)

	response := w.Result()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected response status")
}

func TestAuthHeaderWrongToken(t *testing.T) {
	docs.Description("authorization header with wrong token should lead to http status unauthorized")
	w := httptest.NewRecorder()
	r := tstMockGetRequest("Bearer hallo")

	cut := Create(nil)
	cut.Handle(context.TODO(), w, r)

	response := w.Result()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected response status")
}

// helpers

func tstMockGetRequest(authHeaderValue string) *http.Request {
	r, err := http.NewRequest(http.MethodGet, "/unused/url", nil)
	if err != nil {
		log.Fatal(err)
	}
	if authHeaderValue != "omit" {
		r.Header.Set(headers.Authorization, authHeaderValue)
	}
	return r
}
