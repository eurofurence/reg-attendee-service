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
		Permissions: "sponsordesk",
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
		Permissions: "sponsordesk",
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
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "admin.data.invalid", url.Values{"flags": []string{"flags field must be a comma separated combination of any of guest,skip_ban_check"}})

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

func TestAdminWrite_GuestAfterApproveDoesNotSuppressEmail(t *testing.T) {
	testcase := "admguest2a-"

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

	docs.When("when an admin gives them guest status (with suppressMinorUpdateEmail set)")
	token := tstValidAdminToken(t)
	body := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin?suppressMinorUpdateEmail=yes", tstRenderJson(body), token)
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

	docs.Then("and the expected email messages were STILL sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstGuestMail(testcase),
	})
}

func TestAdminWrite_CancelledGuest(t *testing.T) {
	testcase := "admguest3-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a guest attendee in status paid")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting to guest",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	token := tstValidAdminToken(t)
	bodyGuest := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin", tstRenderJson(bodyGuest), token)
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")

	docs.When("when an admin changes their status to cancelled")
	bodyCancel := status.StatusChangeDto{
		Status:  status.Cancelled,
		Comment: "guest set to status cancelled",
	}
	responseCancel := tstPerformPost(loc+"/status", tstRenderJson(bodyCancel), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseCancel.status)

	docs.Then("then the status changes to 'cancelled'")
	tstVerifyStatus(t, loc, status.Cancelled)

	docs.Then("and the expected transactions were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-25500, "admin info change"),
		// the change introduces no further transactions
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstGuestMail(testcase),
		tstNewCancelMail(testcase, "guest set to status cancelled", 0),
	})
}

func TestAdminWrite_GuestMadeNormal(t *testing.T) {
	testcase := "admguest4-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a guest attendee in status paid")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting to guest",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	token := tstValidAdminToken(t)
	bodyGuest := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin", tstRenderJson(bodyGuest), token)
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")

	docs.When("when an admin removes the guest admin flag")
	bodyGuestRevoke := admin.AdminInfoDto{
		Flags:         "",
		AdminComments: "removed guest again",
	}
	responseRevoke := tstPerformPut(loc+"/admin", tstRenderJson(bodyGuestRevoke), token)
	require.Equal(t, http.StatusNoContent, responseRevoke.status, "unexpected http response status")

	docs.Then("then the status changes to 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected transactions were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-25500, "admin info change"),
		tstValidAttendeeDues(25500, "admin info change"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstGuestMail(testcase),
		tstNewStatusMail(testcase, status.Approved),
	})
}

func TestAdminWrite_GuestMadeNormalEmailSuppressWorks(t *testing.T) {
	testcase := "admguest4a-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a guest attendee in status paid")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting to guest",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	token := tstValidAdminToken(t)
	bodyGuest := admin.AdminInfoDto{
		Flags:         "guest",
		AdminComments: "set to guest",
	}
	response := tstPerformPut(loc+"/admin", tstRenderJson(bodyGuest), token)
	require.Equal(t, http.StatusNoContent, response.status, "unexpected http response status")

	docs.When("when an admin removes the guest admin flag and sets the suppressMinorUpdateEmail flag")
	bodyGuestRevoke := admin.AdminInfoDto{
		Flags:         "",
		AdminComments: "removed guest again",
	}
	responseRevoke := tstPerformPut(loc+"/admin?suppressMinorUpdateEmail=yes", tstRenderJson(bodyGuestRevoke), token)
	require.Equal(t, http.StatusNoContent, responseRevoke.status, "unexpected http response status")

	docs.Then("then the status changes to 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected transactions were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-25500, "admin info change"),
		tstValidAttendeeDues(25500, "admin info change"),
	})

	docs.Then("and the expected email messages were sent via the mail service (minus the notification for the change back to approved)")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstGuestMail(testcase),
		// tstNewStatusMail(testcase, status.Approved), -- NOT sent due to suppressMinorUpdateEmail
	})
}

func TestAdminWrite_ManualDuesPositive_BeforeApprove(t *testing.T) {
	testcase := "admman1-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status new")
	loc, _ := tstRegisterAttendee(t, testcase)

	docs.Given("given an admin has added positive manual dues")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            8000,
		ManualDuesDescription: "you still need to pay for last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.When("when the attendee is approved")
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve after manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.Then("then the status changes to 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(33500, "dues adjustment due to change in status or selected packages"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMailWithAmounts(testcase, status.Approved, 335, 335),
	})
}

func TestAdminWrite_ManualDuesPositive_AfterApprove(t *testing.T) {
	testcase := "admman2-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin adds manual dues")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            8000,
		ManualDuesDescription: "you still need to pay for last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status remains on 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(8000, "you still need to pay for last year"),
	})

	mail1 := tstNewStatusMail(testcase, status.Approved)
	mail2 := tstNewStatusMail(testcase, status.Approved)
	mail2.Variables["total_dues"] = "EUR 335.00"
	mail2.Variables["remaining_dues"] = "EUR 335.00"
	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		mail1,
		mail2,
	})
}

