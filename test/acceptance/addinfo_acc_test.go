package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/addinfo"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

// ----------------------------------------------------
// acceptance tests for the additional info subresource
// ----------------------------------------------------

// --- read access

func TestGetAdditionalInfo_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "air1-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"air1":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the additional info for the attendee")
	response := tstPerformGet(location1+"/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestGetAdditionalInfo_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "air2-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"air2":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to access even their own additional info but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet(location1+"/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")
}

func TestGetAdditionalInfo_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "air3-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"air3":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given the attendee has been granted access to the additional info area")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to access the additional info")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet(location1+"/additional-info/myarea", token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"air3":"something"}`, response.body)
}

func TestGetAdditionalInfo_UserSelfAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "air3a-")
	created := tstPerformPost(location1+"/additional-info/selfread", `{"air3a":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to access an additional info field with self read permissions")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet(location1+"/additional-info/selfread", token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"air3a":"something"}`, response.body)
}

func TestGetAdditionalInfo_UserOtherDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "air3b-")
	created := tstPerformPost(location1+"/additional-info/selfread", `{"air3b":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given another attendee with a different identity and no special permissions")
	token := tstValidUserToken(t, 101)
	_, _ = tstRegisterAttendeeWithToken(t, "air3b1-", token)

	docs.When("when they attempt to access a self readable additional info field of another attendee")
	response := tstPerformGet(location1+"/additional-info/selfread", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")
}

func TestGetAdditionalInfo_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "air4-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"air4":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when an admin attempts to access the additional info")
	token := tstValidAdminToken(t)
	response := tstPerformGet(location1+"/additional-info/myarea", token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"air4":"something"}`, response.body)
}

func TestGetAdditionalInfo_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to access additional info but supply an invalid badge number")
	response := tstPerformGet("/api/rest/v1/attendees/kitty/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestGetAdditionalInfo_IdDoesNotExist(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to access additional info but supply a badge number that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42342/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

func TestGetAdditionalInfo_InvalidArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "air7-")

	docs.When("when an admin attempts to access additional info but supplies an invalid area")
	token := tstValidAdminToken(t)
	response := tstPerformGet(location1+"/additional-info/area-cannot-contain-dashes", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.invalid", url.Values{"area": []string{"must match [a-z]+"}})
}

func TestGetAdditionalInfo_NotConfiguredArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "air8-")

	docs.When("when an admin attempts to access additional info but asks for an area that is not listed in the configuration")
	token := tstValidAdminToken(t)
	response := tstPerformGet(location1+"/additional-info/unlisted", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.unlisted", url.Values{"area": []string{"areas must be enabled in configuration"}})
}

func TestGetAdditionalInfo_Unset(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "air9-")

	docs.When("when an admin attempts to access additional info using an area that is not assigned")
	token := tstValidAdminToken(t)
	response := tstPerformGet(location1+"/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "addinfo.notfound.error", url.Values{})
}

// --- read access for global (id 0) ---

func TestGetAdditionalInfoGlobal_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"airg1":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the global additional info")
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestGetAdditionalInfoGlobal_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"airg2":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee")
	_, att1 := tstRegisterAttendee(t, "airg2-")

	docs.When("when they attempt to access the global additional info but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")
}

func TestGetAdditionalInfoGlobal_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"airg3":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee who has been granted access to the additional info area")
	location1, att1 := tstRegisterAttendee(t, "air3-")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to access the global additional info")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is successful and they can retrieve the global additional info")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"airg3":"something"}`, response.body)
}

func TestGetAdditionalInfoGlobal_SelfReadAllowsGlobalRead(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set for an area with self_read permissions")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/selfread", `{"airg4":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee with no specific access to the additional info area")
	_, att1 := tstRegisterAttendee(t, "airg4-")
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to access the global additional info value")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/selfread", token)

	docs.Then("then the request is successful and they can retrieve the global additional info")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"airg4":"something"}`, response.body)
}

func TestGetAdditionalInfoGlobal_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"airg5":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when an admin attempts to access the global additional info")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusOK, response.status)
	require.Equal(t, `{"airg5":"something"}`, response.body)
}

func TestGetAdditionalInfoGlobal_Unset(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts to access global additional info using an area that is not assigned")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "addinfo.notfound.error", url.Values{})
}

// --- write access

func TestWriteAdditionalInfo_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aiw1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to write additional info for the attendee")
	response := tstPerformPost(location1+"/additional-info/myarea", `{"aiw1":"something"}`, token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no additional info has been written")
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestWriteAdditionalInfo_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, att1 := tstRegisterAttendee(t, "aiw2-")

	docs.When("when they attempt to write even their own additional info but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost(location1+"/additional-info/myarea", `{"aiw2":"something"}`, token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and no additional info has been written")
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestWriteAdditionalInfo_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aiw3-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aiw3":"original-value"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given the attendee has been granted access to the additional info area")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to overwrite the additional info")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost(location1+"/additional-info/myarea", `{"aiw3":"new-value"}`, token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aiw3":"new-value"}`, readAgain.body)
}

