package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"net/http"
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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

// --- write access

func TestAdminWrite_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.Given("given an existing attendee right after registration")
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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
	existingAttendee := tstBuildValidAttendee("adm1-")
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

// TODO validation stuff (wrong flags, trying to change id etc.)

// helper functions

func tstReadAdminInfo(t *testing.T, location string, bearerToken string) admin.AdminInfoDto {
	response := tstPerformGet(location, bearerToken)
	require.Equal(t, http.StatusOK, response.status)

	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response.body, &adminInfo)
	return adminInfo
}
