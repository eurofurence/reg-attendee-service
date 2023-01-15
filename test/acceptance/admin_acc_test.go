package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

// ------------------------------------------
// acceptance tests for the admin subresource
// ------------------------------------------

// --- read access

func TestAdminDefaults_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "admr1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the admin information for the attendee")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestAdminDefaults_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admr2-")

	docs.Given("given the same regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access the admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has registered")
	location1, attendee1 := tstRegisterAttendee(t, "admr3-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or anybody else's admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestAdminDefaults_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admr4-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the admin information")
	response := tstPerformGet(location1+"/admin", token)

	docs.Then("then the request is successful and the default admin information is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response.body)
}

func TestReadAdminInfo_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the admin information for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/admin", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestReadAdminInfo_InvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the admin information for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/kittycat/admin", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

// --- write access

func TestAdminWrite_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admw1-")

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Permissions: "admin",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admw2-")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to update the admin information")
	body := admin.AdminInfoDto{
		Permissions: "admin",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is staff")
	location1, attendee1 := tstRegisterAttendee(t, "admw3-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to update their own admin information")
	body := admin.AdminInfoDto{
		Flags: "guest",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no changes have been made")
	response2 := tstPerformGet(location1+"/admin", tstValidAdminToken(t))
	tstRequireAdminInfoMatches(t, admin.AdminInfoDto{Id: attendee1.Id}, response2.body)
}

func TestAdminWrite_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "admw4-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information")
	body := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected response body")

	docs.Then("and the changed admin info can be read again")
	response2 := tstPerformGet(location1+"/admin", token)
	expectedAdminInfo := admin.AdminInfoDto{
		Id:            attendee1.Id,
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	tstRequireAdminInfoMatches(t, expectedAdminInfo, response2.body)
}

func TestAdminWrite_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee that does not exist")
	body := admin.AdminInfoDto{
		AdminComments: "existence is fleeting",
	}
	response := tstPerformPut("/api/rest/v1/attendees/789789/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")
}

func TestAdminWrite_InvalidAttendeeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the admin information for an attendee with an invalid id")
	body := admin.AdminInfoDto{
		AdminComments: "kittens are cuter",
	}
	response := tstPerformPut("/api/rest/v1/attendees/puppy/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestAdminWrite_InvalidBody(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, _ := tstRegisterAttendee(t, "admw5-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they change the admin information but send an invalid json body")
	body := "{{{{:::"
	response := tstPerformPut(location1+"/admin", body, token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.parse.error", "")
}

func TestAdminWrite_CannotChangeId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "admw6-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they attempt to change the id")
	body := admin.AdminInfoDto{
		Id: 9999,
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"id": []string{"id field must be empty or correctly assigned for incoming requests"}})
}

func TestAdminWrite_WrongFlagType(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "admw7-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to set a non-admin only flag")
	body := admin.AdminInfoDto{
		Flags: "ev",
	}
	response := tstPerformPut(location1+"/admin", tstRenderJson(body), token)

	docs.Then("then the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"flags": []string{"flags field must be a comma separated combination of any of guest"}})

	docs.Then("and the admin info is unchanged")
	response2 := tstPerformGet(location1+"/admin", token)
	expectedAdminInfo := admin.AdminInfoDto{
		Id: attendee1.Id,
	}
	tstRequireAdminInfoMatches(t, expectedAdminInfo, response2.body)
}

// --- guest and manual dues ---

