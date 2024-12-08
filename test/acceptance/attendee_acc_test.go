package acceptance

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------
// acceptance tests for the attendee resource
// ------------------------------------------

// --- create new attendee ---

// -- validation and duplicate handling tests --

func TestCreateNewAttendeeInvalid(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee with invalid data")
	attendeeSent := tstBuildValidAttendee("nav1-")
	attendeeSent.Nickname = "$%&^@!$"
	tstAddPackages(&attendeeSent, "sponsor") // a constraint violation
	attendeeSent.Birthday = "2004-11-23"     // too young
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"birthday": []string{"birthday must be no earlier than 1901-01-01 and no later than 2001-08-14"},
		"nickname": []string{"nickname field must contain at least one alphanumeric character", "nickname field must not contain more than two non-alphanumeric characters (not counting spaces)"},
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
	})
}

func TestCreateNewAttendeeInvalidClassicPackages(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee with invalid data using a packages list")
	attendeeSent := tstBuildValidAttendee("nav1b-")
	attendeeSent.Packages = attendeeSent.Packages + ",sponsor" // constraint violation
	attendeeSent.PackagesList = nil                            // make packages field actually count
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response, and the packages field was used")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
	})
}

func TestCreateNewAttendeeInvalidPackagesList(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee with invalid data using a packages list with omitted count field")
	attendeeSent := tstBuildValidAttendee("nav1c-")
	attendeeSent.Packages = ""                                                                            // should be ignored, so let's produce a different error if used
	attendeeSent.PackagesList = append(attendeeSent.PackagesList, attendee.PackageState{Name: "sponsor"}) // constraint violation, also tests that Count: 0 means Count: 1
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response, the packages field was ignored, and count was interpreted correctly")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
	})
}

func TestCreateNewAttendeeInvalid_NoMandatoryPackage(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee, but pick none of the at-least-one-mandatory packages")
	attendeeSent := tstBuildValidAttendee("nav1d-")
	tstOverridePackages(&attendeeSent, "room-none,sponsor")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"you must pick at least one of the mandatory options (attendance,day-fri,day-sat,day-thu)"},
	})
}

func TestCreateNewAttendeeSyntaxInvalid(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they try to create a new attendee with syntactically invalid data")
	attendeeSent := tstBuildValidAttendee("nav2-")
	syntaxErrorJson := "{" + tstRenderJson(attendeeSent)
	response := tstPerformPost("/api/rest/v1/attendees", syntaxErrorJson, tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.parse.error", url.Values{})
}

func TestCreateNewAttendeeAdminOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard public registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they send a new attendee and attempt to set an admin only flag (guest)")
	attendeeSent := tstBuildValidAttendee("nav3-")
	attendeeSent.Flags = "guest,hc,terms-accepted"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"flags field must be a comma separated combination of any of anon,ev,hc,terms-accepted"},
	})
}

func TestCreateNewAttendeeReadOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard public registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they send a new attendee and attempt to set a read only flag (ev)")
	attendeeSent := tstBuildValidAttendee("nav4-")
	attendeeSent.Flags = "ev,anon,terms-accepted"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"forbidden select or deselect of flag ev - only an admin can do that"},
	})
}

func TestCreateNewAttendeeReadOnlyFlagRemove(t *testing.T) {
	docs.Given("given the configuration for standard public registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they send a new attendee and attempt to leave out a read only flag (terms-accepted)")
	attendeeSent := tstBuildValidAttendee("nav4-")
	attendeeSent.Flags = "anon"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"forbidden select or deselect of flag terms-accepted - only an admin can do that"},
	})
}

func TestCreateNewAttendeeAdminOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard public registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send a new attendee and attempt to set an admin only flag (guest)")
	attendeeSent := tstBuildValidAttendee("nav5-")
	attendeeSent.Flags = "guest"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is rejected with an error response, because admin only flags belong in adminInfo")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"flags field must be a comma separated combination of any of anon,ev,hc,terms-accepted"},
	})
}

func TestCreateNewAttendeeDefaultReadOnlyPackage(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration before public registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given a staffer")
	staffToken := tstValidStaffToken(t, 1)

	docs.When("when they send a new attendee and attempt to leave out a read only default package (room-none)")
	attendeeSent := tstBuildValidAttendee("nav6-")
	tstOverridePackages(&attendeeSent, "attendance,stage,sponsor")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), staffToken)

	docs.Then("then the attendee is rejected with an error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"forbidden select or deselect of package room-none - only an admin can do that"},
	})
}

func TestCreateNewAttendeeDuplicateHandling(t *testing.T) {
	docs.Given("given the configuration for standard public registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user and an existing registration")
	attendeeSent := tstBuildValidAttendee("na10-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	docs.When("when they attempt to create a new attendee with the same nickname, zip, email")
	duplicateAttendeeSent := tstBuildValidAttendee("na10-")
	duplicateAttendeeSent.Nickname = attendeeSent.Nickname
	duplicateAttendeeSent.Zip = attendeeSent.Zip
	duplicateAttendeeSent.Email = attendeeSent.Email
	duplicateResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(duplicateAttendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response indicating a duplicate")
	tstRequireErrorResponse(t, duplicateResponse, http.StatusConflict, "attendee.data.duplicate", url.Values{
		"attendee": []string{"there is already an attendee with this information (looking at nickname, email, and zip code)"},
	})
}

// -- no login required for new registrations --

// - before both target times -

func TestCreateNewAttendee_NoLoginRequired_TooEarly_Anon(t *testing.T) {
	docs.Given("given the configuration for public registration before any registration target time")
	tstSetup(false, false, false)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data before registration has begun")
	attendeeSent := tstBuildValidAttendee("na1-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attempt is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_NoLoginRequired_TooEarly_User(t *testing.T) {
	docs.Given("given the configuration for public registration before any registration target time")
	tstSetup(false, false, false)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 101)

	docs.When("when they attempt to create a new attendee with valid data before registration has begun")
	attendeeSent := tstBuildValidAttendee("na2-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_NoLoginRequired_TooEarly_Staff(t *testing.T) {
	docs.Given("given the configuration for staff registration before any registration target time")
	tstSetup(false, true, false)
	defer tstShutdown()

	docs.Given("given a logged in user who is staff")
	token := tstValidStaffToken(t, 202)

	docs.When("when they attempt to create a new attendee with valid data before even staff registration has begun")
	attendeeSent := tstBuildValidAttendee("na3-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"staff registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_NoLoginRequired_TooEarly_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for public registration before any registration target time")
	tstSetup(false, false, false)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data before registration has begun")
	attendeeSent := tstBuildValidAttendee("na4-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected with an appropriate error response, that is, admins are treated just like staff")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
}

// - between both target times -

func TestCreateNewAttendee_NoLoginRequired_Between_Anon(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration before public registration but after the staff start time")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na10-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the request is denied with the appropriate error response and no location header is supplied")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
	require.Equal(t, "", response.location, "non-empty location header in response")
}

func TestCreateNewAttendee_NoLoginRequired_Between_User(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration before public registration but after the staff start time")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an authenticated regular user")
	userToken := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na11-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), userToken)

	docs.Then("then the request is denied with the appropriate error response and no location header is supplied")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
	require.Equal(t, "", response.location, "non-empty location header in response")
}

func TestCreateNewAttendee_NoLoginRequired_Between_Staff(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration before public registration but after the staff start time")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given a staffer")
	staffToken := tstValidStaffToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na12-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), staffToken)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_NoLoginRequired_Between_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration before public registration but after the staff start time")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an admin")
	adminToken := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na13-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), adminToken)

	docs.Then("then the attendee is successfully created, that is an admin is treated just like staff")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

