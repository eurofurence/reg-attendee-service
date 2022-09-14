package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/errorapi"
	"net/http"
	"testing"

	"github.com/eurofurence/reg-attendee-service/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------
// acceptance tests for the attendee resource
// ------------------------------------------

// TODO test invalid id in path (400)
// TODO test non-existent id in path (404)

// --- create new attendee ---

func TestCreateNewAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee with valid data after public registration has begun")
	attendeeSent := tstBuildValidAttendee("na1-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee with invalid data")
	attendeeSent := tstBuildValidAttendee("na2-")
	attendeeSent.Nickname = "$%&^@!$"
	attendeeSent.Packages = attendeeSent.Packages + ",sponsor" // a constraint violation
	attendeeSent.Birthday = "2004-11-23"                       // too young
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
	require.Equal(t, "nickname field must contain at least one alphanumeric character", errorDto.Details.Get("nickname"))
	require.Equal(t, "cannot pick both sponsor2 and sponsor - constraint violated", errorDto.Details.Get("packages"))
	require.Equal(t, "birthday must be no earlier than 1901-01-01 and no later than 2001-08-14", errorDto.Details.Get("birthday"))
}

func TestCreateNewAttendeeSyntaxInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they try to create a new attendee with syntactically invalid data")
	attendeeSent := tstBuildValidAttendee("na2a-")
	syntaxErrorJson := "{" + tstRenderJson(attendeeSent)
	response := tstPerformPost("/api/rest/v1/attendees", syntaxErrorJson, tstNoToken())

	docs.Then("then the attendee is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.parse.error", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeCanBeReadAgainByAdmin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they create a new attendee")
	attendeeSent := tstBuildValidAttendee("na3-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is successfully created and its data can be read again by an admin")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	attendeeReadAgain := tstReadAttendee(t, response.location)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match sent data")
}

func TestCreateNewAttendeeStaffregNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na4-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and no location header is supplied")
	require.Equal(t, http.StatusUnauthorized, response.status, "unexpected http response status")
	require.Equal(t, "", response.location, "non-empty location header in response")
}

func TestCreateNewAttendeeStaffregStaff(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given a staffer")
	staffToken := tstValidStaffToken(t)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na5-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), staffToken)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeStaffregUser(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated regular user")
	userToken := tstValidUserToken(t)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na6-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), userToken)

	docs.Then("then the request is denied as unauthorized (403) and no location header is supplied")
	require.Equal(t, http.StatusForbidden, response.status, "unexpected http response status")
	require.Equal(t, "", response.location, "non-empty location header in response")
}

func TestCreateNewAttendeeStaffregAdmin(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	adminToken := tstValidAdminToken(t)

	docs.When("when they attempt to create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee("na7-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), adminToken)

	docs.Then("then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeAdminOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they send a new attendee and attempt to set an admin only flag (guest)")
	attendeeSent := tstBuildValidAttendee("na8-")
	attendeeSent.Flags = "guest,hc"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeReadOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they send a new attendee and attempt to set a read only flag (ev)")
	attendeeSent := tstBuildValidAttendee("na8-")
	attendeeSent.Flags = "ev,anon"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attendee is rejected with an error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeAdminOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send a new attendee and attempt to set an admin only flag (guest)")
	attendeeSent := tstBuildValidAttendee("na8-")
	attendeeSent.Flags = "guest"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), token)

	docs.Then("then the attendee is rejected with an error response, because admin only flags belong in adminInfo")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeDefaultReadOnlyPackage(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given a staffer")
	staffToken := tstValidStaffToken(t)

	docs.When("when they send a new attendee and attempt to leave out a read only default package (room-none)")
	attendeeSent := tstBuildValidAttendee("na9-")
	attendeeSent.Packages = "attendance,stage,sponsor"
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), staffToken)

	docs.Then("then the attendee is rejected with an error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeDuplicateHandling(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
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
	require.Equal(t, http.StatusBadRequest, duplicateResponse.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(duplicateResponse.body, &errorDto)
	require.Equal(t, "attendee.data.duplicate", errorDto.Message, "unexpected error code")
	require.Equal(t, "there is already an attendee with this information (looking at nickname, email, and zip code)",
		errorDto.Details["attendee"][0], "unexpected detail info")
}

func TestCreateNewAttendeeTooEarly(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFileBeforeTarget)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")

	docs.When("when they attempt to create a new attendee with valid data before public registration has begun")
	attendeeSent := tstBuildValidAttendee("na11-")
	response := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then("then the attempt is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
	require.Equal(t, "public registration has not opened at this time, please come back later", errorDto.Details.Get("timing"))
}

// --- update attendee ---

func TestUpdateExistingAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they send updated attendee info while logged in")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. This is ok because only the old regsys will know the user / admin tokens.")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), tstValidUserToken(t))

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, creationResponse.location, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}

func TestUpdateExistingAttendeeSyntaxInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua1a-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they try to update the attendee with syntactically invalid data")
	syntaxErrorJson := "{" + tstRenderJson(attendeeReadAfterCreation)
	response := tstPerformPut(creationResponse.location, syntaxErrorJson, tstValidUserToken(t))

	docs.Then("then the update is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.parse.error", errorDto.Message, "unexpected error code")
}

func TestUpdateExistingAttendeeDataInvalid(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua1b-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they try to update their information with invalid data")
	attendeeReadAfterCreation.Nickname = "$%&^@!$"
	attendeeReadAfterCreation.Packages = attendeeReadAfterCreation.Packages + ",sponsor" // a constraint violation
	attendeeReadAfterCreation.Birthday = "2004-11-23"                                    // too young
	response := tstPerformPut(creationResponse.location, tstRenderJson(attendeeReadAfterCreation), tstValidUserToken(t))

	docs.Then("then the update is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
	require.Equal(t, "nickname field must contain at least one alphanumeric character", errorDto.Details.Get("nickname"))
	require.Equal(t, "cannot pick both sponsor2 and sponsor - constraint violated", errorDto.Details.Get("packages"))
	require.Equal(t, "birthday must be no earlier than 1901-01-01 and no later than 2001-08-14", errorDto.Details.Get("birthday"))
}

func TestUpdateNonExistingAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update a non-existent attendee")
	nonExistingAttendee := tstBuildValidAttendee("ua1c-")
	response := tstPerformPut("/api/rest/v1/attendees/42", tstRenderJson(nonExistingAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response")
	require.Equal(t, http.StatusNotFound, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.id.notfound", errorDto.Message, "unexpected error code")
}

func TestUpdateAttendeeInvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to update an attendee with an invalid id")
	nonExistingAttendee := tstBuildValidAttendee("ua1d-")
	response := tstPerformPut("/api/rest/v1/attendees/helloworld", tstRenderJson(nonExistingAttendee), token)

	docs.Then("then the update is rejected with an appropriate error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.id.invalid", errorDto.Message, "unexpected error code")
}

func TestDenyUpdateExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee and a user who is not logged in")
	existingAttendee := tstBuildValidAttendee("ua2-")
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

func TestDenyUpdateExistingAttendeeWithStaffToken(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration (the other config doesn't even have a staff token)")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua3-")
	existingAttendee.FirstName = "Marianne"
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstValidAdminToken(t))
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when a logged in staffer sends updated attendee info")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), tstValidStaffToken(t))

	docs.Then("then the request is denied as unauthorized (403) and the data remains unchanged")
	require.Equal(t, http.StatusForbidden, updateResponse.status, "unexpected http response status for insecure update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "Marianne", attendeeReadAgain.FirstName, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeAdminOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they send updated attendee info and attempt to add an admin-only flag (guest)")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.Flags = changedAttendee.Flags + ",guest"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), tstValidUserToken(t))

	docs.Then("then the request is denied and the data remains unchanged, because admin only flags belong in adminInfo")
	require.Equal(t, http.StatusBadRequest, updateResponse.status, "unexpected http response status for malicious update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "anon,hc", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeAdminOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send updated attendee info and attempt to add an admin-only flag (guest)")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.Flags = changedAttendee.Flags + ",guest"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), token)

	docs.Then("then the request is denied and the data remains unchanged, because admin only flags belong in adminInfo")
	require.Equal(t, http.StatusBadRequest, updateResponse.status, "unexpected http response status for malicious update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "anon,hc", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeReadOnlyFlag(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.When("when they send updated attendee info and attempt to add a read-only flag (ev)")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.Flags = changedAttendee.Flags + ",guest"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), tstValidUserToken(t))

	docs.Then("then the request is denied and the data remains unchanged")
	require.Equal(t, http.StatusBadRequest, updateResponse.status, "unexpected http response status for malicious update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, "anon,hc", attendeeReadAgain.Flags, "attendee data read did not match original data")
}

func TestUpdateExistingAttendeeReadOnlyFlag_Admin(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ua4-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(t, creationResponse.location)

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they send updated attendee info and add a read-only flag (ev)")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.Flags = changedAttendee.Flags + ",ev"
	updateResponse := tstPerformPut(creationResponse.location, tstRenderJson(changedAttendee), token)

	docs.Then("then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, creationResponse.location, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(t, creationResponse.location)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
	require.EqualValues(t, "anon,hc,ev", attendeeReadAgain.Flags, "attendee data read did not match expected flags value")
}

// --- get attendee ---

func TestDenyReadExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee and a user who is not logged in")
	existingAttendee := tstBuildValidAttendee("ga1-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")

	docs.When("when they attempt to read attendee info while not logged in")
	readResponse := tstPerformGet(creationResponse.location, tstNoToken())

	docs.Then("then the request is denied")
	require.Equal(t, http.StatusUnauthorized, readResponse.status, "unexpected http response status for insecure read")
}

func TestDenyReadExistingAttendeeWithStaffToken(t *testing.T) {
	docs.Given("given the configuration for staff pre-registration (the other config doesn't even have a staff token)")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	existingAttendee := tstBuildValidAttendee("ga2-")
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstValidAdminToken(t))
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")

	docs.When("when a logged in staffer attempts to read the attendee info")
	readResponse := tstPerformGet(creationResponse.location, tstValidStaffToken(t))

	docs.Then("then the request is denied as unauthorized (403)")
	require.Equal(t, http.StatusForbidden, readResponse.status, "unexpected http response status for insecure read")
}

