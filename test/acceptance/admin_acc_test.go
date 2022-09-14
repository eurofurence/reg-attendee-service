package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

// ------------------------------------------
// acceptance tests for the admin subresource
// ------------------------------------------

// --- read access

func TestAdminDefaults_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admr1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to access the admin information")
	response := tstPerformGet(creationResponse.location+"/admin", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")
}

func TestAdminDefaults_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admr2-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t)

	docs.When("when they attempt to access the admin information")
	response := tstPerformGet(creationResponse.location+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer")
	token := tstValidStaffToken(t)

	docs.Given("who has made a valid registration")
	existingAttendee := tstBuildValidAttendee("admr3-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to access their own admin information")
	response := tstPerformGet(creationResponse.location+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and no body is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admr4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the admin information")
	response := tstPerformGet(creationResponse.location+"/admin", token)

	docs.Then("then the request is successful and the default admin information is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response.body, &adminInfo)

	expectedAdminInfo := admin.AdminInfoDto{
		Id: adminInfo.Id,
	}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

func TestReadAdminInfo_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the admin information for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/admin", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestReadAdminInfo_InvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the admin information for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/kittycat/admin", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

// --- write access

func TestAdminWrite_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admw1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Flags:         "",
		Permissions:   "admin",
		AdminComments: "",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(creationResponse.location+"/admin", tstValidAdminToken(t))
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response2.body, &adminInfo)
	expectedAdminInfo := admin.AdminInfoDto{Id: adminInfo.Id}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

func TestAdminWrite_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t)

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admw2-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Flags:         "",
		Permissions:   "admin",
		AdminComments: "",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(creationResponse.location+"/admin", tstValidAdminToken(t))
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response2.body, &adminInfo)
	expectedAdminInfo := admin.AdminInfoDto{Id: adminInfo.Id}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

func TestAdminWrite_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer")
	token := tstValidStaffToken(t)

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admw3-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Flags:         "",
		Permissions:   "admin",
		AdminComments: "",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(creationResponse.location+"/admin", tstValidAdminToken(t))
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response2.body, &adminInfo)
	expectedAdminInfo := admin.AdminInfoDto{Id: adminInfo.Id}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

func TestAdminWrite_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admw4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information")
	body := admin.AdminInfoDto{
		Flags:         "guest",
		Permissions:   "",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected response body")

	docs.Then("and the changed admin info can be read again")
	response2 := tstPerformGet(creationResponse.location+"/admin", token)
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response2.body, &adminInfo)

	expectedAdminInfo := admin.AdminInfoDto{
		Id:            adminInfo.Id,
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

func TestAdminWrite_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee that does not exist")
	body := admin.AdminInfoDto{
		Flags:         "guest",
		Permissions:   "",
		AdminComments: "set to guest",
	}
	response := tstPerformPut("/api/rest/v1/attendees/789789/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestAdminWrite_InvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee with an invalid id")
	body := admin.AdminInfoDto{
		Flags:         "guest",
		Permissions:   "",
		AdminComments: "set to guest",
	}
	response := tstPerformPut("/api/rest/v1/attendees/puppy/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestAdminWrite_InvalidBody(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("admw5-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information but send an invalid json body")
	body := "{{{{:::"
	response := tstPerformPut(creationResponse.location+"/admin", body, token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.parse.error", "")
}

func TestAdminWrite_CannotChangeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("admw6-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the id")
	body := admin.AdminInfoDto{
		Id: "9999",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"id": []string{"id field must be empty or correctly assigned for incoming requests"}})
}

func TestAdminWrite_WrongFlagType(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("admw7-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to set a non-admin only flag")
	body := admin.AdminInfoDto{
		Flags: "ev",
	}
	response := tstPerformPut(creationResponse.location+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"flags": []string{"flags field must be a comma separated combination of any of guest"}})

	docs.Then("and the admin info is unchanged")
	response2 := tstPerformGet(creationResponse.location+"/admin", token)
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response2.body, &adminInfo)

	expectedAdminInfo := admin.AdminInfoDto{
		Id: adminInfo.Id,
	}
	require.EqualValues(t, expectedAdminInfo, adminInfo, "admin data read did not match expected values")
}

// helper functions

func tstReadAdminInfo(t *testing.T, location string, bearerToken string) admin.AdminInfoDto {
	response := tstPerformGet(location, bearerToken)
	require.Equal(t, http.StatusOK, response.status)

	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response.body, &adminInfo)
	return adminInfo
}