// - after public target time -

func TestCreateNewAttendee_NoLoginRequired_After_Anon(t *testing.T) {
	docs.Given("given the configuration for public standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee")
	attendeeSent := tstBuildValidAttendee("na20-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	docs.Then("and its data can be read again by an admin")
	attendeeReadAgain := tstReadAttendee(t, response.location)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match sent data")
}

func TestCreateNewAttendee_NoLoginRequired_After_User(t *testing.T) {
	docs.Given("given the configuration for public registration after normal reg is open")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 101)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na21-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	docs.Then("and its data can be read again by the same user")
	readAgainResponse := tstPerformGet(response.location, token)
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match sent data")
}

func TestCreateNewAttendee_NoLoginRequired_After_Staff(t *testing.T) {
	docs.Given("given the configuration for public registration after normal reg is open")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, 202)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na22-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_NoLoginRequired_After_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for public registration after normal reg is open")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na23-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created, that is, admins are treated just like staff")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

// -- login required for all new registrations --

// we only test the timing-related cases, validation is no different except for the email validation against the token

func TestCreateNewAttendee_LoginRequired_EmailWrong(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with a different email address")
	attendeeSent := tstBuildValidAttendee("na24-")
	attendeeSent.Email = "nobody@example.com"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as invalid (400) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"email": []string{"you can only use the email address you're logged in with"},
	})
}

func TestCreateNewAttendee_LoginRequired_EmailUnverified(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user who has not verified their email address")
	token := valid_JWT_is_not_staff_sub101_unverified

	docs.When("when they attempt to create a new attendee")
	attendeeSent := tstBuildValidAttendee("na25-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as invalid (400) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"email": []string{"you must verify your email address with the identity provider first"},
	})
}

// - before both target times -

func TestCreateNewAttendee_LoginRequired_TooEarly_Anon(t *testing.T) {
	docs.Given("given the configuration for login-only registration before any registration target time")
	tstSetup(true, false, false)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na30-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attempt is rejected as unauthenticated (401) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCreateNewAttendee_LoginRequired_TooEarly_User(t *testing.T) {
	docs.Given("given the configuration for login-only registration before any registration target time")
	tstSetup(true, false, false)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na31-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as too early with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_LoginRequired_TooEarly_Staff(t *testing.T) {
	docs.Given("given the configuration for login-only registration before any registration target time")
	tstSetup(true, true, false)
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na32-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"staff registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_LoginRequired_TooEarly_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for login-only registration before any registration target time")
	tstSetup(true, true, false)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na33-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected with an appropriate error response, that is, admins are treated just like staff")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"staff registration has not opened at this time, please come back later"},
	})
}

// - between both target times -

func TestCreateNewAttendee_LoginRequired_Between_Anon(t *testing.T) {
	docs.Given("given the configuration for login-only registration after staff reg is open but before normal reg is open")
	tstSetup(true, true, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na40-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attempt is rejected as unauthenticated (401) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCreateNewAttendee_LoginRequired_Between_User(t *testing.T) {
	docs.Given("given the configuration for login-only registration after staff reg is open but before normal reg is open")
	tstSetup(true, true, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na41-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as too early with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"timing": []string{"public registration has not opened at this time, please come back later"},
	})
}

func TestCreateNewAttendee_LoginRequired_Between_Staff(t *testing.T) {
	docs.Given("given the configuration for login-only registration after staff reg is open but before normal reg is open")
	tstSetup(true, true, true)
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data after staff registration has begun")
	attendeeSent := tstBuildValidAttendee("na42-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_LoginRequired_Between_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for login-only registration after staff reg is open but before normal reg is open")
	tstSetup(true, true, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data after staff registration has begun")
	attendeeSent := tstBuildValidAttendee("na43-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created, that is, admins are treated just like staff")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

// - after public target time -

func TestCreateNewAttendee_LoginRequired_After_Anon(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na50-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attempt is rejected as unauthenticated (401) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestCreateNewAttendee_LoginRequired_After_User(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na51-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_LoginRequired_After_Staff(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in staffer")
	token := tstValidStaffToken(t, 1)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na52-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_LoginRequired_After_AdminIsLikeStaff(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na53-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created, that is, admins are treated just like staff")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

// ... using different email

func TestCreateNewAttendee_LoginRequired_User_CannotUseDifferentEmail(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee with an email address they are not logged in with")
	attendeeSent := tstBuildValidAttendee("na60-")
	attendeeSent.Email = "na60-" + attendeeSent.Email
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"email": []string{"you can only use the email address you're logged in with"},
	})
}

func TestCreateNewAttendee_LoginRequired_Admin_MayUseDifferentEmail(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with an email address they are not logged in with")
	attendeeSent := tstBuildValidAttendee("na61-")
	attendeeSent.Email = "na61-" + attendeeSent.Email
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created, that is admins are allowed to use somebody else's email address")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_AutomaticGroupFlag(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user who has the 'ev' group")
	token := tstValidUserToken(t, 102)

	docs.When("when they create a new attendee")
	attendeeSent := tstBuildValidAttendee("na62-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	docs.Then("and it has been automatically assigned the 'ev' flag even though it was not provided")
	readAgainResponse := tstPerformGet(response.location, token)
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	// we expect the 'ev' flag added
	attendeeSent.Flags += ",ev"
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match expected data")
}

func TestCreateNewAttendee_AutomaticGroupFlag_CannotSet(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user who has the 'ev' group")
	token := tstValidUserToken(t, 102)

	docs.When("when they attempt to create a new attendee with the automatic 'ev' flag")
	attendeeSent := tstBuildValidAttendee("na63-")
	attendeeSent.Flags += ",ev"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as invalid (400) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"forbidden select or deselect of flag ev - only an admin can do that"},
	})
}

func TestCreateNewAttendee_ReadonlyDefaultPackageWithConstraintRemovable(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 101)

	docs.When("when they create a new attendee and remove a read-only default package with matching constraint (stage)")
	attendeeSent := tstBuildValidAttendee("na63-")
	tstOverridePackages(&attendeeSent, "room-none,day-sat,boat-trip")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendee_ReadonlyDefaultPackageNoConstraintNotRemovable(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 101)

	docs.When("when they create a new attendee and try to remove a read-only default package with no matching constraint (room-none)")
	attendeeSent := tstBuildValidAttendee("na65-")
	tstOverridePackages(&attendeeSent, "day-sat")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as invalid (400) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"forbidden select or deselect of package room-none - only an admin can do that"},
	})
}

func TestCreateNewAttendee_MultiPackage(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee, validly adding a package multiple times")
	attendeeSent := tstBuildValidAttendee("na66-")
	tstAddPackages(&attendeeSent, "mountain-trip,mountain-trip")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is successfully created and can be read again")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
	attendeeReadAgain := tstReadAttendee(t, response.location)
	attendeeSent.Id = attendeeReadAgain.Id // mask expected diff in id
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match updated data")
}

func TestCreateNewAttendee_MultiPackageTooMany(t *testing.T) {
	docs.Given("given the configuration for login-only registration after normal reg is open")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a logged in user")
	token := tstValidUserToken(t, 1)

	docs.When("when they attempt to create a new attendee, adding a package too many times")
	attendeeSent := tstBuildValidAttendee("na67-")
	tstAddPackages(&attendeeSent, "mountain-trip,mountain-trip,mountain-trip")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attempt is rejected as invalid (400) with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages":      []string{"package mountain-trip occurs too many times, can occur at most 2 times"},
		"packages_list": []string{"package mountain-trip occurs too many times, can occur at most 2 times"},
	})
}

// --- update attendee ---

func TestUpdateExistingAttendee_Self(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua1-", token)

	docs.When("when they send updated attendee info while logged in")
	changedAttendee := attendee1
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, location1, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}

