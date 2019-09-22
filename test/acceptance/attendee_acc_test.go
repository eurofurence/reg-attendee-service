package acceptance

import (
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
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

	docs.Then( "then the attendee is successfully created")
	require.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	require.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeInvalid(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee with invalid data")
	attendeeSent := tstBuildValidAttendee()
	attendeeSent.Nickname = "$%&^@!$"
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

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
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

	docs.Then( "then the attendee is successfully created and its data can be read again")
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
	// TODO the "logged in" part is not implemented yet
	existingAttendee := tstBuildValidAttendee()
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee))
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(creationResponse.location)

	docs.When( "when they send updated attendee info")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	// TODO change all fields except id
	updateResponse := tstPerformPost(creationResponse.location, tstRenderJson(changedAttendee))

	docs.Then( "then the attendee is successfully updated and the changed data can be read again")
	require.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	require.Equal(t, creationResponse.location, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(creationResponse.location)
	require.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}

// helper functions

func tstReadAttendee(location string) attendee.AttendeeDto {
	readAgainResponse := tstPerformGet(location)
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	return attendeeReadAgain
}

