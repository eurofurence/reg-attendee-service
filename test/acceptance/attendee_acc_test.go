package acceptance

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/stretchr/testify/require"
	"net/http"
	"github.com/jumpy-squirrel/rexis-go-attendee/api/v1/attendee"
	"github.com/jumpy-squirrel/rexis-go-attendee/docs"
	"testing"
)

// ------------------------------------------
// acceptance tests for the attendee resource
// ------------------------------------------

// see config in setup_acc_test.go

func TestCreateNewAttendee(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee()
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then( "then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeInvalid(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee with invalid data")
	attendeeSent := tstBuildValidAttendee()
	attendeeSent.Nickname = "$%&^@!$"
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then( "then the attendee is rejected with an error response")
	require.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := attendee.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeCanBeReadAgain(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee")
	attendeeSent := tstBuildValidAttendee()
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent), tstNoToken())

	docs.Then( "then the attendee is successfully created and its data can be read again by an admin")
	// TODO would need admin authentication, not implemented yet
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	attendeeReadAgain := tstReadAttendee(response.location)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	require.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match sent data")
}

func TestUpdateExistingAttendee(t *testing.T) {
	docs.Given("given an existing attendee, who is logged in")

	existingAttendee := tstBuildValidAttendee()
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstValidToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(creationResponse.location)

	docs.When( "when they send updated attendee info")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	// TODO change all fields except id
	updateResponse := tstPerformPost(creationResponse.location, tstRenderJson(changedAttendee), tstValidToken())

	docs.Then( "then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, creationResponse.location, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(creationResponse.location)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}

func TestDenyUpdateExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given an existing attendee and a user who is not logged in")
	existingAttendee := tstBuildValidAttendee()
	existingAttendee.FirstName = "Marianne"
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(creationResponse.location)

	docs.When( "when they send updated attendee info while not logged in")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	updateResponse := tstPerformPost(creationResponse.location, tstRenderJson(changedAttendee), tstNoToken())

	docs.Then( "then the request is denied and the data remains unchanged")
	require.Equal(t, http.StatusUnauthorized, updateResponse.status, "unexpected http response status for insecure update")
	attendeeReadAgain := tstReadAttendee(creationResponse.location)
	require.EqualValues(t, "Marianne", attendeeReadAgain.FirstName, "attendee data read did not match original data")
}

func TestDenyReadExistingAttendeeWhileNotLoggedIn(t *testing.T) {
	docs.Given("given an existing attendee and a user who is not logged in")
	existingAttendee := tstBuildValidAttendee()
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee), tstNoToken())
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")

	docs.When( "when they attempt to read attendee info while not logged in")
	readResponse := tstPerformGet(creationResponse.location, tstNoToken())

	docs.Then( "then the request is denied")
	require.Equal(t, http.StatusUnauthorized, readResponse.status, "unexpected http response status for insecure read")
}

// helper functions

func tstReadAttendee(location string) attendee.AttendeeDto {
	readAgainResponse := tstPerformGet(location, tstValidToken())
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	return attendeeReadAgain
}

func tstNoToken() string {
	return ""
}

func tstValidToken() string {
	return config.FixedToken()
}