func TestUpdateExistingAttendee_Self_ChangeEmail(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua1b-", token)

	docs.When("when they attempt to change the email address to one they are not logged in as")
	changedAttendee := attendee1
	changedAttendee.Email = "ua1b-" + changedAttendee.Email
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"email": []string{"you can only use the email address you're logged in with"},
	})
}

func TestUpdateExistingAttendee_Self_NoMandatoryPackages(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua1b-", token)

	docs.When("when they attempt to change the selected packages, and pick none of the at-least-one-mandatory packages")
	changedAttendee := attendee1
	tstOverridePackages(&changedAttendee, "room-none,sponsor")
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"you must pick at least one of the mandatory options (attendance,day-fri,day-sat,day-thu)"},
	})
}

func TestUpdateExistingAttendee_Other(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given two existing users, the second of which has registered")
	location2, attendee2 := tstRegisterAttendee(t, "ua2b-")
	token := tstValidUserToken(t, 101)

	docs.When("when the first user sends updated attendee info for the second user, i.e. for someone else")
	changedAttendee := attendee2
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	response := tstPerformPut(location2, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied as unauthorized (403) and the data remains unchanged")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestUpdateExistingAttendeeSyntaxInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "ua3-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update with syntactically invalid data")
	syntaxErrorJson := "{" + tstRenderJson(attendee1)
	response := tstPerformPut(location1, syntaxErrorJson, token)

	docs.Then("then the update is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.parse.error", "")
}

func TestUpdateExistingAttendeeDataInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	location1, attendee1 := tstRegisterAttendee(t, "ua4-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update the information with invalid data")
	changedAttendee := attendee1
	changedAttendee.Nickname = "$%&^@!$"        // not allowed
	tstAddPackages(&changedAttendee, "sponsor") // a constraint violation
	changedAttendee.Birthday = "2004-11-23"     // too young
	response := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"nickname": []string{"nickname field must contain at least one alphanumeric character", "nickname field must not contain more than two non-alphanumeric characters (not counting spaces)"},
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
		"birthday": []string{"birthday must be no earlier than 1901-01-01 and no later than 2001-08-14"},
	})
}

func TestUpdateExistingAttendeeClassicPackagesInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	location1, attendee1 := tstRegisterAttendee(t, "ua4b-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update the information with invalid data")
	changedAttendee := attendee1
	changedAttendee.Packages = changedAttendee.Packages + ",sponsor" // constraint violation
	changedAttendee.PackagesList = nil                               // make packages field actually count
	response := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response, and the packages field was used")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
	})
}

func TestUpdateExistingAttendeePackagesListInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	location1, attendee1 := tstRegisterAttendee(t, "ua4c-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update the information with invalid data")
	changedAttendee := attendee1
	changedAttendee.Packages = ""                                                                               // should be ignored, so let's produce a different error if used
	changedAttendee.PackagesList = append(changedAttendee.PackagesList, attendee.PackageState{Name: "sponsor"}) // constraint violation, also tests that Count: 0 means Count: 1
	response := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response, the packages field was ignored, and count was interpreted correctly")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"cannot pick both sponsor2 and sponsor - constraint violated"},
	})
}

func TestUpdateNonExistingAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update a non-existent attendee")
	nonExistingAttendee := tstBuildValidAttendee("ua5-")
	response := tstPerformPut("/api/rest/v1/attendees/42", tstRenderJson(nonExistingAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestUpdateAttendeeInvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update an attendee with an invalid id")
	nonExistingAttendee := tstBuildValidAttendee("ua6-")
	response := tstPerformPut("/api/rest/v1/attendees/helloworld", tstRenderJson(nonExistingAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestDenyUpdateExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee and a user who is not logged in")
	existingAttendee := tstBuildValidAttendee("ua7-")
	existingAttendee.FirstName = "Marianne"
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they send updated attendee info while not logged in")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the data remains unchanged")
	require.Equal(t, http.StatusUnauthorized, updateResponse.status, "unexpected http response status for insecure update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "Marianne", attendeeReadAgain.FirstName, "attendee data read did not match original data")
}

func TestDenyUpdateExistingOtherAttendeeWithStaffToken(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration (the other config doesn't even have a staff token)")
	tstSetup(true, true, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua8-")
	existingAttendee.FirstName = "Marianne"
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstValidAdminToken(t))
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when a logged in staffer, who is not that attendee, sends updated attendee info for them, i.e. for someone else")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	token := tstValidStaffToken(t, 202)
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied as unauthorized (403) and the data remains unchanged")
	require.Equal(t, http.StatusForbidden, updateResponse.status, "unexpected http response status for insecure update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "Marianne", attendeeReadAgain.FirstName, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeAdminOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	location1, attendee1 := tstRegisterAttendee(t, "ua9-")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they send updated attendee info and attempt to add an admin-only flag (guest)")
	changedAttendee := attendee1
	changedAttendee.Flags = changedAttendee.Flags + ",guest"
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied with an appropriate error")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"flags field must be a comma separated combination of any of anon,ev,hc,terms-accepted"},
	})

	docs.Then("and the data remains unchanged")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, "anon,hc,terms-accepted", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeAdminOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "ua10-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send updated attendee info and attempt to add an admin-only flag (guest)")
	changedAttendee := attendee1
	changedAttendee.Flags = changedAttendee.Flags + ",guest"
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied with an appropriate error, because admin only flags belong in adminInfo")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"flags field must be a comma separated combination of any of anon,ev,hc,terms-accepted"},
	})

	docs.Then("and the data remains unchanged")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, "anon,hc,terms-accepted", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeReadOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua11-", token)

	docs.When("when they send updated attendee info and attempt to add a read-only flag (ev)")
	changedAttendee := attendee1
	changedAttendee.Flags = changedAttendee.Flags + ",ev"
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied with an appropriate error, because only admins can change read only flags")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"forbidden select or deselect of flag ev - only an admin can do that"},
	})

	docs.Then("and the data remains unchanged")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, "anon,hc,terms-accepted", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeRemoveReadOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua11-", token)

	docs.When("when they send updated attendee info and attempt to remove a read-only flag (terms-accepted)")
	changedAttendee := attendee1
	changedAttendee.Flags = strings.ReplaceAll(changedAttendee.Flags, ",terms-accepted", "")
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied with an appropriate error, because only admins can change read only flags")
	tstRequireErrorResponse(t, updateResponse, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"flags": []string{"forbidden select or deselect of flag terms-accepted - only an admin can do that"},
	})

	docs.Then("and the data remains unchanged")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, "anon,hc,terms-accepted", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeReadOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "ua12-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send updated attendee info and add a read-only flag (ev)")
	changedAttendee := attendee1
	changedAttendee.Flags = changedAttendee.Flags + ",ev"
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, location1, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "anon,hc,terms-accepted,ev", attendeeReadAgain.Flags, "attendee data read did not match expected flags value")
}

func TestUpdateExistingAttendee_Admin_ChangeEmail(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ua1b-", token)

	docs.Given("given an admin")
	admin := tstValidAdminToken(t)

	docs.When("when they attempt to change the email address to one they are not logged in as")
	changedAttendee := attendee1
	changedAttendee.Email = "ua13-" + changedAttendee.Email
	updateResponse := tstPerformPut(location1, tstRenderJson(changedAttendee), admin)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, location1, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, location1)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}

