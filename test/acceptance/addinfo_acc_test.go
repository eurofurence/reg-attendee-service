package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
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