func TestAdminWrite_ManualDuesPositive_AfterApproveSuppressEmailWorks(t *testing.T) {
	testcase := "admman2-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin adds manual dues (with the suppressMinorUpdateEmail flag set)")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            8000,
		ManualDuesDescription: "you still need to pay for last year",
	}
	responseManual := tstPerformPut(loc+"/admin?suppressMinorUpdateEmail=yes", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status remains on 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(8000, "you still need to pay for last year"),
	})

	mail1 := tstNewStatusMail(testcase, status.Approved)
	docs.Then("and the expected email messages were sent via the mail service, not including the manual dues update")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		mail1,
	})
}

func TestAdminWrite_ManualDuesPositive_AfterPaid(t *testing.T) {
	testcase := "admman3-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status paid")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, status.Paid)

	docs.When("when an admin adds manual dues")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            8000,
		ManualDuesDescription: "you still need to pay for last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status goes back to 'partially paid'")
	tstVerifyStatus(t, loc, status.PartiallyPaid)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(8000, "you still need to pay for last year"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMailWithAmounts(testcase, status.PartiallyPaid, 80, 335),
	})
}

func TestAdminWrite_ManualDuesPositive_AfterCheckedIn(t *testing.T) {
	testcase := "admman4-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status checked in")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, status.CheckedIn)

	docs.When("when an admin adds manual dues")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            8000,
		ManualDuesDescription: "you still need to pay for last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status stays at 'checked in'")
	tstVerifyStatus(t, loc, status.CheckedIn)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(8000, "you still need to pay for last year"),
	})

	docs.Then("and no email messages were sent via the mail service")
	tstRequireMailRequests(t, nil)
}

func TestAdminWrite_ManualDuesNegativePartial_BeforeApprove(t *testing.T) {
	testcase := "admman5-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status new")
	loc, _ := tstRegisterAttendee(t, testcase)

	docs.Given("given an admin has added positive manual dues")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            -12000,
		ManualDuesDescription: "we owe you this from last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.When("when the attendee is approved")
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve after negative partial manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.Then("then the status changes to 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(13500, "dues adjustment due to change in status or selected packages"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMailWithAmounts(testcase, status.Approved, 135, 135),
	})
}

// TODO maybe it would make it easier to understand if manual dues always caused separate transactions?

func TestAdminWrite_ManualDuesNegativeFull_BeforeApprove(t *testing.T) {
	testcase := "admman6-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status new")
	loc, _ := tstRegisterAttendee(t, testcase)

	docs.Given("given an admin has added positive manual dues that cover their complete fee")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            -25500,
		ManualDuesDescription: "comped from last year",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.When("when the attendee is approved")
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve after negative partial manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.Then("then the status goes right to 'paid'")
	tstVerifyStatus(t, loc, status.Paid)

	docs.Then("and no dues were booked in the payment service")
	tstRequireTransactions(t, nil)

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMailWithAmounts(testcase, status.Paid, 0, 0),
	})
}

func TestAdminWrite_ManualDuesNegativePartial_AfterApprove(t *testing.T) {
	testcase := "admman7-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin adds negative manual dues that cover part of the fee")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            -8000,
		ManualDuesDescription: "we owe you this for something",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status stays at 'approved'")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-8000, "we owe you this for something"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	mail1 := tstNewStatusMail(testcase, status.Approved)
	mail2 := tstNewStatusMail(testcase, status.Approved)
	mail2.Variables["total_dues"] = "EUR 175.00"
	mail2.Variables["remaining_dues"] = "EUR 175.00"
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		mail1,
		mail2,
	})
}

func TestAdminWrite_ManualDuesNegativeFull_AfterApprove(t *testing.T) {
	testcase := "admman8-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin adds negative manual dues that cover their full fee")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            -26000,
		ManualDuesDescription: "we are so sorry",
	}
	responseManual := tstPerformPut(loc+"/admin", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status goes to 'paid'")
	tstVerifyStatus(t, loc, status.Paid)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-26000, "we are so sorry"),
	})

	docs.Then("and the expected email messages were sent via the mail service")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
		tstNewStatusMailWithAmounts(testcase, status.Paid, -5, -5),
	})
}

func TestAdminWrite_ManualDuesNegativeFull_AfterApproveSuppressEmailWorks(t *testing.T) {
	testcase := "admman8a-"

	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendee(t, testcase)
	bodyApprove := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "approve before setting manual dues",
	}
	responseApprove := tstPerformPost(loc+"/status", tstRenderJson(bodyApprove), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseApprove.status)

	docs.When("when an admin adds negative manual dues that cover their full fee (setting the suppressMinorUpdateEmail flag)")
	bodyManual := admin.AdminInfoDto{
		ManualDues:            -26000,
		ManualDuesDescription: "we are so sorry",
	}
	responseManual := tstPerformPut(loc+"/admin?suppressMinorUpdateEmail=yes", tstRenderJson(bodyManual), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, responseManual.status, "unexpected http response status")
	require.Equal(t, "", responseManual.body, "unexpected response body")

	docs.Then("then the status goes to 'paid'")
	tstVerifyStatus(t, loc, status.Paid)

	docs.Then("and the expected dues were booked in the payment service")
	tstRequireTransactions(t, []paymentservice.Transaction{
		tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages"),
		tstValidAttendeeDues(-26000, "we are so sorry"),
	})

	docs.Then("and the expected email messages were sent via the mail service, minus the paid status email")
	tstRequireMailRequests(t, []mailservice.MailSendDto{
		tstNewStatusMail(testcase, status.Approved),
	})
}

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