func TestReadAttendeeNotFound(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42", token)

	docs.Then("then the appropriate error response is returned")
	require.Equal(t, http.StatusNotFound, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.id.notfound", errorDto.Message, "unexpected error code")
}

func TestReadAttendeeInvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to read an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/smiling", token)

	docs.Then("then the appropriate error response is returned")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.id.invalid", errorDto.Message, "unexpected error code")
}

// --- attendee max id ---

func TestAttendeeMaxIdAvailable(t *testing.T) {
	docs.Given("given an existing attendee")
	tstSetup(tstDefaultConfigFile)
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

// helper functions

func tstReadAttendee(t *testing.T, location string) attendee.AttendeeDto {
	readAgainResponse := tstPerformGet(location, tstValidAdminToken(t))
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	return attendeeReadAgain
}

func tstNoToken() string {
	return ""
}

func tstValidUserToken(t *testing.T) string {
	token, err := config.FixedToken(config.TokenForLoggedInUser)
	require.Nil(t, err)
	return token
}

func tstValidAdminToken(t *testing.T) string {
	token, err := config.FixedToken(config.TokenForAdmin)
	require.Nil(t, err)
	return token
}

func tstValidStaffToken(t *testing.T) string {
	token, err := config.FixedToken(config.OptionalTokenForInitialReg)
	require.Nil(t, err)
	require.NotEqual(t, "", token)
	return token
}
