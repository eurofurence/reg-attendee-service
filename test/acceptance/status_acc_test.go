package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

// -------------------------------------------
// acceptance tests for the status subresource
// -------------------------------------------

// -- read status

func TestStatus_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	attendeeLocation, _ := tstRegisterAttendee(t, "stat1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status")
	response := tstPerformGet(attendeeLocation+"/status", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")
}

func TestStatus_UserDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given two existing attendees")
	_, attendee1 := tstRegisterAttendee(t, "stat2a-")
	location2, _ := tstRegisterAttendee(t, "stat2b-")

	docs.Given("given the first attendee logs in and is a regular user")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access somebody else's status")
	response := tstPerformGet(location2+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatus_UserAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "stat3-")

	docs.Given("given the attendee logs in and is a regular user")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they access their own status")
	_ = tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model
}

func TestStatus_StaffDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given two existing attendees")
	_, attendee1 := tstRegisterAttendee(t, "stat4a-")
	location2, _ := tstRegisterAttendee(t, "stat4b-")

	docs.Given("given the first attendee logs in and is staff")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access somebody else's status")
	response := tstPerformGet(location2+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatus_StaffAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee with no special privileges")
	location1, attendee1 := tstRegisterAttendee(t, "stat5-")

	docs.Given("given the attendee logs in")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they access their own status")
	_ = tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model
}

func TestStatus_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat6-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status for any attendee")
	response := tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and the default status is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	tstRequireAttendeeStatus(t, "new", response.body)
}

func TestStatus_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status for an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/panther/status", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestStatus_Nonexistent(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/status", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

// -- status history

func TestStatusHistory_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat20-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")
}

func TestStatusHistory_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "stat21-")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatusHistory_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has made a valid registration")
	location1, attendee1 := tstRegisterAttendee(t, "stat22-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatusHistory_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "stat23-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status history for any attendee")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is successful and the default status history is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	statusHistoryDto := status.StatusHistoryDto{}
	tstParseJson(response.body, &statusHistoryDto)

	require.Equal(t, 1, len(statusHistoryDto.StatusHistory))
	expectedStatusHistory := status.StatusHistoryDto{
		Id: attendee1.Id,
		StatusHistory: []status.StatusChange{{
			Timestamp: statusHistoryDto.StatusHistory[0].Timestamp,
			Status:    "new",
		}},
	}
	require.EqualValues(t, expectedStatusHistory, statusHistoryDto, "status history did not match expected value")
}

// helper functions

func tstRequireAttendeeStatus(t *testing.T, expected string, responseBody string) {
	statusDto := status.StatusDto{}
	tstParseJson(responseBody, &statusDto)

	expectedStatusDto := status.StatusDto{
		Status: expected,
	}
	require.EqualValues(t, expectedStatusDto, statusDto, "status did not match expected value")
}
