package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/countdown"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// ------------------------------------------
// acceptance tests for the countdown resource
// ------------------------------------------

func TestCountdownBeforeTarget(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFileBeforeTarget)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When( "when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", tstNoToken())

	docs.Then( "then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownAfterTarget(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When( "when they request the countdown resource after the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", tstNoToken())

	docs.Then( "then a valid response is sent with countdown = 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}