package acceptance

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"github.com/jumpy-squirrel/rexis-go-attendee/docs"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/util/media"
	"testing"
)

// -------------------------------------------
// acceptance tests for the info resource
// (status information about the microservice)
// -------------------------------------------

// see config and setup/teardown in setup_acc_test.go

func TestHealthEndpoint(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they perform GET on the health endpoint")
	response := tstPerformGet("/info/health")

	docs.Then( "then OK is returned, and no further information is available")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	require.Equal(t, media.ContentTypeTextPlain, response.contentType, "unexpected response content type")
	require.Equal(t, "OK", response.body, "unexpected response from health endpoint")
}

func TestErrorFallback(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When("when they perform GET on an unimplemented endpoint")
	response := tstPerformGet("/info/does-not-exist")

	docs.Then( "then they receive a 404 error with no body")
	require.Equal(t, http.StatusNotFound, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected body")
}