func tstUpdateExistingAttendee_RemovePackage_Forbidden(t *testing.T, testcase string, who string, targetStatus status.Status, token string) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing %s in status %s", who, targetStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, targetStatus)

	docs.When("when they send updated attendee info and change their packages in a way that reduces total dues")
	changedAttendee := att
	tstOverridePackages(&changedAttendee, "room-none,attendance,stage") // removes sponsor2
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.Then("then the update is rejected with the corresponding error message")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages": []string{"deselect of package sponsor2 after payment leads to dues reduction - only an admin can do that at this time"},
	})
}

func tstUpdateExistingAttendee_RemovePackage_NoCostReduce_Allowed(t *testing.T, testcase string, who string, targetStatus status.Status, token string) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing %s in status %s", who, targetStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, targetStatus)
	// update to prepare test case
	changedAttendee := att
	tstAddPackages(&changedAttendee, "boat-trip")
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	require.Equal(t, loc, response.location, "location unexpectedly changed during update")

	docs.When("when they send updated attendee info and change packages in a way that does not reduce total dues")
	changedAttendee2 := att
	tstOverridePackages(&changedAttendee2, "attendance,mountain-trip,room-none,sponsor2,stage") // removes boat-trip, but adds mountain-trip which is more expensive
	response2 := tstPerformPut(loc, tstRenderJson(changedAttendee2), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, response2.status, "unexpected http response status for update")
	require.Equal(t, loc, response2.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	require.EqualValues(t, changedAttendee2, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "attendance,mountain-trip,room-none,sponsor2,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")
}

func tstUpdateExistingAttendee_RemovePackage_Allowed(t *testing.T, testcase string, who string, targetStatus status.Status, token string) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing %s in status %s", who, targetStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, targetStatus)

	docs.When("when they send updated attendee info and remove a package that has associated cost")
	changedAttendee := att
	tstOverridePackages(&changedAttendee, "attendance,room-none,stage") // removes sponsor2
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	require.Equal(t, loc, response.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "attendance,room-none,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")
}

func TestUpdateExistingAttendee_RemovePackageWithCostReduction_AfterPaid_UserForbidden(t *testing.T) {
	for i, targetStatus := range []status.Status{status.PartiallyPaid, status.Paid, status.CheckedIn} {
		testcase := fmt.Sprintf("ua2b-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		t.Run(string(targetStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_RemovePackage_Forbidden(t, testcase, "attendee", targetStatus, token)
		})
	}
}

func TestUpdateExistingAttendee_RemovePackageWithCost_BeforePaid_UserAllowed(t *testing.T) {
	for i, targetStatus := range []status.Status{status.New, status.Approved} {
		testcase := fmt.Sprintf("ua2c-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		t.Run(string(targetStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_RemovePackage_Allowed(t, testcase, "attendee", targetStatus, token)
		})
	}
}

func TestUpdateExistingAttendee_RemovePackageWithCost_Whenever_AdminOk(t *testing.T) {
	for i, targetStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua2d-%d", i+1)
		token := tstValidAdminToken(t)
		t.Run(string(targetStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_RemovePackage_Allowed(t, testcase, "admin", targetStatus, token)
		})
	}
}

func TestUpdateExistingAttendee_RemovePackageNoCostReduction_AnyTime_UserAllowed(t *testing.T) {
	for i, targetStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua2e-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		t.Run(string(targetStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_RemovePackage_NoCostReduce_Allowed(t, testcase, "attendee", targetStatus, token)
		})
	}
}

func tstUpdateExistingAttendee_AddOnePackage(t *testing.T, testcase string, origStatus status.Status, token string, expectedMails []mailservice.MailSendDto) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing attendee in status %s", origStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, origStatus)

	docs.When("when they send updated attendee info and add a package that has associated cost")
	changedAttendee := att
	changedAttendee.Packages = "room-none,attendance,stage,sponsor2,boat-trip" // adds boat-trip (and tests automatic reordering)
	// let's also test that the package List is populated from sending classic packages, so clear the list in this request
	changedAttendee.PackagesList = nil
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	require.Equal(t, loc, response.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	// test automatic reordering of packages to alphabetical order
	changedAttendee.Packages = "attendance,boat-trip,room-none,sponsor2,stage"
	// and test that package list is also sent in responses, also sorted alphabetically by names
	changedAttendee.PackagesList = tstPackagesListFromPackages(changedAttendee.Packages)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "attendance,boat-trip,room-none,sponsor2,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")

	docs.Then("and the expected mail messages have been sent")
	tstRequireMailRequests(t, expectedMails)
}

func TestUpdateExistingAttendee_AddOnePackage_AnyTime_UserAllowed(t *testing.T) {
	for i, origStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua3a-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		targetStatus := origStatus
		mails := []mailservice.MailSendDto{}
		if origStatus == status.Approved {
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 275.00, 275.00)
			mails = append(mails, mail)
		} else if origStatus == status.PartiallyPaid {
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 120.00, 275.00)
			mails = append(mails, mail)
		} else if origStatus == status.Paid {
			targetStatus = status.PartiallyPaid
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 20.00, 275.00)
			mails = append(mails, mail)
		}
		t.Run(string(origStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_AddOnePackage(t, testcase, origStatus, token, mails)
		})
	}
}

func tstUpdateExistingAttendee_AddTwoPackages(t *testing.T, testcase string, origStatus status.Status, token string, expectedMails []mailservice.MailSendDto) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing attendee in status %s who has been granted a due date extension", origStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, origStatus)
	tstUpdateCacheRelative(context.Background(), att.Id, 0, 0, "2023-04-20")

	docs.When("when they send updated attendee info and add a package that has associated cost")
	changedAttendee := att
	tstAddPackages(&changedAttendee, "boat-trip")
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.When("and then send updated attendee info and add a second package that has associated cost")
	changedAttendee2 := att
	tstAddPackages(&changedAttendee2, "boat-trip,mountain-trip")
	response2 := tstPerformPut(loc, tstRenderJson(changedAttendee2), token)

	docs.Then("then the attendee is successfully updated both times and the final data can be read again")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	require.Equal(t, http.StatusOK, response2.status, "unexpected http response status for update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	require.EqualValues(t, changedAttendee2, attendeeReadAgain, "attendee data read did not match final updated data")
	require.EqualValues(t, "attendance,boat-trip,mountain-trip,room-none,sponsor2,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")

	docs.Then("and the expected mail messages have been sent with the extended due date kept in place")
	for i := range expectedMails {
		expectedMails[i].Variables["due_date"] = "20.04.2023" // TODO format language dependent
	}
	tstRequireMailRequests(t, expectedMails)
}

func TestUpdateExistingAttendee_AddTwoPackages_AnyTime_UserAllowed(t *testing.T) {
	for i, origStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua3b-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		targetStatus := origStatus
		mails := []mailservice.MailSendDto{}
		if origStatus == status.Approved {
			mail1 := tstNewStatusMailWithAmounts(testcase, targetStatus, 275.00, 275.00)
			mail2 := tstNewStatusMailWithAmounts(testcase, targetStatus, 305.00, 305.00)
			mails = append(mails, mail1, mail2)
		} else if origStatus == status.PartiallyPaid {
			mail1 := tstNewStatusMailWithAmounts(testcase, targetStatus, 120.00, 275.00)
			mail2 := tstNewStatusMailWithAmounts(testcase, targetStatus, 150.00, 305.00)
			mails = append(mails, mail1, mail2)
		} else if origStatus == status.Paid {
			targetStatus = status.PartiallyPaid
			mail1 := tstNewStatusMailWithAmounts(testcase, targetStatus, 20.00, 275.00)
			mail2 := tstNewStatusMailWithAmounts(testcase, targetStatus, 50.00, 305.00)
			mails = append(mails, mail1, mail2)
		}
		t.Run(string(origStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_AddTwoPackages(t, testcase, origStatus, token, mails)
		})
	}
}

