package acceptance

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/counts"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
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

func TestPackageRecalc(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	_, _ = tstPkgStatRegisterAndProgressWithPackages(t, "pkgrec1", status.Approved, "mountain-trip,mountain-trip,mountain-trip", 202)

	docs.Given("the sales limit cache has become outdated")
	// simulated by database manipulation
	_, err := database.GetRepository().AddCount(context.TODO(), &entity.Count{
		Area:      entity.CountAreaPackage,
		Name:      "mountain-trip",
		Pending:   0,
		Attending: -1,
	})
	require.NoError(t, err)

	docs.When("when the sales limit cache is refreshed")
	response := tstPerformPost("/api/rest/v1/packages/mountain-trip/limit", "", tstValidAdminToken(t))

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")

	docs.Then("and the sales limit cache is correct again")
	tstRequirePackageCount(t, "mountain-trip", counts.PackageCount{
		Attending: 3,
		Limit:     4,
	})
}

func TestPackageLimitStatusTransitions(t *testing.T) {
	testcases := []struct {
		name            string
		description     string
		count1          int
		count2          int
		oldStatus       status.Status
		newStatus       status.Status
		expectError     string
		expectPending   int
		expectAttending int
	}{
		// always succeed because no effect on total limit, or reduction
		{
			name:          "pkgst1",
			description:   "successful transition new -> waiting below limit",
			count1:        1,
			oldStatus:     status.New,
			newStatus:     status.Waiting,
			expectPending: 1,
		},
		{
			name:            "pkgst2",
			description:     "successful transition new -> approved below limit",
			count1:          3,
			oldStatus:       status.New,
			newStatus:       status.Approved,
			expectAttending: 3,
		},
		{
			name:            "pkgst3",
			description:     "successful reactivation waiting -> approved below limit",
			count1:          3,
			oldStatus:       status.Waiting,
			newStatus:       status.Approved,
			expectAttending: 3,
		},
		{
			name:        "pkgst4",
			description: "cancellation frees up the limit again",
			count1:      1,
			oldStatus:   status.New,
			newStatus:   status.Cancelled,
		},
		{
			name:        "pkgst5",
			description: "cancellation from attending frees up the limit again",
			count1:      1,
			oldStatus:   status.Approved,
			newStatus:   status.Cancelled,
		},
		{
			name:            "pkgst6",
			description:     "reactivation cancelled -> approved fails due to overrun",
			count1:          3,
			count2:          3,
			oldStatus:       status.Cancelled,
			newStatus:       status.Approved,
			expectError:     "status.package.overrun",
			expectAttending: 3,
		},
		{
			name:            "pkgst7",
			description:     "reactivation cancelled -> new fails due to overrun",
			count1:          3,
			count2:          3,
			oldStatus:       status.Cancelled,
			newStatus:       status.New,
			expectError:     "status.package.overrun",
			expectAttending: 3,
		},
		{
			name:            "pkgst8",
			description:     "reactivation cancelled -> waiting succeeds at the limit",
			count1:          3,
			count2:          1,
			oldStatus:       status.Cancelled,
			newStatus:       status.Waiting,
			expectPending:   3,
			expectAttending: 1,
		},
	}

	pkgFromCount := func(count int) string {
		return strings.Join([]string{"mountain-trip", "mountain-trip", "mountain-trip", "mountain-trip"}[0:count], ",")
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tstSetup(false, false, true)
			defer tstShutdown()

			docs.Description(tc.description)

			docs.Given("given an attendee in status " + string(tc.oldStatus))
			loc, _ := tstPkgStatRegisterAndProgressWithPackages(t, tc.name, tc.oldStatus, pkgFromCount(tc.count1), 1)

			if tc.count2 > 0 {
				docs.Given("given a second approved attendee")
				_, _ = tstPkgStatRegisterAndProgressWithPackages(t, tc.name+"b", status.Approved, pkgFromCount(tc.count2), 202)
			}

			docs.When("when an admin changes the status of the first attendee to " + string(tc.newStatus))
			body := status.StatusChangeDto{
				Status:  tc.newStatus,
				Comment: tc.name,
			}
			response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

			if tc.expectError != "" {
				docs.Then("then the request fails with the expected error")
				tstRequireErrorResponse(t, response, http.StatusConflict, tc.expectError, "")
			} else {
				docs.Then("then the request succeeds")
				require.Equal(t, http.StatusNoContent, response.status)
			}

			docs.Then("and the package counts are as expected")
			tstRequirePackageCount(t, "mountain-trip", counts.PackageCount{
				Pending:   tc.expectPending,
				Attending: tc.expectAttending,
				Limit:     4,
			})
		})
	}
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

func tstPkgStatRegisterAndProgressWithPackages(t *testing.T, testcase string, targetStatus status.Status, addPackages string, id uint) (string, attendee.AttendeeDto) {
	// id only makes a difference if set to 202
	token := tstValidStaffToken(t, id)

	dto := tstBuildValidAttendee(testcase)
	dto.Packages = dto.Packages + "," + addPackages
	dto.PackagesList = tstPackagesListFromPackages(dto.Packages)
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(dto), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	rereadResponse := tstPerformGet(creationResponse.location, token)
	require.Equal(t, http.StatusOK, rereadResponse.status, "unexpected http response status")
	tstParseJson(rereadResponse.body, &dto)

	if targetStatus == status.New {
		return creationResponse.location, dto
	}
	// a single status change is all we need here
	body := status.StatusChangeDto{
		Status:  targetStatus,
		Comment: testcase,
	}
	response := tstPerformPost(creationResponse.location+"/status", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, response.status)

	return creationResponse.location, dto
}
