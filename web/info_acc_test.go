package web

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"rexis/rexis-go-attendee/docs"
	"rexis/rexis-go-attendee/web/util/media"
	"testing"
)

// -------------------------------------------
// acceptance tests for the info resource
// (status information about the microservice)
// -------------------------------------------

// see config and setup/teardown in setup_all_test.go

func TestHealthEndpoint(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they perform GET on the health endpoint")
	response := tstPerformGet("/info/health")

	docs.Then( "then OK is returned, and no further information is available")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	require.Equal(t, media.ContentTypeTextPlain, response.contentType, "unexpected response content type")
	require.Equal(t, "OK", response.body, "unexpected response from health endpoint")
}