func tstUpdateExistingAttendee_AddOnePackage_Admin_SuppressEmailWorks(t *testing.T, testcase string, origStatus status.Status, token string, expectedMails []mailservice.MailSendDto) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing attendee in status %s", origStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, origStatus)

	docs.When("when an admin sends updated attendee info and adds a package that has associated cost (with suppressMinorUpdateEmail flag set)")
	changedAttendee := att
	tstAddPackages(&changedAttendee, "boat-trip")
	response := tstPerformPut(loc+"?suppressMinorUpdateEmail=yes", tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	require.Equal(t, loc, response.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "attendance,boat-trip,room-none,sponsor2,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")

	docs.Then("and no mail messages have been sent because of the suppressMinorUpdateEmail flag")
	tstRequireMailRequests(t, nil)
}

func TestUpdateExistingAttendee_AddOnePackage_AnyTime_AdminWithSuppressEmail(t *testing.T) {
	for i, origStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua3c-%d", i+1)
		token := tstValidAdminToken(t)
		noMails := []mailservice.MailSendDto{}
		t.Run(string(origStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_AddOnePackage_Admin_SuppressEmailWorks(t, testcase, origStatus, token, noMails)
		})
	}
}

func tstUpdateExistingAttendee_AddPackageMultipleTimes(t *testing.T, testcase string, origStatus status.Status, token string, expectedMails []mailservice.MailSendDto) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given(fmt.Sprintf("given an existing attendee in status %s", origStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, origStatus)

	docs.When("when they send updated attendee info and add a package multiple times")
	changedAttendee := att
	tstAddPackages(&changedAttendee, "mountain-trip,mountain-trip")
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and can be read again")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status for update")
	attendeeReadAgain := tstReadAttendee(t, loc)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match")
	require.EqualValues(t, "attendance,mountain-trip,mountain-trip,room-none,sponsor2,stage", attendeeReadAgain.Packages, "attendee data read did not match expected package value")

	docs.Then("and the expected mail messages have been sent")
	tstRequireMailRequests(t, expectedMails)
}

func TestUpdateExistingAttendee_AddPackageMultipleTimes_UserAllowed(t *testing.T) {
	for i, origStatus := range []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Cancelled} {
		testcase := fmt.Sprintf("ua3d-%d", i+1)
		token := tstValidStaffToken(t, 1) // user who registered, staff makes no difference
		targetStatus := origStatus
		mails := []mailservice.MailSendDto{}
		if origStatus == status.Approved {
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 315.00, 315.00)
			mails = append(mails, mail)
		} else if origStatus == status.PartiallyPaid {
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 160.00, 315.00)
			mails = append(mails, mail)
		} else if origStatus == status.Paid {
			targetStatus = status.PartiallyPaid
			mail := tstNewStatusMailWithAmounts(testcase, targetStatus, 60.00, 315.00)
			mails = append(mails, mail)
		}
		t.Run(string(origStatus), func(t *testing.T) {
			tstUpdateExistingAttendee_AddPackageMultipleTimes(t, testcase, origStatus, token, mails)
		})
	}
}

func TestUpdateExistingAttendee_AddPackageTooManyTimes(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee in status paid")
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, "ua3e", status.Paid)
	token := tstValidUserToken(t, 1)

	docs.When("when they send updated attendee info and try to add a package too many times")
	changedAttendee := att
	tstAddPackages(&changedAttendee, "mountain-trip,mountain-trip,mountain-trip")
	changedAttendee.Packages = att.Packages // should be ignored, so let's make it produce no error if used
	response := tstPerformPut(loc, tstRenderJson(changedAttendee), token)

	docs.Then("then the request fails with the expected error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.data.invalid", url.Values{
		"packages_list": []string{"package mountain-trip occurs too many times, can occur at most 2 times"},
	})
}

// --- get attendee ---

func TestDenyReadExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "ga1-")

	docs.Given("given a user who is not logged in")

	docs.When("when they attempt to read attendee info while not logged in")
	readResponse := tstPerformGet(location1, tstNoToken())

	docs.Then("then the request is denied")
	require.Equal(t, http.StatusUnauthorized, readResponse.status, "unexpected http response status for insecure read")
}

func TestDenyReadExistingAttendee_Other(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given two users, the second of which has registered")
	token1 := tstValidUserToken(t, 101)
	location2, _ := tstRegisterAttendee(t, "ga2b-")

	docs.When("when the first one attempts to read the attendee info of the second one, i.e. of someone else")
	readResponse := tstPerformGet(location2, token1)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, readResponse, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestAllowReadExistingAttendee_Self(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, attendee1 := tstRegisterAttendeeWithToken(t, "ga3-", token)

	docs.When("when they request their own information")
	readResponse := tstPerformGet(location1, token)

	docs.Then("then the attendee is successfully read and the response is as expected")
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readResponse.body, &attendeeReadAgain)
	require.EqualValues(t, attendee1, attendeeReadAgain, "attendee data read did not match updated data")
}

func TestDenyReadExistingAttendeeWithStaffToken(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration (the other config doesn't even have a staff token)")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given two existing users, the first of which is staff, and the second of which has registered")
	token := tstValidStaffToken(t, 202)
	location2, _ := tstRegisterAttendee(t, "ga4b-")

	docs.When("when the staffer attempts to read the attendee info of another user")
	readResponse := tstPerformGet(location2, token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, readResponse, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestReadAttendeeNotFound(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42", token)

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestReadAttendeeInvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/smiling", token)

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestDenyReadExistingAttendee_Self_AccessToken_OtherAudience(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "ga5-", token)

	docs.When("when they attempt to read their information using a token for a different audience")
	readResponse := tstPerformGet(location1, "access_other_audience_101")

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, readResponse, http.StatusUnauthorized, "auth.unauthorized", "invalid bearer token")
}

// --- get due date ---

func tstPrepareApprovedAttendeeWithDueDate(t *testing.T, testcase string, token string) (location string, dtoWithId attendee.AttendeeDto) {
	location, dtoWithId = tstRegisterAttendeeWithToken(t, testcase, token)

	ctx := context.Background()
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(dtoWithId.Id, status.Approved))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(dtoWithId.Id, paymentservice.Due, 25500))
	tstUpdateCache(ctx, dtoWithId.Id, 25500, 0, "2022-12-22")

	return
}

func tstExpectDueDate(t *testing.T, responseBody string, expectedDueDate string) {
	actualDueDateResponse := attendee.DueDate{}
	expectedDueDateResponse := attendee.DueDate{
		DueDate: expectedDueDate,
	}
	tstParseJson(responseBody, &actualDueDateResponse)
	require.EqualValues(t, expectedDueDateResponse, actualDueDateResponse, "due date did not match expectations")
}

func TestDenyReadDueDateWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "gdd1-")

	docs.Given("given a user who is not logged in")

	docs.When("when they attempt to read the due date while not logged in")
	readResponse := tstPerformGet(location1+"/due-date", tstNoToken())

	docs.Then("then the request is denied")
	require.Equal(t, http.StatusUnauthorized, readResponse.status, "unexpected http response status for insecure read")
}

func TestDenyReadDueDate_Other(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given two users, the second of which has registered")
	token1 := tstValidUserToken(t, 101)
	location2, _ := tstRegisterAttendee(t, "gdd2-")

	docs.When("when the first one attempts to read the due date of the second one, i.e. of someone else")
	readResponse := tstPerformGet(location2+"/due-date", token1)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, readResponse, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestAllowReadDueDateAttendee_Self(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a new existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "gdd3-", token)

	docs.When("when they request their own due date")
	readResponse := tstPerformGet(location1+"/due-date", token)

	docs.Then("then the due date is successfully read, but it is empty")
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "")
}

