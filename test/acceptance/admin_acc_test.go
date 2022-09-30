package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
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
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "admr1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the admin information for the attendee")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestAdminDefaults_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admr2-")

	docs.Given("given the same regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access the admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstConfigFile(false, true, true))
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has registered")
	location1, attendee1 := tstRegisterAttendee(t, "admr3-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or anybody else's admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admr4-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is successful and the default admin information is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response.body)
}

func TestReadAdminInfo_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
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
	tstSetup(tstConfigFile(false, false, true))
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
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admw1-")

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Permissions: "admin",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admw2-")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Permissions: "admin",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstConfigFile(false, true, true))
	defer tstShutdown()

	docs.Given("given an existing attendee who is staff")
	location1, attendee1 := tstRegisterAttendee(t, "admw3-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to update their own admin information")
	body := admin.AdminInfoDto{
		Flags: "guest",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admw4-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information")
	body := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected response body")

	docs.Then("and the changed admin info can be read again")
	response2 := tstPerformGet(location1+"/admin", token)
	expectedAdminInfo := admin.AdminInfoDto{
		Id:            attendee1.Id,
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	tstRequireAdminInfoMatches(t, expectedAdminInfo, response2.body)
}

func TestAdminWrite_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee that does not exist")
	body := admin.AdminInfoDto{
		AdminComments: "existence is fleeting",
	}
	response := tstPerformPut("/api/rest/v1/attendees/789789/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestAdminWrite_InvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee with an invalid id")
	body := admin.AdminInfoDto{
		AdminComments: "kittens are cuter",
	}
	response := tstPerformPut("/api/rest/v1/attendees/puppy/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestAdminWrite_InvalidBody(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, _ := tstRegisterAttendee(t, "admw5-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information but send an invalid json body")
	body := "{{{{:::"
	response := tstPerformPut(location1+"/admin", body, token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.parse.error", "")
}

func TestAdminWrite_CannotChangeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "admw6-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the id")
	body := admin.AdminInfoDto{
		Id: "9999",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"id": []string{"id field must be empty or correctly assigned for incoming requests"}})
}

func TestAdminWrite_WrongFlagType(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admw7-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to set a non-admin only flag")
	body := admin.AdminInfoDto{
		Flags: "ev",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"flags": []string{"flags field must be a comma separated combination of any of guest"}})

	docs.Then("and the admin info is unchanged")
	response2 := tstPerformGet(location1+"/admin", token)
	expectedAdminInfo := admin.AdminInfoDto{
		Id: attendee1.Id,
	}
	tstRequireAdminInfoMatches(t, expectedAdminInfo, response2.body)
}

// TODO test dues changes caused by setting and removing guest status and corresponding status change logic

// helper functions

func tstRequireAdminInfoMatches(t *testing.T, expected admin.AdminInfoDto, body string) {
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(body, &adminInfo)
	require.EqualValues(t, expected, adminInfo, "admin data did not match expected values")
}
