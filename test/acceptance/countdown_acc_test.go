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

// before both target times

func TestCountdownBeforeTarget_Public(t *testing.T) {
	docs.Given("given the configuration for public standard registration with a target date in the future")
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

func TestCountdownBeforeTarget_NeedsLogin(t *testing.T) {
	docs.Given("given the configuration for login-only standard registration with a target date in the future")
	tstSetup(tstConfigFile(true, false, false))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then the request is denied as unauthenticated (401)")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCountdownBeforeTarget_LoggedIn(t *testing.T) {
	docs.Given("given the configuration for login-only standard registration with a target date in the future")
	tstSetup(tstConfigFile(true, false, false))
	defer tstShutdown()

	docs.Given("given a logged in user who is not staff")
	token := tstValidUserToken(t, "1")

	docs.When("when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownBeforeTarget_LoggedInStaff(t *testing.T) {
	docs.Given("given the configuration for login-only staff registration with a both target dates in the future")
	tstSetup(tstConfigFile(true, true, false))
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, "1")

	docs.When("when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownBeforeTarget_LoggedInAdmin(t *testing.T) {
	docs.Given("given the configuration for login-only staff registration with a both target dates in the future")
	tstSetup(tstConfigFile(true, true, false))
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they request the countdown resource before the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0, meaning even an admin may not register yet")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

// between staff and public target time

func TestCountdownStaffregBetweenTargets_Public(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need NO login)")
	tstSetup(tstConfigFile(false, true, true))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownStaffregBetweenTargets_PublicLoggedInStaff(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need NO login)")
	tstSetup(tstConfigFile(false, true, true))
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, "1")

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown = 0, meaning the staffer will be allowed to register")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}

func TestCountdownStaffregBetweenTargets_NeedsLogin(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need to log in)")
	tstSetup(tstConfigFile(true, true, true))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then the request is denied as unauthenticated (401)")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCountdownStaffregBetweenTargets_LoggedIn(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need to log in)")
	tstSetup(tstConfigFile(true, true, true))
	defer tstShutdown()

	docs.Given("given a logged in user who is not staff")
	token := tstValidUserToken(t, "1")

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

func TestCountdownStaffregBetweenTargets_LoggedInStaff(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need to log in)")
	tstSetup(tstConfigFile(true, true, true))
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, "1")

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown = 0, meaning the staffer will be allowed to register")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}

func TestCountdownStaffregBetweenTargets_LoggedInAdminNotSpecial(t *testing.T) {
	docs.Given("given the configuration for staff registration (normal users need to log in)")
	tstSetup(tstConfigFile(true, true, true))
	defer tstShutdown()

	docs.Given("given an admin who does not have the staff role")
	token := tstValidAdminToken(t)

	docs.When("when they request the countdown resource after the staff target time, but before the public target time")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown > 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds > 0, "unexpected countdown value is not positive")
}

// after public target time

func TestCountdownAfterTarget_Public(t *testing.T) {
	docs.Given("given the configuration for public standard registration with a target date in the past")
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

func TestCountdownAfterTarget_NeedsLogin(t *testing.T) {
	docs.Given("given the configuration for login-only standard registration with a target date in the past")
	tstSetup(tstConfigFile(true, false, true))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they request the countdown resource after the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then the request is denied as unauthenticated (401)")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCountdownAfterTarget_LoggedIn(t *testing.T) {
	docs.Given("given the configuration for login-only standard registration with a target date in the past")
	tstSetup(tstConfigFile(true, false, true))
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, "1")

	docs.When("when they request the countdown resource after the target time has been reached")
	response := tstPerformGet("/api/rest/v1/countdown", token)

	docs.Then("then a valid response is sent with countdown = 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}

// testing the mock

func TestMockedCountdownAfterTarget_Public(t *testing.T) {
	docs.Given("given the configuration for public standard registration with a target date in the future")
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

func TestMockedCountdownAfterTarget_LoggedIn(t *testing.T) {
	docs.Given("given the configuration for login-only standard registration with a target date in the future")
	tstSetup(tstConfigFile(true, false, false))
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, "1")

	docs.When("when they request the mocked countdown resource before the target time has been reached, but pass a mock time after the target")
	response := tstPerformGet("/api/rest/v1/countdown?currentTime=2030-12-22T14:33:20-01:00", token)

	docs.Then("then a valid response is sent with countdown = 0")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	responseDto := countdown.CountdownResultDto{}
	tstParseJson(response.body, &responseDto)
	require.True(t, responseDto.CountdownSeconds == 0, "unexpected countdown value is not zero")
}