func TestWriteAdditionalInfo_UserSelfWriteAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aiw3a-")
	created := tstPerformPost(location1+"/additional-info/selfwrite", `{"aiw3a":"original-value"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to overwrite an additional info that has self write permissions configured")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost(location1+"/additional-info/selfwrite", `{"aiw3a":"new-value"}`, token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/selfwrite", tstValidAdminToken(t))
	require.Equal(t, `{"aiw3a":"new-value"}`, readAgain.body)
}

func TestWriteAdditionalInfo_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aiw4-")

	docs.When("when an admin attempts to write to a configured additional info area")
	token := tstValidAdminToken(t)
	response := tstPerformPost(location1+"/additional-info/myarea", `{"aiw4":"something"}`, token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aiw4":"something"}`, readAgain.body)
}

func TestWriteAdditionalInfo_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to write additional info but supply an invalid badge number")
	response := tstPerformPost("/api/rest/v1/attendees/d0gg13/additional-info/myarea", `{"aiw5":"something"}`, token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestWriteAdditionalInfo_IdDoesNotExist(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to write additional info but supply a badge number that does not exist")
	response := tstPerformPost("/api/rest/v1/attendees/42342/additional-info/myarea", `{"aiw6":"something"}`, token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

func TestWriteAdditionalInfo_InvalidArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aiw7-")

	docs.When("when an admin attempts to write additional info but supplies an invalid area")
	token := tstValidAdminToken(t)
	response := tstPerformPost(location1+"/additional-info/1nv4lid", `{"aiw7":"something"}`, token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.invalid", url.Values{"area": []string{"must match [a-z]+"}})
}

func TestWriteAdditionalInfo_NotConfiguredArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aiw8-")

	docs.When("when an admin attempts to write additional info for an area that is not listed in the configuration")
	token := tstValidAdminToken(t)
	response := tstPerformPost(location1+"/additional-info/unlisted", `{"aiw8":"something"}`, token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.unlisted", url.Values{"area": []string{"areas must be enabled in configuration"}})
}

// --- write access for global (id 0)

func TestWriteAdditionalInfoGlobal_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to write global additional info")
	response := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aiwg1":"something"}`, token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no additional info has been written")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestWriteAdditionalInfoGlobal_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	_, att1 := tstRegisterAttendee(t, "aiw2-")

	docs.When("when they attempt to write global additional info but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aiwg2":"something"}`, token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and no additional info has been written")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestWriteAdditionalInfoGlobal_UserWithPermissionAllowOverwrite(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"airwg3":"original-value"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee who has been granted access to the additional info area")
	location1, att1 := tstRegisterAttendee(t, "aiwg3-")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to overwrite the additional info")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aiwg3":"new-value"}`, token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aiwg3":"new-value"}`, readAgain.body)
}

func TestWriteAdditionalInfoGlobal_UserSelfWriteFails(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set for an area with self_write permissions configured")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/selfwrite", `{"aiwg4":"original-value"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee with no specific area permissions")
	_, att1 := tstRegisterAttendee(t, "aiwg4-")

	docs.When("when they attempt to overwrite the global additional info that has self write permissions configured")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformPost("/api/rest/v1/attendees/0/additional-info/selfwrite", `{"aiwg4":"new-value"}`, token)

	docs.Then("then the request is denied with the expected error despite the self_write permissions")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and the additional info value is unchanged")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/selfwrite", tstValidAdminToken(t))
	require.Equal(t, `{"aiwg4":"original-value"}`, readAgain.body)
}

func TestWriteAdditionalInfoGlobal_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts to write a global value to a configured additional info area")
	token := tstValidAdminToken(t)
	response := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aiwg5":"something"}`, token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aiwg5":"something"}`, readAgain.body)
}

// --- deletion

func TestDeleteAdditionalInfo_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "aid1-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aid1":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to delete additional info for the attendee")
	response := tstPerformDelete(location1+"/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and the additional info is untouched")
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aid1":"something"}`, readAgain.body)
}

func TestDeleteAdditionalInfo_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aid2-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aid2":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to delete even their own additional info but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete(location1+"/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and the additional info is untouched")
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aid2":"something"}`, readAgain.body)
}

