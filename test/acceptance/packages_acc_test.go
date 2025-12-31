package acceptance

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/counts"
	"github.com/stretchr/testify/require"
)

// --- getPackageLimit ---

func TestPackageLimitDenyWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a user who is not logged in")

	docs.When("when they attempt to read package limits while not logged in")
	response := tstPerformGet("/api/rest/v1/packages/mountain-trip/limit", tstNoToken())

	docs.Then("then the request is denied")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestPackageLimitAllowWhileLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a user who is logged in but does not have a valid registration")
	token := tstValidUserToken(t, 101)

	docs.When("when they read package limits")
	readResponse := tstPerformGet("/api/rest/v1/packages/mountain-trip/limit", token)

	docs.Then("then the response is as expected")
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	expected := counts.PackageCount{
		Pending:   0,
		Attending: 0,
		Limit:     4,
	}
	actual := counts.PackageCount{}
	tstParseJson(readResponse.body, &actual)
	require.EqualValues(t, expected, actual, "unexpected counts in response")
}

func TestPackageLimitUnlimited(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a user who is logged in")
	token := tstValidUserToken(t, 101)

	docs.When("when they try to read package limits for an unlimited package")
	response := tstPerformGet("/api/rest/v1/packages/attendance/limit", token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "package.param.unlimited", "this package is unlimited, we do not track allocations for unlimited packages")
}

func TestPackageLimitNotFound(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a user who is logged in")
	token := tstValidUserToken(t, 101)

	docs.When("when they try to read package limits for a package that does not exist")
	response := tstPerformGet("/api/rest/v1/packages/kitten/limit", token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "package.param.notfound", "")
}

// other tests are baked into various reg and status cases - find usages on this function to find them

func tstRequirePackageCount(t *testing.T, pkg string, expected counts.PackageCount) {
	t.Helper()

	response := tstPerformGet(fmt.Sprintf("/api/rest/v1/packages/%s/limit", pkg), tstValidUserToken(t, 101))
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")

	actual := counts.PackageCount{}
	tstParseJson(response.body, &actual)
	require.EqualValues(t, expected, actual, "unexpected counts in response")
}