func TestSearch_RegdeskOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who has been given the regdesk permission")
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, "search2a-", status.Approved)
	permBody := admin.AdminInfoDto{
		Permissions: "regdesk",
	}
	permissionResponse := tstPerformPut(loc+"/admin", tstRenderJson(permBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, permissionResponse.status)

	docs.When("when they search for attendees")
	token := tstValidUserToken(t, att.Id)
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
		FillFields: []string{"all"},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is successful and the list of attendees is returned, but they get a limited set of fields")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	expected := `{
  "attendees": [
    {
      "id": 1,
      "badge_id": "1C",
      "nickname": "BlackCheetah",
      "first_name": "Hans",
      "last_name": "Mustermann",
      "country": "DE",
      "birthday": "1998-11-23",
      "pronouns": "he/him",
      "tshirt_size": "XXL",
      "spoken_languages": "de,en",
      "spoken_languages_list": ["de","en"],
      "registration_language": "en-US",
      "flags": "anon,hc,terms-accepted",
      "flags_list": ["anon","hc","terms-accepted"],
      "options": "music,suit",
      "options_list": ["music","suit"],
      "packages": "room-none,attendance,stage,sponsor2",
      "packages_list": [
        {
          "name": "attendance",
          "count": 1
        },
        {
          "name": "room-none",
          "count": 1
        },
        {
          "name": "sponsor2",
          "count": 1
        },
        {
          "name": "stage",
          "count": 1
        }
      ],
      "status": "approved",
      "total_dues": 25500,
      "payment_balance": 0,
      "current_dues": 25500
    }
  ]
}`
	tstRequireSearchResultMatches(t, expected, response.body)
}

func TestSearch_SponsordeskOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who has been given the sponsordesk permission")
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, "search2b-", status.Paid)
	permBody := admin.AdminInfoDto{
		Permissions: "sponsordesk",
	}
	permissionResponse := tstPerformPut(loc+"/admin", tstRenderJson(permBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, permissionResponse.status)

	docs.When("when they search for attendees")
	token := tstValidUserToken(t, att.Id)
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
		FillFields: []string{"all"},
	}
	response := tstPerformPost("/api/rest/v1/attendees/find", tstRenderJson(searchAll), token)

	docs.Then("then the request is successful and the list of attendees is returned, but they get a limited set of fields")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	expected := `{
  "attendees": [
    {
      "id": 1,
      "badge_id": "1C",
      "nickname": "BlackCheetah",
      "first_name": "Hans",
      "last_name": "Mustermann",
      "country": "DE",
      "birthday": "1998-11-23",
      "pronouns": "he/him",
      "tshirt_size": "XXL",
      "spoken_languages": "de,en",
      "spoken_languages_list": ["de","en"],
      "registration_language": "en-US",
      "flags": "anon,hc,terms-accepted",
      "flags_list": ["anon","hc","terms-accepted"],
      "options": "music,suit",
      "options_list": ["music","suit"],
      "packages": "room-none,attendance,stage,sponsor2",
      "packages_list": [
        {
          "name": "attendance",
          "count": 1
        },
        {
          "name": "room-none",
          "count": 1
        },
        {
          "name": "sponsor2",
          "count": 1
        },
        {
          "name": "stage",
          "count": 1
        }
      ],
      "status": "paid",
      "total_dues": 25500,
      "payment_balance": 25500,
      "current_dues": 0
    }
  ]
}`
	tstRequireSearchResultMatches(t, expected, response.body)
}

func TestSearch_OtherPermissionsDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who has been given the myarea permission")
	loc, att := tstRegisterAttendee(t, "search2c-")
	permBody := admin.AdminInfoDto{
		Permissions: "myarea",
	}
	permissionResponse := tstPerformPut(loc+"/admin", tstRenderJson(permBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, permissionResponse.status)

	docs.When("when they attempt to search for attendees")
	token := tstValidUserToken(t, att.Id)
	searchAll := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{},
		},
		FillFields: []string{"all"},
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
      "badge_id": "1C",
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
      "spoken_languages": "de,en",
      "spoken_languages_list": ["de","en"],
      "registration_language": "en-US",
      "flags": "anon,hc,terms-accepted",
      "flags_list": ["anon","hc","terms-accepted"],
      "options": "music,suit",
      "options_list": ["music","suit"],
      "packages": "room-none,attendance,stage,sponsor2",
      "packages_list": [
        {
          "name": "attendance",
          "count": 1
        },
        {
          "name": "room-none",
          "count": 1
        },
        {
          "name": "sponsor2",
          "count": 1
        },
        {
          "name": "stage",
          "count": 1
        }
      ],
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