func TestDeleteAdditionalInfo_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aid3-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aid3":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given the attendee has been granted access to the additional info area")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to delete the additional info entry")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete(location1+"/additional-info/myarea", token)

	docs.Then("then the request is successful and the entry has been deleted")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestDeleteAdditionalInfo_UserSelfWriteAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aid3a-")
	created := tstPerformPost(location1+"/additional-info/selfwrite", `{"aid3a":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to delete the additional info entry for a field that has self write permission configured")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete(location1+"/additional-info/selfwrite", token)

	docs.Then("then the request is successful and the entry has been deleted")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/selfwrite", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestDeleteAdditionalInfo_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "aid4-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aid4":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when an admin attempts to delete the additional info")
	token := tstValidAdminToken(t)
	response := tstPerformDelete(location1+"/additional-info/myarea", token)

	docs.Then("then the request is successful and the entry has been deleted")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet(location1+"/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestDeleteAdditionalInfo_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to delete additional info but supply an invalid badge number")
	response := tstPerformDelete("/api/rest/v1/attendees/fluffy/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestDeleteAdditionalInfo_IdDoesNotExist(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to access additional info but supply a badge number that does not exist")
	response := tstPerformDelete("/api/rest/v1/attendees/42342/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

func TestDeleteAdditionalInfo_InvalidArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aid7-")

	docs.When("when an admin attempts to delete additional info but supplies an invalid area")
	token := tstValidAdminToken(t)
	response := tstPerformDelete(location1+"/additional-info/area-cannot-contain-dashes", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.invalid", url.Values{"area": []string{"must match [a-z]+"}})
}

func TestDeleteAdditionalInfo_NotConfiguredArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aid8-")

	docs.When("when an admin attempts to delete additional info for an area that is not listed in the configuration")
	token := tstValidAdminToken(t)
	response := tstPerformDelete(location1+"/additional-info/unlisted", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.unlisted", url.Values{"area": []string{"areas must be enabled in configuration"}})
}

func TestDeleteAdditionalInfo_Unset(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "aid9-")

	docs.When("when an admin attempts to delete additional info using an area that is not assigned")
	token := tstValidAdminToken(t)
	response := tstPerformDelete(location1+"/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "addinfo.notfound.error", url.Values{})
}

// --- deletion for global (id 0)

func TestDeleteAdditionalInfoGlobal_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aidg1":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to delete the global additional info value")
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and the additional info is untouched")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aidg1":"something"}`, readAgain.body)
}

func TestDeleteAdditionalInfoGlobal_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aidg2":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee who does not have specific access to the additional info area")
	_, att1 := tstRegisterAttendee(t, "aidg2-")

	docs.When("when they attempt to delete the global additional info value")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and the additional info is untouched")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, `{"aidg2":"something"}`, readAgain.body)
}

func TestDeleteAdditionalInfoGlobal_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aidg3":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee who has been granted access to the additional info area")
	location1, att1 := tstRegisterAttendee(t, "aidg3-")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to delete the additional info entry")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is successful and the entry has been deleted")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestDeleteAdditionalInfoGlobal_UserSelfWriteAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set for an area with the self_write permission")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/selfwrite", `{"aidg4":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an existing attendee with no specific permissions for the area")
	_, att1 := tstRegisterAttendee(t, "aidg4-")

	docs.When("when they attempt to delete the additional info entry")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/selfwrite", token)

	docs.Then("then the request is denied with the expected error despite the self_write permissions")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")

	docs.Then("and the additional info value is unchanged")
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/selfwrite", tstValidAdminToken(t))
	require.Equal(t, `{"aidg4":"something"}`, readAgain.body)
}

func TestDeleteAdditionalInfoGlobal_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a global additional info field is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aidg5":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when an admin attempts to delete the additional info")
	token := tstValidAdminToken(t)
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request is successful and the entry has been deleted")
	require.Equal(t, http.StatusNoContent, response.status)
	readAgain := tstPerformGet("/api/rest/v1/attendees/0/additional-info/myarea", tstValidAdminToken(t))
	require.Equal(t, http.StatusNotFound, readAgain.status)
}

func TestDeleteAdditionalInfoGlobal_Unset(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts to delete global additional info using an area that is not assigned a value")
	token := tstValidAdminToken(t)
	response := tstPerformDelete("/api/rest/v1/attendees/0/additional-info/myarea", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "addinfo.notfound.error", url.Values{})
}

// getAllAdditionalInfo

func TestGetAllAdditionalInfo_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "aia1-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aia1":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to read the additional info area for all attendees")
	response := tstPerformGet("/api/rest/v1/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestGetAllAdditionalInfo_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aia2-")
	created := tstPerformPost(location1+"/additional-info/myarea", `{"aia2":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to access the additional info area for all attendees but do not have access")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/additional-info/myarea", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")
}

