package web

import (
	"rexis/rexis-go-attendee/docs"
	"testing"
)

// see config and setup/teardown in setup_all_test.go

func TestHealthEndpoint(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they perform GET on the health endpoint")
	actualbody := tstPerformGetReturnBody("/info/health")

	docs.Then( "then OK is returned, and no further information is available")
	tstAssertStringEqual(t, "unexpected response from health endpoint", "OK", actualbody)
}
