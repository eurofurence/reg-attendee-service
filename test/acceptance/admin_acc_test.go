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
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	docs.When("when they attempt to access the admin information")
	response := tstPerformGet(creationResponse.location+"/admin", token)

	docs.Then("then the request is denied as unauthenticated (401) and no body is returned")
	require.Equal(t, http.StatusUnauthorized, response.status, "unexpected http response status")
	require.Equal(t, "", response.body)
}

// helper functions

func tstReadAdminInfo(t *testing.T, location string, bearerToken string) admin.AdminInfoDto {
	response := tstPerformGet(location, bearerToken)
	require.Equal(t, http.StatusOK, response.status)

	adminInfo := admin.AdminInfoDto{}
	tstParseJson(response.body, &adminInfo)
	return adminInfo
}