func TestGetAllAdditionalInfo_UserWithPermissionAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given three existing attendees, two of which have an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aia3a-")
	created1 := tstPerformPost(location1+"/additional-info/myarea", `{"aia3a":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created1.status)

	location2, _ := tstRegisterAttendeeWithToken(t, "aia3b-", tstValidUserToken(t, 101))
	created2 := tstPerformPost(location2+"/additional-info/myarea", `{"aia3b":"something else"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created2.status)

	_, _ = tstRegisterAttendeeWithToken(t, "aia3c-", tstValidUserToken(t, 102))

	docs.Given("given the first attendee has been granted access to the additional info area")
	body := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	accessGranted := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, accessGranted.status)

	docs.When("when they attempt to access the additional info area for all attendees")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/additional-info/myarea", token)

	docs.Then("then the request is successful and they can retrieve the additional info again")
	expectedValues := map[string]string{
		"1": "{\"aia3a\":\"something\"}",
		"4": "{\"aia3b\":\"something else\"}", // ids in-memory are a global sequence
	}
	expected := addinfo.AdditionalInfoFullArea{
		Area:   "myarea",
		Values: expectedValues,
	}
	actual := addinfo.AdditionalInfoFullArea{}
	tstRequireSuccessResponse(t, response, http.StatusOK, &actual)
	require.Equal(t, expected.Area, actual.Area)
	require.EqualValues(t, expected.Values, actual.Values)
}

func TestGetAllAdditionalInfo_UserSelfDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, att1 := tstRegisterAttendee(t, "aia3a-")
	created := tstPerformPost(location1+"/additional-info/selfread", `{"aia3a":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when they attempt to access the additional info area with self read permissions")
	token := tstValidUserToken(t, att1.Id)
	response := tstPerformGet("/api/rest/v1/additional-info/selfread", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this additional info area - the attempt has been logged")
}

func TestGetAllAdditionalInfo_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee with an additional info field set")
	location1, _ := tstRegisterAttendee(t, "aia4-")
	created2 := tstPerformPost(location1+"/additional-info/myarea", `{"aia4":"something"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created2.status)

	docs.Given("given a global additional info value is set")
	created := tstPerformPost("/api/rest/v1/attendees/0/additional-info/myarea", `{"aia4":"meow"}`, tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, created.status)

	docs.When("when an admin attempts to access the additional info")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/additional-info/myarea", token)

	docs.Then("then the request is successful and the response is as expected")
	expectedValues := map[string]string{
		"0": "{\"aia4\":\"meow\"}",
		"1": "{\"aia4\":\"something\"}",
	}
	expected := addinfo.AdditionalInfoFullArea{
		Area:   "myarea",
		Values: expectedValues,
	}
	actual := addinfo.AdditionalInfoFullArea{}
	tstRequireSuccessResponse(t, response, http.StatusOK, &actual)
	require.Equal(t, expected.Area, actual.Area)
	require.EqualValues(t, expected.Values, actual.Values)
}

func TestGetAllAdditionalInfo_InvalidArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts to access additional info but supplies an invalid area")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/additional-info/area-cannot-contain-dashes", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.invalid", url.Values{"area": []string{"must match [a-z]+"}})
}

func TestGetAllAdditionalInfo_NotConfiguredArea(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts to access all additional info but asks for an area that is not listed in the configuration")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/additional-info/unlisted", token)

	docs.Then("then the request fails and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "addinfo.area.unlisted", url.Values{"area": []string{"areas must be enabled in configuration"}})
}

func TestGetAllAdditionalInfo_Unset(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	_, _ = tstRegisterAttendee(t, "aia7-")

	docs.When("when an admin reads all additional info for a valid area that has no values assigned")
	token := tstValidAdminToken(t)
	response := tstPerformGet("/api/rest/v1/additional-info/myarea", token)

	docs.Then("then the request is successful with an appropriate response with an empty values object, which is not missing")
	expected := addinfo.AdditionalInfoFullArea{
		Area:   "myarea",
		Values: map[string]string{},
	}
	actual := addinfo.AdditionalInfoFullArea{}
	tstRequireSuccessResponse(t, response, http.StatusOK, &actual)
	require.Equal(t, expected.Area, actual.Area)
	require.EqualValues(t, expected.Values, actual.Values)
}
