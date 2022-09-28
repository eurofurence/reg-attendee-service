package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/countdown"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// ------------------------------------------
// acceptance tests for the countdown resource
// ------------------------------------------

func TestCountdownBeforeTarget(t *testing.T) {
	docs.Given("given the configuration for standard registration with a target date in the future")
	tstSetup(tstConfigFile(false, false, false))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownAfterTarget(t *testing.T) {
	docs.Given("given the configuration for standard registration with a target date in the past")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource after the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown = 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}

func TestMockedCountdownAfterTarget(t *testing.T) {
	docs.Given("given the configuration for standard registration with a target date in the future")
	tstSetup(tstConfigFile(false, false, false))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the mocked countdown resource before the target time has been reached, but pass a mock time after the target")
	response := tstPerformGet("/api/rest/v1/countdown?currentTime=2030-12-22T14:33:20-01:00", token)

	docs.Then("then a valid response is sent with countdown = 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}
