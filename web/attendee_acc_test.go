package web

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"rexis/rexis-go-attendee/api/v1/attendee"
	"rexis/rexis-go-attendee/docs"
	"testing"
)

// see config in setup_all_test.go

func tstBuildValidAttendee() attendee.AttendeeDto {
	return attendee.AttendeeDto{
		Nickname:     "BlackCheetah",
		FirstName:    "Hans",
		LastName:     "Mustermann",
		Street:       "Teststraße 24",
		Zip:          "12345",
		City:         "Berlin",
		Country:      "DE",
		CountryBadge: "DE",
		State:        "Sachsen",
		Email:        "jsquirrel_github_9a6d@packetloss.de",
		Phone:        "+49-30-123",
		Telegram:     "@ihopethisuserdoesnotexist",
		Birthday:     "1998-11-23",
		Gender:       "other",
		Flags:        "anon,ev",
		Packages:     "room-none,attendance,stage,sponsor2",
		Options:      "music,suit",
		TshirtSize:   "XXL",
	}
}

func tstRenderJson(v interface{}) string {
	representationBytes, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return string(representationBytes)
}

// tip: dto := &attendee.AttendeeDto{}
func tstParseJson(body string, dto interface{}) {
	err := json.Unmarshal([]byte(body), dto)
	if err != nil {
		log.Fatal(err)
	}
}

func tstReadAttendee(location string) attendee.AttendeeDto {
	readAgainResponse := tstPerformGet(location)
	attendeeReadAgain := attendee.AttendeeDto{}
	tstParseJson(readAgainResponse.body, &attendeeReadAgain)
	return attendeeReadAgain
}

func TestCreateNewAttendee(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee with valid data")
	attendeeSent := tstBuildValidAttendee()
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

	docs.Then( "then the attendee is successfully created")
	assert.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	assert.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")
}

func TestCreateNewAttendeeInvalid(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee with invalid data")
	attendeeSent := tstBuildValidAttendee()
	attendeeSent.Nickname = "$%&^@!$"
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

	docs.Then( "then the attendee is rejected with an error response")
	assert.Equal(t, http.StatusBadRequest, response.status, "unexpected http response status")
	errorDto := attendee.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	assert.Equal(t, "attendee.data.invalid", errorDto.Message, "unexpected error code")
}

func TestCreateNewAttendeeCanBeReadAgain(t *testing.T) {
	docs.Given("given an unauthenticated user")

	docs.When( "when they create a new attendee")
	attendeeSent := tstBuildValidAttendee()
	response := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(attendeeSent))

	docs.Then( "then the attendee is successfully created and its data can be read again")
	// TODO would need admin authentication, not implemented yet
	assert.Equal(t, http.StatusCreated, response.status, "unexpected http response status")
	assert.Regexp(t, "^\\/api\\/rest\\/v1\\/attendees\\/[1-9][0-9]*$", response.location, "invalid location header in response")

	attendeeReadAgain := tstReadAttendee(response.location)
	// difference in id is ok, so copy it over to expected
	attendeeSent.Id = attendeeReadAgain.Id
	assert.EqualValues(t, attendeeSent, attendeeReadAgain, "attendee data read did not match sent data")
}

func TestUpdateExistingAttendee(t *testing.T) {
	docs.Given("given an existing attendee, who is logged in")
	// TODO the "logged in" part is not implemented yet
	existingAttendee := tstBuildValidAttendee()
	creationResponse := tstPerformPut("/api/rest/v1/attendees", tstRenderJson(existingAttendee))
	assert.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status for create")
	attendeeReadAfterCreation := tstReadAttendee(creationResponse.location)

	docs.When( "when they send updated attendee info")
	changedAttendee := attendeeReadAfterCreation
	changedAttendee.FirstName = "Eva"
	changedAttendee.LastName = "Musterfrau"
	// TODO change all fields except id
	updateResponse := tstPerformPost(creationResponse.location, tstRenderJson(changedAttendee))

	docs.Then( "then the attendee is successfully updated and the changed data can be read again")
	assert.Equal(t, http.StatusOK, updateResponse.status, "unexpected http response status for update")
	assert.Equal(t, creationResponse.location, updateResponse.location, "location unexpectedly changed during update")
	attendeeReadAgain := tstReadAttendee(creationResponse.location)
	assert.EqualValues(t, changedAttendee, attendeeReadAgain, "attendee data read did not match updated data")
}