func TestAllowReadDueDateAttendee_Self_HasDueDate(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdd4-", token)

	docs.When("when they request their own due date")
	readResponse := tstPerformGet(location1+"/due-date", token)

	docs.Then("then the due date is successfully read")
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2022-12-22")
}

func TestReadDueDateAttendeeNotFound(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read the due date for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/due-date", token)

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestReadDueDateInvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/smiling/due-date", token)

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestDenyReadDueDate_Self_AccessToken_OtherAudience(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is logged in")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "ga5-", token)

	docs.When("when they attempt to read their due date using a token for a different audience")
	readResponse := tstPerformGet(location1+"/due-date", "access_other_audience_101")

	docs.Then("then the appropriate error response is returned")
	tstRequireErrorResponse(t, readResponse, http.StatusUnauthorized, "auth.unauthorized", "invalid bearer token")
}

// --- override due date ---

func TestOverrideDueDate_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw1-", tstValidUserToken(t, 101))

	docs.When("when an anonymous user tries to override the due date to a later date")
	dueDate := attendee.DueDate{DueDate: "2023-01-16"}
	response := tstPerformPut(location1+"/due-date", tstRenderJson(dueDate), tstNoToken())

	docs.Then("then the request is denied (401) and the response is as expected")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and the due date has not been changed")
	readResponse := tstPerformGet(location1+"/due-date", tstValidAdminToken(t))
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2022-12-22")
}

func TestOverrideDueDate_SelfDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	token := tstValidUserToken(t, 101)
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw2-", token)

	docs.When("when they try to override their due date to a later date")
	dueDate := attendee.DueDate{DueDate: "2023-01-16"}
	response := tstPerformPut(location1+"/due-date", tstRenderJson(dueDate), token)

	docs.Then("then the request is denied (403) and the response is as expected")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and the due date has not been changed")
	readResponse := tstPerformGet(location1+"/due-date", tstValidAdminToken(t))
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2022-12-22")
}

func TestOverrideDueDate_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw3-", tstValidUserToken(t, 101))

	docs.When("when a staffer tries to override their due date to a later date")
	token := tstValidStaffToken(t, 202)
	dueDate := attendee.DueDate{DueDate: "2023-01-16"}
	response := tstPerformPut(location1+"/due-date", tstRenderJson(dueDate), token)

	docs.Then("then the request is denied (403) and the response is as expected")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and the due date has not been changed")
	readResponse := tstPerformGet(location1+"/due-date", tstValidAdminToken(t))
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2022-12-22")
}

func TestOverrideDueDate_AdminAllow(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw4-", tstValidUserToken(t, 101))

	docs.When("when an admin overrides their due date to a later date")
	token := tstValidAdminToken(t)
	dueDate := attendee.DueDate{DueDate: "2023-01-16"}
	response := tstPerformPut(location1+"/due-date", tstRenderJson(dueDate), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status for override due date")

	docs.Then("and the due date has been changed")
	readResponse := tstPerformGet(location1+"/due-date", token)
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2023-01-16")

	docs.Then("and no emails have been sent")
	tstRequireMailRequests(t, nil)
}

func TestOverrideDueDate_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin tries to override a due date but specifies an invalid badge number")
	token := tstValidAdminToken(t)
	dueDate := attendee.DueDate{DueDate: "2023-01-16"}
	response := tstPerformPut("/api/rest/v1/attendees/fur/due-date", tstRenderJson(dueDate), token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestOverrideDueDate_InvalidJson(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw6-", tstValidUserToken(t, 101))

	docs.When("when an admin tries to override a due date but sends a json body with wrong fields")
	token := tstValidAdminToken(t)
	jsonStr := `{"due_days":24}`
	response := tstPerformPut(location1+"/due-date", jsonStr, token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "duedate.parse.error", url.Values{})
}

func TestOverrideDueDate_Admin_Unchanged(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an approved attendee")
	location1, _ := tstPrepareApprovedAttendeeWithDueDate(t, "gdw7-", tstValidUserToken(t, 101))

	docs.When("when an admin overrides their due date to an earlier date")
	token := tstValidAdminToken(t)
	dueDate := attendee.DueDate{DueDate: "2022-08-16"}
	response := tstPerformPut(location1+"/due-date", tstRenderJson(dueDate), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status for override due date")

	docs.Then("and the due date is unchanged")
	readResponse := tstPerformGet(location1+"/due-date", token)
	require.Equal(t, http.StatusOK, readResponse.status, "unexpected http response status")
	tstExpectDueDate(t, readResponse.body, "2022-12-22")

	docs.Then("and no emails have been sent")
	tstRequireMailRequests(t, nil)
}

// --- attendee max id ---

func TestAttendeeMaxIdAvailable(t *testing.T) {
	docs.Given("given an existing attendee")
	tstSetup(false, false, true)
	defer tstShutdown()

	someAttendee := tstBuildValidAttendee("max1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(someAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")

	docs.When("when an unauthenticated user queries the maximum id")
	maxIdResponse := tstPerformGet("/api/rest/v1/attendees/max-id", tstNoToken())
	require.Equal(t, http.StatusOK, maxIdResponse.status, "unexpected http response status for max-id")

	docs.Then("then a positive number is returned")
	responseDto := attendee.AttendeeMaxIdDto{}
	tstParseJson(maxIdResponse.body, &responseDto)
	require.True(t, responseDto.MaxId > 0, "expected a positive number as maximum attendee id")
}

// --- my registrations ---

func TestMyRegistrations_Anon(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my1a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given an anonymous user")

	docs.When("when they request the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", tstNoToken())

	docs.Then("then the request fails (401) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestMyRegistrations_User_None(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my10a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given a different user, who has made no registrations")
	token101 := tstValidUserToken(t, 101)

	docs.When("when they request the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", token101)

	docs.Then("then the request fails (404) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.owned.notfound", "")
}

func TestMyRegistrations_User_One(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my11a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given a different user, who has made a single registration")
	token101 := tstValidUserToken(t, 101)
	reg2 := tstBuildValidAttendee("my11b-")
	reg2response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg2), token101)
	require.Equal(t, http.StatusCreated, reg2response.status, "unexpected http response status")

	docs.When("when they request the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", token101)

	docs.Then("then the request is successful and returns only their registration number")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	actualResult := attendee.AttendeeIdList{}
	tstParseJson(response.body, &actualResult)
	require.Equal(t, 1, len(actualResult.Ids))
	actualLocation := fmt.Sprintf("/api/rest/v1/attendees/%d", actualResult.Ids[0])
	require.Equal(t, reg2response.location, actualLocation, "unexpected id returned")
}

func TestMyRegistrations_User_Two(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my11c-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given a different user, who already has a registration")
	token101 := tstValidUserToken(t, 101)
	reg2 := tstBuildValidAttendee("my11d-")
	reg2response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg2), token101)
	require.Equal(t, http.StatusCreated, reg2response.status, "unexpected http response status")

	docs.When("when they try to make a second registration under their identity")
	reg3 := tstBuildValidAttendee("my11e-")
	reg3response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg3), token101)

	docs.Then("then the request fails with the appropriate error message")
	tstRequireErrorResponse(t, reg3response, http.StatusConflict, "attendee.user.duplicate",
		url.Values{"user": []string{"you already have a registration - please use a separate email address and matching account per person"}})
}

func TestMyRegistrations_User_Deleted(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	testcase := "my11f-"

	docs.Given("given a user, who has made a single registration")
	token101 := tstValidUserToken(t, 101)
	reg1 := tstBuildValidAttendee(testcase)
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token101)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given that registration has been deleted by an admin")
	body := status.StatusChangeDto{
		Status:  status.Deleted,
		Comment: testcase,
	}
	statusResponse := tstPerformPost(reg1response.location+"/status", tstRenderJson(body), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, statusResponse.status)

	docs.When("when the user requests the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", token101)

	docs.Then("then the request fails (404) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.owned.notfound", "")
}

