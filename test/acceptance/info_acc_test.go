package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// -------------------------------------------
// acceptance tests for the info resource
// (status information about the microservice)
// -------------------------------------------

// see config and setup/teardown in setup_acc_test.go

func TestHealthEndpoint(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they perform GET on the health endpoint")
	response := tstPerformGet("/info/health", tstNoToken())

	docs.Then("then OK is returned, and no further information is available")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	require.Equal(t, media.ContentTypeTextPlain, response.contentType, "unexpected response content type")
	require.Equal(t, "OK", response.body, "unexpected response from health endpoint")
}

func TestErrorFallback(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they perform GET on an unimplemented endpoint")
	response := tstPerformGet("/info/does-not-exist", tstNoToken())

	docs.Then("then they receive a 404 error")
	require.Equal(t, http.StatusNotFound, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected body")
}
