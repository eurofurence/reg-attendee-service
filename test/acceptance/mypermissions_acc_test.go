package acceptance

import (
	"net/http"
	"testing"

	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------
// acceptance tests for GET /attendees/my-permissions
// -------------------------------------------------------

// --- auth ---

func TestMyPermissions_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they request their permissions")
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401)")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

// --- no registration ---

// valid_JWT_is_not_staff_sub1234567890 has no groups claim in the JWT, sub=1234567890
func TestMyPermissions_LoggedInNoRegistration(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user with no registration (sub=1234567890, no groups in JWT)")
	token := valid_JWT_is_not_staff_sub1234567890

	docs.When("when they request their permissions")
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then the request is successful and both lists are empty")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{}, dto.Groups)
	require.EqualValues(t, []string{}, dto.Permissions)
}

// --- groups from JWT ---

// valid_JWT_is_not_staff_sub101 has groups=["somegroup"] in the JWT, sub=101
func TestMyPermissions_GroupsFromJwt(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user whose JWT contains a group claim (sub=101, groups=[somegroup])")
	token := valid_JWT_is_not_staff_sub101

	docs.When("when they request their permissions with no registration")
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then the response contains the OIDC group from the JWT and an empty permissions list")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{"somegroup"}, dto.Groups)
	require.EqualValues(t, []string{}, dto.Permissions)
}

// valid_JWT_is_admin_sub1234567890 has groups=["admin"] in the JWT
func TestMyPermissions_AdminGroupFromJwt(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin with no registration (JWT groups=[admin])")
	token := tstValidAdminToken(t)

	docs.When("when they request their permissions")
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then the response contains the admin OIDC group and an empty permissions list")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{"admin"}, dto.Groups)
	require.EqualValues(t, []string{}, dto.Permissions)
}

// --- permissions from registration admin info ---

// valid_JWT_is_not_staff_sub1234567890 has no groups in JWT; tstRegisterAttendee uses staff token (sub=1234567890)
func TestMyPermissions_PermissionsFromAdminInfo(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee registered with sub=1234567890")
	location1, _ := tstRegisterAttendee(t, "myp1-")

	docs.Given("given an admin grants them regdesk and sponsordesk permissions")
	body := admin.AdminInfoDto{
		Permissions: "regdesk,sponsordesk",
	}
	adminResponse := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, adminResponse.status)

	docs.When("when the attendee requests their permissions (no groups in their JWT)")
	token := valid_JWT_is_not_staff_sub1234567890
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then the response contains both permissions sorted alphabetically and no groups")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{}, dto.Groups)
	require.EqualValues(t, []string{"regdesk", "sponsordesk"}, dto.Permissions)
}

// valid_JWT_is_staff_sub1234567890 has groups=["staff"] in JWT
func TestMyPermissions_PermissionsAndGroupsCombined(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee registered with sub=1234567890")
	location1, _ := tstRegisterAttendee(t, "myp2-")

	docs.Given("given an admin grants them regdesk permission")
	body := admin.AdminInfoDto{
		Permissions: "regdesk",
	}
	adminResponse := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, adminResponse.status)

	docs.When("when they request their permissions using a staff JWT (same sub=1234567890, groups=[staff])")
	token := valid_JWT_is_staff_sub1234567890
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then the response contains the staff group and the regdesk permission")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{"staff"}, dto.Groups)
	require.EqualValues(t, []string{"regdesk"}, dto.Permissions)
}

// Permissions not in the allowed list (from config) must not be returned
func TestMyPermissions_UnknownPermissionsNotReturned(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee registered with sub=1234567890")
	location1, _ := tstRegisterAttendee(t, "myp3-")

	docs.Given("given an admin sets a permission not present in the allowed-permissions configuration")
	body := admin.AdminInfoDto{
		Permissions: "regdesk,notaknownpermission",
	}
	adminResponse := tstPerformPut(location1+"/admin", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, adminResponse.status)

	docs.When("when they request their permissions")
	token := valid_JWT_is_not_staff_sub1234567890
	response := tstPerformGet("/api/rest/v1/attendees/my-permissions", token)

	docs.Then("then only the known permission appears in the response, the unknown one is filtered out")
	require.Equal(t, http.StatusOK, response.status)
	dto := attendee.UserPermissionsDto{}
	tstParseJson(response.body, &dto)
	require.EqualValues(t, []string{}, dto.Groups)
	require.EqualValues(t, []string{"regdesk"}, dto.Permissions)
}