func TestMyRegistrations_Admin_One(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token101 := tstValidUserToken(t, 101)
	reg1 := tstBuildValidAttendee("my12a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token101)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given an admin, who has made one registration")
	admToken := tstValidAdminToken(t)
	reg2 := tstBuildValidAttendee("my12b-")
	reg2response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg2), admToken)
	require.Equal(t, http.StatusCreated, reg2response.status, "unexpected http response status")

	docs.When("when they request the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", admToken)

	docs.Then("then the request is successful and returns only their registration number, even for an admin")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	actualResult := attendee.AttendeeIdList{}
	tstParseJson(response.body, &actualResult)
	require.Equal(t, 1, len(actualResult.Ids))
	actualLocation2 := fmt.Sprintf("/api/rest/v1/attendees/%d", actualResult.Ids[0])
	require.Equal(t, reg2response.location, actualLocation2, "unexpected id returned")
}

func TestMyRegistrations_Admin_Two(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token101 := tstValidUserToken(t, 101)
	reg1 := tstBuildValidAttendee("my12a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token101)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given an admin, who already has a registration")
	admToken := tstValidAdminToken(t)
	reg2 := tstBuildValidAttendee("my12b-")
	reg2response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg2), admToken)
	require.Equal(t, http.StatusCreated, reg2response.status, "unexpected http response status")

	docs.When("when they try to make another registration under their identity")
	reg3 := tstBuildValidAttendee("my12c-")
	reg3response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg3), admToken)

	docs.Then("then the request fails with the appropriate error message")
	tstRequireErrorResponse(t, reg3response, http.StatusConflict, "attendee.user.duplicate",
		url.Values{"user": []string{"you already have a registration - please use a separate email address and matching account per person"}})
}

func TestMyRegistrations_ApiToken(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my20a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.When("when an api requests the list of registrations they own")
	response := tstPerformGet("/api/rest/v1/attendees", tstValidApiToken())

	docs.Then("then the request fails (401) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestMyRegistrations_AccessToken_OtherAudience(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	token1 := tstValidUserToken(t, 1)
	reg1 := tstBuildValidAttendee("my21a-")
	reg1response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg1), token1)
	require.Equal(t, http.StatusCreated, reg1response.status, "unexpected http response status")

	docs.Given("given a different user, who has made a single registration")
	token101 := tstValidUserToken(t, 101)
	reg2 := tstBuildValidAttendee("my21b-")
	reg2response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(reg2), token101)
	require.Equal(t, http.StatusCreated, reg2response.status, "unexpected http response status")

	docs.When("when they request the list of registrations they own using an access token for a different audience")
	response := tstPerformGet("/api/rest/v1/attendees", "access_other_audience_101")

	docs.Then("then the request is successful and returns only their registration number")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	actualResult := attendee.AttendeeIdList{}
	tstParseJson(response.body, &actualResult)
	require.Equal(t, 1, len(actualResult.Ids))
	actualLocation := fmt.Sprintf("/api/rest/v1/attendees/%d", actualResult.Ids[0])
	require.Equal(t, reg2response.location, actualLocation, "unexpected id returned")
}

// --- flags, options, packages, admin-only flags ---

func tstGetChoice_Anon(t *testing.T, testcase string, what string, endpoint string) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, testcase+"-")

	docs.Given("given an anonymous user")

	docs.When("when they request " + what)
	response := tstPerformGet(loc1+endpoint, tstNoToken())

	docs.Then("then the request fails (401) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func tstGetChoice_Self(t *testing.T, testcase string, what string, endpoint string, expectAllow bool, expectValue bool) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given a registered attendee")
	loc1, _ := tstRegisterAttendee(t, testcase+"-")
	token := tstValidUserToken(t, 1)

	docs.When("when they request " + what)
	response := tstPerformGet(loc1+endpoint, token)

	if expectAllow {
		docs.Then("then the request is successful and the response is as expected")
		require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
		actualResult := attendee.ChoiceState{}
		tstParseJson(response.body, &actualResult)
		require.Equal(t, expectValue, actualResult.Present)
	} else {
		docs.Then("then the request fails (403) and the error is as expected")
		tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
	}
}

func tstGetChoice_Other(t *testing.T, testcase string, what string, endpoint string, permissions string, expectAllow bool, expectValue bool) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, testcase+"a-")

	if permissions == "" {
		docs.Given("given another registered attendee, who has no permissions set")
	} else {
		docs.Given("given another registered attendee, who has permissions set to " + permissions)
	}
	token := tstValidUserToken(t, 101)
	loc2, _ := tstRegisterAttendeeWithToken(t, testcase+"b-", token)
	if permissions != "" {
		body := admin.AdminInfoDto{
			Permissions: permissions,
		}
		response := tstPerformPut(loc2+"/admin", tstRenderJson(body), tstValidAdminToken(t))
		require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	}

	docs.When("when they request " + what)
	response := tstPerformGet(loc1+endpoint, token)

	if expectAllow {
		docs.Then("then the request is successful and the response is as expected")
		require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
		actualResult := attendee.ChoiceState{}
		tstParseJson(response.body, &actualResult)
		require.Equal(t, expectValue, actualResult.Present)
	} else {
		docs.Then("then the request fails (403) and the error is as expected")
		tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
	}
}

func tstGetChoice_Admin(t *testing.T, testcase string, what string, endpoint string, expectValue bool) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, testcase+"a-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request " + what)
	response := tstPerformGet(loc1+endpoint, token)

	docs.Then("then the request is successful and the response is as expected")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	actualResult := attendee.ChoiceState{}
	tstParseJson(response.body, &actualResult)
	require.Equal(t, expectValue, actualResult.Present)
}

func tstGetChoice_ApiToken(t *testing.T, testcase string, what string, endpoint string, expectValue bool) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, testcase+"a-")

	docs.Given("given a system authenticating by api token")
	token := tstValidApiToken()

	docs.When("when it requests " + what)
	response := tstPerformGet(loc1+endpoint, token)

	docs.Then("then the request is successful and the response is as expected")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	actualResult := attendee.ChoiceState{}
	tstParseJson(response.body, &actualResult)
	require.Equal(t, expectValue, actualResult.Present)
}

func TestGetFlags_Anon(t *testing.T) {
	tstGetChoice_Anon(t, "flg1", "a flag", "/flags/ev")
}

func TestGetFlagsUnset_Self(t *testing.T) {
	tstGetChoice_Self(t, "flg2", "a visible regular flag from their own registration, which is not set",
		"/flags/ev", true, false)
}

func TestGetFlagsSet_Self(t *testing.T) {
	tstGetChoice_Self(t, "flg3", "a visible regular flag from their own registration, which is set",
		"/flags/hc", true, true)
}

func TestGetAdminFlagsUnset_Self(t *testing.T) {
	tstGetChoice_Self(t, "flg4", "a visible admin flag from their own registration, which is not set",
		"/flags/guest", true, false)
}

func TestGetAdminFlagsInvisible_Self(t *testing.T) {
	tstGetChoice_Self(t, "flg5", "an invisible admin flag from their own registration",
		"/flags/skip_ban_check", false, false)
}

func TestGetFlags_Other_NoPerm(t *testing.T) {
	tstGetChoice_Other(t, "flg6", "a visible regular flag from another registration",
		"/flags/ev", "", false, false)
}

func TestGetAdminFlags_Other_NoPerm(t *testing.T) {
	tstGetChoice_Other(t, "flg7", "a visible admin flag from another registration",
		"/flags/guest", "", false, false)
}

