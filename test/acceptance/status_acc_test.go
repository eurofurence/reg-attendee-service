package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// -------------------------------------------
// acceptance tests for the status subresource
// -------------------------------------------

// --- read access

// -- status

func TestStatus_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status")
	response := tstPerformGet(creationResponse.location+"/status", token)

	docs.Then("then the request is denied as unauthenticated (401) and no body is returned")
	require.Equal(t, http.StatusUnauthorized, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

// note: users get their own current status as a read-only field in the attendee resource

func TestStatus_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a regular authenticated attendee")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. This is ok because only the old regsys will know the user / admin tokens.")
	token := tstValidUserToken(t)

	docs.When("when they attempt to access their own or somebody else's status")
	response := tstPerformGet(creationResponse.location+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	require.Equal(t, http.StatusForbidden, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

func TestStatus_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. This is ok because only the old regsys will know the user / admin tokens.")
	token := tstValidStaffToken(t)

	docs.Given("who has made a valid registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to access their own or somebody else's status")
	response := tstPerformGet(creationResponse.location+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	require.Equal(t, http.StatusForbidden, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

func TestStatus_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status for any attendee")
	response := tstPerformGet(creationResponse.location+"/status", token)

	docs.Then("then the request is successful and the default status is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)

	expectedStatus := status.StatusDto{
		Status: "new",
	}
	require.EqualValues(t, expectedStatus, statusDto, "status did not match expected value")
}

// -- status history

func TestStatusHistory_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status history")
	response := tstPerformGet(creationResponse.location+"/status-history", token)

	docs.Then("then the request is denied as unauthenticated (401) and no body is returned")
	require.Equal(t, http.StatusUnauthorized, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

// note: users get their own current status as a read-only field in the attendee resource

func TestStatusHistory_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a regular authenticated attendee")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. This is ok because only the old regsys will know the user / admin tokens.")
	token := tstValidUserToken(t)

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(creationResponse.location+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	require.Equal(t, http.StatusForbidden, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

func TestStatusHistory_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. This is ok because only the old regsys will know the user / admin tokens.")
	token := tstValidStaffToken(t)

	docs.Given("who has made a valid registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(creationResponse.location+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	require.Equal(t, http.StatusForbidden, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

func TestStatusHistory_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status history for any attendee")
	response := tstPerformGet(creationResponse.location+"/status-history", token)

	docs.Then("then the request is successful and the default status history is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	statusHistoryDto := status.StatusHistoryDto{}
	tstParseJson(response.body, &statusHistoryDto)

	require.Equal(t, 1, len(statusHistoryDto.StatusHistory))
	expectedStatusHistory := status.StatusHistoryDto{
		Id: statusHistoryDto.Id,
		StatusHistory: []status.StatusChange{{
			Timestamp: statusHistoryDto.StatusHistory[0].Timestamp,
			Status:    "new",
		}},
	}
	require.EqualValues(t, expectedStatusHistory, statusHistoryDto, "status history did not match expected value")
}

// helper functions