func TestAdminWrite_GuestBeforeApprove(t *testing.T) {
	testcase := "admguest1-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status new")
	loc, _ := tstRegisterAttendee(t, testcase)

	docs.Given("given an admin has given them guest status")
	token := tstValidAdminToken(t)
	body := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin", tstRenderJson(body), token)
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected response body")

	docs.When("when the attendee is approved")
	body2 := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve after setting to guest",
	}
	response2 := tstPerformPost(loc+"/status", tstRenderJson(body2), tstValidAdminToken(t))

	docs.Then("then the status goes right to paid")
	require.Equal(t, http.StatusNoContent, response2.status)
	tstVerifyStatus(t, loc, "paid")

	docs.Then("and NO dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{})

	docs.Then("and the guest email message was sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{tstGuestMail(testcase)})
}

func TestAdminWrite_GuestAfterApprove(t *testing.T) {
	testcase := "admguest2-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	body2 := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting to guest",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(body2), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin gives them guest status")
	token := tstValidAdminToken(t)
	body := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin", tstRenderJson(body), token)
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")
	require.Equal(t, "", response.body, "unexpected response body")

	docs.Then("then the status changes to 'paid'")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, status.Paid)

	docs.Then("and the compensating negative dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-25500, "admin info change"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstGuestMail(testcase),
	})
}

// TODO test dues changes caused by setting and removing guest status and corresponding status change logic

// --- search ---

func TestSearch_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	_, _ = tstRegisterAttendee(t, "search1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to search for attendees")
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is denied as unauthenticated (401) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestSearch_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	_, attendee1 := tstRegisterAttendee(t, "search2-")

	docs.Given("given the same regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to search for attendees")
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestSearch_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has registered")
	_, attendee1 := tstRegisterAttendee(t, "search3-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to search for attendees")
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is denied as unauthorized (403) and the correct error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestSearch_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	_, _ = tstRegisterAttendee(t, "search4-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they search for attendees")
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
		FillFields: []string{"all"},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is successful and the list of attendees is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	expected := `{
  "attendees": [
    {
      "id": 1,
      "badge_id": "1Y",
      "nickname": "BlackCheetah",
      "first_name": "Hans",
      "last_name": "Mustermann",
      "street": "Teststraße 24",
      "zip": "search4-12345",
      "email": "jsquirrel_github_9a6d@packetloss.de",
      "city": "Berlin",
      "country": "DE",
      "state": "Sachsen",
      "phone": "+49-30-123",
      "telegram": "@ihopethisuserdoesnotexist",
      "birthday": "1998-11-23",
      "gender": "other",
      "pronouns": "he/him",
      "tshirt_size": "XXL",
      "spoken_languages": "de-DE,en-US",
      "registration_language": "en-US",
      "flags": "anon,hc",
      "options": "music,suit",
      "packages": "room-none,attendance,stage,sponsor2",
      "status": "new",
      "total_dues": 0,
      "payment_balance": 0,
      "current_dues": 0
    }
  ]
}`
	tstRequireSearchResultMatches(t, expected, response.body)
}

func TestSearch_NonexistentAttendee(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	_, _ = tstRegisterAttendee(t, "search5-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they search for attendees, but specify non-matching criteria")
	searchNon := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{
				Address: "Not something that matches",
			},
		},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchNon), token)

	docs.Then("then the request is successful and an empty list of attendees is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	expected := `{
  "attendees": []
}`
	tstRequireSearchResultMatches(t, expected, response.body)
}

func TestSearch_InvalidJson(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	_, _ = tstRegisterAttendee(t, "search5-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they search for attendees, but send an invalid json body")
	response := tstPerformPost("/api/rest/v1/attendees/find", "{{{{", token)

	docs.Then("then the request fails with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "search.parse.error", url.Values{})
}

// helper functions

func tstRequireAdminInfoMatches(t *testing.T, expected admin.AdminInfoDto, body string) {
	adminInfo := admin.AdminInfoDto{}
	tstParseJson(body, &adminInfo)
	require.EqualValues(t, expected, adminInfo, "admin data did not match expected values")
}

func tstRequireSearchResultMatches(t *testing.T, expectedBody string, body string) {
	expected := attendee.AttendeeSearchResultList{}
	tstParseJson(expectedBody, &expected)

	actual := attendee.AttendeeSearchResultList{}
	tstParseJson(body, &actual)

	// ignore emails and registered field because they contain a timer
	for i := range actual.Attendees {
		actual.Attendees[i].Registered = nil
	}

	require.EqualValues(t, expected, actual, "search result did not match expected values")
}