func TestGetAdminFlagsInvisible_Other_NoPerm(t *testing.T) {
	tstGetChoice_Other(t, "flg8", "an invisible admin flag from another registration",
		"/flags/skip_ban_check", "", false, false)
}

func TestGetFlags_Other_PermNonMatch(t *testing.T) {
	tstGetChoice_Other(t, "flg9", "a visible regular flag from another registration (non-matching permission)",
		"/flags/ev", "sponsordesk", false, false)
}

func TestGetFlags_Other_PermMatch_Unset(t *testing.T) {
	tstGetChoice_Other(t, "flg10", "a visible regular flag from another registration (matching permission)",
		"/flags/ev", "regdesk", true, false)
}

func TestGetFlags_Other_PermMatch_Set(t *testing.T) {
	tstGetChoice_Other(t, "flg11", "a visible regular flag from another registration (matching permission)",
		"/flags/anon", "regdesk", true, true)
}

func TestGetAdminFlags_Other_Perm(t *testing.T) {
	tstGetChoice_Other(t, "flg12", "a visible admin flag from another registration",
		"/flags/guest", "sponsordesk", true, false)
}

func TestGetAdminFlagsInvisible_Other_Perm(t *testing.T) {
	tstGetChoice_Other(t, "flg13", "an invisible admin flag from another registration",
		"/flags/skip_ban_check", "sponsordesk", false, false)
}

func TestGetFlags_Admin_Unset(t *testing.T) {
	tstGetChoice_Admin(t, "flg14", "a regular flag which is unset",
		"/flags/ev", false)
}

func TestGetFlags_Admin_Set(t *testing.T) {
	tstGetChoice_Admin(t, "flg15", "a regular flag which is set",
		"/flags/anon", true)
}

func TestGetAdminFlags_Admin(t *testing.T) {
	tstGetChoice_Admin(t, "flg16", "a visible admin flag",
		"/flags/guest", false)
}

func TestGetAdminFlagsInvisible_Admin(t *testing.T) {
	tstGetChoice_Admin(t, "flg17", "an invisible admin flag",
		"/flags/skip_ban_check", false)
}

func TestGetFlags_ApiToken_Set(t *testing.T) {
	tstGetChoice_ApiToken(t, "flg18", "a regular flag which is set",
		"/flags/anon", true)
}

func TestGetAdminFlags_ApiToken(t *testing.T) {
	tstGetChoice_ApiToken(t, "flg19", "a visible admin flag",
		"/flags/guest", false)
}

func TestGetFlags_InvalidId(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a flag, but provide an invalid attendee id")
	response := tstPerformGet("/api/rest/v1/attendees/THIS-is-nonsense/flags/hc", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestGetFlags_FlagDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, "flg21-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a flag that does not exist")
	response := tstPerformGet(loc1+"/flags/does_not_exist", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.param.invalid", url.Values{})
}

func TestGetFlags_AttendeeDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a flag for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/flags/guest", token)

	docs.Then("then the request fails (404) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

func TestGetOptions_Anon(t *testing.T) {
	tstGetChoice_Anon(t, "opt1", "an option", "/options/art")
}

func TestGetOptionsUnset_Self(t *testing.T) {
	tstGetChoice_Self(t, "opt2", "an option from their own registration, which is not set",
		"/options/art", true, false)
}

func TestGetOptionsSet_Self(t *testing.T) {
	tstGetChoice_Self(t, "opt3", "an option from their own registration, which is set",
		"/options/music", true, true)
}

func TestGetOptions_Other_NoPerm(t *testing.T) {
	tstGetChoice_Other(t, "opt4", "an option from another registration",
		"/options/music", "", false, false)
}

func TestGetOptions_Other_PermNonMatch(t *testing.T) {
	tstGetChoice_Other(t, "opt5", "an option from another registration (non-matching permission)",
		"/options/music", "regdesk", false, false)
}

func TestGetOptions_Other_PermMatch_Set(t *testing.T) {
	tstGetChoice_Other(t, "opt6", "an option from another registration (matching permission)",
		"/options/music", "sponsordesk", true, true)
}

func TestGetOptions_Admin_Unset(t *testing.T) {
	tstGetChoice_Admin(t, "opt7", "an option, which is not set",
		"/options/anim", false)
}

func TestGetOptions_Admin_Set(t *testing.T) {
	tstGetChoice_Admin(t, "opt8", "an option, which is set",
		"/options/music", true)
}

func TestGetOptions_ApiToken_Set(t *testing.T) {
	tstGetChoice_ApiToken(t, "opt9", "an option, which is set",
		"/options/suit", true)
}

func TestGetOptions_InvalidId(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request an option, but provide an invalid attendee id")
	response := tstPerformGet("/api/rest/v1/attendees/nonSENse/options/suit", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestGetOptions_OptionDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, "opt11-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request an option that does not exist")
	response := tstPerformGet(loc1+"/options/does_not_exist", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.param.invalid", url.Values{})
}

func TestGetOptions_AttendeeDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request an option for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/options/suit", token)

	docs.Then("then the request fails (404) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

func TestGetPackages_Anon(t *testing.T) {
	tstGetChoice_Anon(t, "pkg1", "a package", "/packages/sponsor")
}

func TestGetPackagesUnset_Self(t *testing.T) {
	tstGetChoice_Self(t, "pkg2", "a package from their own registration, which is not set",
		"/packages/sponsor", true, false)
}

func TestGetPackagesSet_Self(t *testing.T) {
	tstGetChoice_Self(t, "pkg3", "a package from their own registration, which is set",
		"/packages/sponsor2", true, true)
}

func TestGetPackages_Other_NoPerm(t *testing.T) {
	tstGetChoice_Other(t, "pkg4", "a package from another registration",
		"/packages/sponsor", "", false, false)
}

func TestGetPackages_Other_PermNonMatch(t *testing.T) {
	tstGetChoice_Other(t, "pkg5", "a package from another registration (non-matching permission)",
		"/packages/sponsor2", "regdesk", false, false)
}

func TestGetPackages_Other_PermMatch_Set(t *testing.T) {
	tstGetChoice_Other(t, "pkg6", "a package from another registration (matching permission)",
		"/packages/sponsor2", "sponsordesk", true, true)
}

func TestGetPackages_Admin_Unset(t *testing.T) {
	tstGetChoice_Admin(t, "pkg7", "a package, which is not set",
		"/packages/sponsor", false)
}

func TestGetPackages_Admin_Set(t *testing.T) {
	tstGetChoice_Admin(t, "pkg8", "a package, which is set",
		"/packages/sponsor2", true)
}

func TestGetPackages_ApiToken_Set(t *testing.T) {
	tstGetChoice_ApiToken(t, "pkg9", "a package, which is set",
		"/packages/sponsor2", true)
}

func TestGetPackages_InvalidId(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a package, but provide an invalid attendee id")
	response := tstPerformGet("/api/rest/v1/attendees/___fun___/packages/sponsor2", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestGetPackages_PackageDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there are registrations")
	loc1, _ := tstRegisterAttendee(t, "pkg21-")

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a package that does not exist")
	response := tstPerformGet(loc1+"/packages/does_not_exist", token)

	docs.Then("then the request fails (400) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.param.invalid", url.Values{})
}

func TestGetPackages_AttendeeDoesNotExist(t *testing.T) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they request a valid package for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/packages/sponsor2", token)

	docs.Then("then the request fails (404) and the error is as expected")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

// helper functions

func tstReadAttendee(t *testing.T, location string) attendee.AttendeeDto {
	readAgainResponse := tstPerformGet(location, tstValidAdminToken(t))
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	return attendeeReadAgain
}
