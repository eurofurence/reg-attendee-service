package acceptance

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

// -------------------------------------------
// acceptance tests for the status subresource
// -------------------------------------------

// -- read status

func TestStatus_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	attendeeLocation, _ := tstRegisterAttendee(t, "stat1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status")
	response := tstPerformGet(attendeeLocation+"/status", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestStatus_UserDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given two existing regular users, the second of which has registered")
	token := tstValidUserToken(t, 101)
	location2, _ := tstRegisterAttendee(t, "stat2b-")

	docs.When("when the first user attempts to access somebody else's status")
	response := tstPerformGet(location2+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestStatus_UserAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "stat3-", token)

	docs.When("when they access their own status")
	response := tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)
	expectedStatus := status.StatusDto{
		Status: status.New,
	}
	require.EqualValues(t, expectedStatus, statusDto, "status did not match expected value")
}

func TestStatus_UserAllowSelf_OtherAudience(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "stat4-", token)

	docs.When("when they access their own status, using a token with a different audience")
	response := tstPerformGet(location1+"/status", "access_other_audience_101")

	docs.Then("then the request is successful and status 'new' is returned")
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)
	expectedStatus := status.StatusDto{
		Status: status.New,
	}
	require.EqualValues(t, expectedStatus, statusDto, "status did not match expected value")
}

func TestStatus_StaffDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given two existing users, the first of which is staff")
	token1 := tstValidStaffToken(t, 202)
	location2, _ := tstRegisterAttendee(t, "stat4b-")

	docs.When("when the staffer attempts to access somebody else's status")
	response := tstPerformGet(location2+"/status", token1)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized to access this data - the attempt has been logged")
}

func TestStatus_StaffAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an existing attendee who is staff")
	token := tstValidStaffToken(t, 202)
	location1, _ := tstRegisterAttendeeWithToken(t, "stat5-", token)

	docs.When("when they access their own status")
	response := tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)
	expectedStatus := status.StatusDto{
		Status: status.New,
	}
	require.EqualValues(t, expectedStatus, statusDto, "status did not match expected value")
}

func TestStatus_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat6-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status for any attendee")
	response := tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and the default status is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	tstRequireAttendeeStatus(t, status.New, response.body)
}

func TestStatus_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status for an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/panther/status", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestStatus_Nonexistent(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/status", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

// -- status history

func TestStatusHistory_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat20-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")
}

func TestStatusHistory_SelfDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	token := tstValidUserToken(t, 101)
	location1, _ := tstRegisterAttendeeWithToken(t, "stat21-", token)

	docs.When("when they attempt to access their own status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned (status history is admin only)")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestStatusHistory_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "stat21-")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access somebody else's status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestStatusHistory_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has made a valid registration")
	token := tstValidStaffToken(t, 202)
	location1, _ := tstRegisterAttendeeWithToken(t, "stat22-", token)

	docs.When("when they attempt to access their own (or somebody else's) status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")
}

func TestStatusHistory_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an existing attendee right after registration")
	location1, attendee1 := tstRegisterAttendee(t, "stat23-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status history for any attendee")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is successful and the default status history is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	statusHistoryDto := status.StatusHistoryDto{}
	tstParseJson(response.body, &statusHistoryDto)

	require.Equal(t, 1, len(statusHistoryDto.StatusHistory))
	expectedStatusHistory := status.StatusHistoryDto{
		Id: attendee1.Id,
		StatusHistory: []status.StatusChangeDto{{
			Timestamp: statusHistoryDto.StatusHistory[0].Timestamp,
			Status:    status.New,
			Comment:   "registration",
		}},
	}
	require.EqualValues(t, expectedStatusHistory, statusHistoryDto, "status history did not match expected value")
}

func TestStatusHistory_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status history for an attendee with an invalid id")
	response := tstPerformGet("/api/rest/v1/attendees/lynx/status-history", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", url.Values{})
}

func TestStatusHistory_Nonexistent(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status history for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/status-history", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

// --- status changes ---

func TestStatusChange_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts a status change for an invalid attendee id")
	body := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "stat40",
	}
	response := tstPerformPost("/api/rest/v1/attendees/tigress/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "attendee.id.invalid", "")
}

func TestStatusChange_Nonexistant(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.When("when an admin attempts a status change for an attendee that does not exist")
	body := status.StatusChangeDto{
		Status:  status.Approved,
		Comment: "stat41",
	}
	response := tstPerformPost("/api/rest/v1/attendees/444/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", "")

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func TestStatusChange_InvalidBodySyntax(t *testing.T) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, "stat42", status.Approved)

	docs.When("when an admin prematurely tries to change their status but sends a syntactically invalid request body")
	response := tstPerformPost(loc+"/status", "{{-}}}}", tstValidAdminToken(t))

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "status.parse.error", url.Values{})

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func TestStatusChange_InvalidBodyValues(t *testing.T) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, "stat43", status.Approved)

	docs.When("when an admin tries to change the status to an invalid value")
	body := status.StatusChangeDto{
		Status:  "fluffy",
		Comment: "why isn't there a status fluffy?",
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusBadRequest, "status.data.invalid", url.Values{
		"status": []string{"status must be one of new,approved,partially paid,paid,checked in,waiting,cancelled,deleted"},
	})

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, status.Approved)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

// - anonymous is always denied -

func TestStatusChange_Anonymous_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			testname := fmt.Sprintf("TestStatusChange_Anonymous_%s_%s", oldStatus, newStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Anonymous_Deny(t, fmt.Sprintf("st%danon%d-", o, n), oldStatus, newStatus)
			})
		}
	}
}

// - other (without regdesk permission) is always denied -

func TestStatusChange_Other_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			testname := fmt.Sprintf("TestStatusChange_Other_%s_%s", oldStatus, newStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Other_Deny(t, fmt.Sprintf("st%dother%d-", o, n), oldStatus, newStatus)
			})
		}
	}
}

// - staff (without regdesk permission they are no different from regular attendees) is always denied for others -

func TestStatusChange_Staff_Other_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			testname := fmt.Sprintf("TestStatusChange_Staff_%s_%s", oldStatus, newStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Staff_Other_Deny(t, fmt.Sprintf("st%dstaff%d-", o, n), oldStatus, newStatus)
			})
		}
	}
}

// - self can do self cancellation from new and approved, but nothing else -
// (note that received payments come in as admin requests either from the payment service or from an admin, so those aren't self reported)

func TestStatusChange_Self_New_Cancelled(t *testing.T) {
	testcase := "st0self6-"
	tstStatusChange_Self_Allow(t, testcase,
		status.New, status.Cancelled,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

func TestStatusChange_Self_Approved_Cancelled(t *testing.T) {
	testcase := "st1self6-"
	tstStatusChange_Self_Allow(t, testcase,
		status.Approved, status.Cancelled,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "void unpaid dues on cancel")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

func TestStatusChange_Self_Waiting_Cancelled(t *testing.T) {
	testcase := "st5self6-"
	tstStatusChange_Self_Allow(t, testcase,
		status.Approved, status.Cancelled,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "void unpaid dues on cancel")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

// TODO refund logic by self cancellation date

func TestStatusChange_Self_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			if (oldStatus == status.New || oldStatus == status.Approved || oldStatus == status.Waiting) && newStatus == status.Cancelled {
				// see individual test cases above
			} else {
				testname := fmt.Sprintf("TestStatusChange_Self_%s_%s", oldStatus, newStatus)
				t.Run(testname, func(t *testing.T) {
					tstStatusChange_Self_Deny(t, fmt.Sprintf("st%dself%d-", o, n), oldStatus, newStatus)
				})
			}
		}
	}
}

// TODO test self cancellation unavailable because there were payments, even if refunded

// - an attendee with regdesk permission can check fully paid people in, but can do nothing else -

func TestStatusChange_Regdesk_Paid_CheckedIn(t *testing.T) {
	testcase := "st3regdsk4-"
	tstStatusChange_Regdesk_Allow(t, testcase,
		status.Paid, status.CheckedIn,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Regdesk_NotCompletelyPaid_CheckedIn(t *testing.T) {
	tstStatusChange_Regdesk_Unavailable(t, "st3regdsk4a-",
		status.Paid, status.CheckedIn,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -42)}, // slight underpayment
		"status.unpaid.dues", "payment amount not sufficient",
	)
}

func TestStatusChange_Regdesk_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			if (oldStatus == status.New || oldStatus == status.Approved) && newStatus == status.Cancelled {
				// see normal user test cases above - everyone may self-cancel here
			} else if oldStatus == status.Paid && newStatus == status.CheckedIn {
				// see individual test case above
			} else {
				testname := fmt.Sprintf("TestStatusChange_Regdesk_%s_%s", oldStatus, newStatus)
				t.Run(testname, func(t *testing.T) {
					tstStatusChange_Regdesk_Deny(t, fmt.Sprintf("st%dregdsk%d-", o, n), oldStatus, newStatus)
				})
			}
		}
	}
}

// - admins can make any available status change, so this tests availability conditions, mails sent and payment bookings
//   in all these cases -

func TestStatusChange_Admin_Same_Same(t *testing.T) {
	for b, bothStatus := range config.AllowedStatusValues() {
		testname := fmt.Sprintf("TestStatusChange_Admin_%s_%s", bothStatus, bothStatus)
		t.Run(testname, func(t *testing.T) {
			tstStatusChange_Admin_Unavailable(t, fmt.Sprintf("st%dadm%d-", b, b),
				bothStatus, bothStatus,
				nil,
				"status.unchanged.invalid", "old and new status are the same")
		})
	}
}

func TestStatusChange_Admin_New_Approved(t *testing.T) {
	testcase := "st0adm1-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.New, status.Approved,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_New_Waiting(t *testing.T) {
	testcase := "st0adm5-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.New, status.Waiting,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_New_Cancelled(t *testing.T) {
	testcase := "st0adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.New, status.Cancelled,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

func TestStatusChange_Admin_New_Deleted(t *testing.T) {
	testcase := "st0adm7-"
	tstStatusChange_Admin_Allow_DeletedCanReregister(t, testcase,
		status.New,
		nil,
		[]paymentservice.Transaction{},
	)
}

func TestStatusChange_Admin_New_Any(t *testing.T) {
	for n, targetStatus := range config.AllowedStatusValues() {
		if targetStatus == status.PartiallyPaid || targetStatus == status.Paid || targetStatus == status.CheckedIn {
			testname := fmt.Sprintf("TestStatusChange_Admin_%s_%s", status.New, targetStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Admin_Unavailable(t, fmt.Sprintf("st%dadm%d-", 0, n),
					status.New, targetStatus,
					nil,
					"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
			})

		}
	}
}

func TestStatusChange_Admin_Approved_New(t *testing.T) {
	testcase := "st1adm0-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.New,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "remove dues balance - status changed to new")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_Approved_PartiallyPaid(t *testing.T) {
	testcase := "st1adm2-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 2040)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMailWithAmounts(testcase, status.PartiallyPaid, 234.60, 255, false)},
	)
}

func TestStatusChange_Admin_Approved_Paid_WithGraceAmount(t *testing.T) {
	testcase := "st1adm3-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.Paid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25400)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMailWithAmounts(testcase, status.Paid, 1, 255, false)},
	)
}

func TestStatusChange_Admin_Approved_CheckedIn(t *testing.T) {
	testcase := "st1adm4-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.CheckedIn,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25500)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Admin_Approved_Waiting(t *testing.T) {
	testcase := "st1adm5-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.Waiting,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "remove dues balance - status changed to waiting")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_Approved_Cancelled(t *testing.T) {
	testcase := "st1adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Approved, status.Cancelled,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "void unpaid dues on cancel")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

func TestStatusChange_Admin_Approved_Deleted(t *testing.T) {
	testcase := "st1adm7-"
	tstStatusChange_Admin_Allow_DeletedCanReregister(t, testcase,
		status.Approved,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "remove dues balance - status changed to deleted")},
	)
}

func TestStatusChange_Admin_PartiallyPaid_New(t *testing.T) {
	testcase := "st2adm0-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.PartiallyPaid, status.New,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_PartiallyPaid_Approved_OkButDoesNothing(t *testing.T) {
	testcase := "st2adm1-"
	tstStatusChange_Admin_Allow_WithStatusAutoProgress(t, testcase,
		status.PartiallyPaid, status.Approved, status.PartiallyPaid,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Admin_PartiallyPaid_Approved_OkAfterRefund(t *testing.T) {
	testcase := "st2adm1r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.PartiallyPaid, status.Approved,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -15500)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_PartiallyPaid_Paid(t *testing.T) {
	testcase := "st2adm3-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.PartiallyPaid, status.Paid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 10000)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid, false)},
	)
}

func TestStatusChange_Admin_PartiallyPaid_CheckedIn(t *testing.T) {
	testcase := "st2adm4-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.PartiallyPaid, status.CheckedIn,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 10000)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Admin_PartiallyPaid_Waiting(t *testing.T) {
	testcase := "st2adm5-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.PartiallyPaid, status.Waiting,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_PartiallyPaid_Waiting_OkAfterRefund(t *testing.T) {
	testcase := "st2adm5r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.PartiallyPaid, status.Waiting,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -15500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to waiting")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_PartiallyPaid_Cancelled(t *testing.T) {
	testcase := "st2adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.PartiallyPaid, status.Cancelled,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-10000, "void unpaid dues on cancel")},
		[]mailservice.MailSendDto{tstNewCancelMail(testcase, testcase, 155)},
	)
}

func TestStatusChange_Admin_PartiallyPaid_Deleted(t *testing.T) {
	testcase := "st2adm7-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.PartiallyPaid, status.Deleted,
		nil,
		"status.cannot.delete", "cannot delete attendee for legal reasons (there were payments or invoices)")
}

func TestStatusChange_Admin_Paid_New(t *testing.T) {
	testcase := "st3adm0-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Paid, status.New,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_Paid_New_OkAfterRefund(t *testing.T) {
	testcase := "st3adm0r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.New,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to new")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_Paid_Approved_OkButDoesNothing(t *testing.T) {
	testcase := "st3adm1-"
	tstStatusChange_Admin_Allow_WithStatusAutoProgress(t, testcase,
		status.Paid, status.Approved, status.Paid,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Admin_Paid_Approved_OkAfterRefund(t *testing.T) {
	testcase := "st3adm1r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.Approved,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_Paid_PartiallyPaid(t *testing.T) {
	testcase := "st3adm2-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -10000)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.PartiallyPaid, false)},
	)
}

func TestStatusChange_Admin_Paid_CheckedIn_HasNoGraceAmount(t *testing.T) {
	testcase := "st3adm4u-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Paid, status.CheckedIn,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -30)},
		"status.unpaid.dues", "payment amount not sufficient")
}

func TestStatusChange_Admin_Paid_CheckedIn(t *testing.T) {
	testcase := "st3adm4-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.CheckedIn,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{},
	)
}

func TestStatusChange_Admin_Paid_Waiting(t *testing.T) {
	testcase := "st3adm5-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Paid, status.Waiting,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_Paid_Waiting_OkAfterRefund(t *testing.T) {
	testcase := "st3adm5r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.Waiting,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to waiting")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_Paid_Cancelled(t *testing.T) {
	testcase := "st3adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Paid, status.Cancelled,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewCancelMail(testcase, testcase, 255)},
	)
}

func TestStatusChange_Admin_Paid_Deleted(t *testing.T) {
	testcase := "st3adm7-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Paid, status.Deleted,
		nil,
		"status.cannot.delete", "cannot delete attendee for legal reasons (there were payments or invoices)")
}

func TestStatusChange_Admin_CheckedIn_New(t *testing.T) {
	testcase := "st4adm0-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.CheckedIn, status.New,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_CheckedIn_New_OkAfterRefund(t *testing.T) {
	testcase := "st4adm0r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.New,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to new")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_Approved_OkButLeadsToPaid(t *testing.T) {
	testcase := "st4adm1-"
	tstStatusChange_Admin_Allow_WithStatusAutoProgress(t, testcase,
		status.CheckedIn, status.Approved, status.Paid,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_Approved_OkAfterRefund(t *testing.T) {
	testcase := "st4adm1r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.Approved,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_PartiallyPaid(t *testing.T) {
	testcase := "st4adm2-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -10000)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.PartiallyPaid, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_Paid(t *testing.T) {
	testcase := "st4adm3-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.Paid,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_Waiting(t *testing.T) {
	testcase := "st4adm5-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.CheckedIn, status.Waiting,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_CheckedIn_Waiting_OkAfterRefund(t *testing.T) {
	testcase := "st4adm5r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.Waiting,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to waiting")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_CheckedIn_Cancelled(t *testing.T) {
	testcase := "st4adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.CheckedIn, status.Cancelled,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewCancelMail(testcase, testcase, 255)},
	)
}

func TestStatusChange_Admin_CheckedIn_Deleted(t *testing.T) {
	testcase := "st4adm7-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.CheckedIn, status.Deleted,
		nil,
		"status.cannot.delete", "cannot delete attendee for legal reasons (there were payments or invoices)")
}

func TestStatusChange_Admin_Waiting_New(t *testing.T) {
	testcase := "st5adm0-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Waiting, status.New,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_Waiting_Approved(t *testing.T) {
	testcase := "st5adm1-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Waiting, status.Approved,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_Waiting_PartiallyPaid(t *testing.T) {
	testcase := "st5adm2-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Waiting, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -10000)},
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Waiting_Paid(t *testing.T) {
	testcase := "st5adm3-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Waiting, status.Paid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25400)},
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Waiting_CheckedIn(t *testing.T) {
	testcase := "st5adm4-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Waiting, status.CheckedIn,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25500)},
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Waiting_Cancelled(t *testing.T) {
	testcase := "st5adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Waiting, status.Cancelled,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

func TestStatusChange_Admin_Waiting_Deleted(t *testing.T) {
	testcase := "st5adm7-"
	tstStatusChange_Admin_Allow_DeletedCanReregister(t, testcase,
		status.Waiting,
		nil,
		[]paymentservice.Transaction{},
	)
}

func TestStatusChange_Admin_Cancelled_New(t *testing.T) {
	testcase := "st6adm0-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.New,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_Cancelled_New_OkAfterRefund(t *testing.T) {
	testcase := "st6adm0r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Cancelled, status.New,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to new")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_Cancelled_Approved_OkButLeadsToPaid(t *testing.T) {
	testcase := "st6adm1-"
	tstStatusChange_Admin_Allow_WithStatusAutoProgress(t, testcase,
		status.Cancelled, status.Approved, status.Paid,
		[]paymentservice.Transaction{},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid, false)},
	)
}

func TestStatusChange_Admin_Cancelled_Approved_OkAfterRefund(t *testing.T) {
	testcase := "st6adm1r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Cancelled, status.Approved,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_Cancelled_PartiallyPaid(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st6adm2-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -10000)},
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Cancelled_Paid(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st6adm3-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.Paid,
		nil,
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Cancelled_CheckedIn(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st6adm4-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.CheckedIn,
		nil,
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Cancelled_Waiting(t *testing.T) {
	testcase := "st6adm5-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.Waiting,
		nil,
		"status.has.paid", "there is a non-zero payment balance, please use partially paid, or refund")
}

func TestStatusChange_Admin_Cancelled_Waiting_OkAfterRefund(t *testing.T) {
	testcase := "st6adm5r-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Cancelled, status.Waiting,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -25500)},
		[]paymentservice.Transaction{tstCreateMatcherTransaction(1, paymentservice.Due, -25500, "remove dues balance - status changed to waiting")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_Cancelled_Deleted(t *testing.T) {
	testcase := "st6adm7-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Cancelled, status.Deleted,
		nil,
		"status.cannot.delete", "cannot delete attendee for legal reasons (there were payments or invoices)")
}

func TestStatusChange_Admin_Deleted_New(t *testing.T) {
	testcase := "st7adm0-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Deleted, status.New,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.New, false)},
	)
}

func TestStatusChange_Admin_Deleted_Approved(t *testing.T) {
	testcase := "st7adm1-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Deleted, status.Approved,
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages")},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved, false)},
	)
}

func TestStatusChange_Admin_Deleted_PartiallyPaid(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st7adm2-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Deleted, status.PartiallyPaid,
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, -10000)},
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Deleted_Paid(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st7adm3-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Deleted, status.Paid,
		nil,
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Deleted_CheckedIn(t *testing.T) {
	// you cannot directly go back, since there may have been flag, package changes while cancelled which are not reflected in dues
	testcase := "st7adm4-"
	tstStatusChange_Admin_Unavailable(t, testcase,
		status.Deleted, status.CheckedIn,
		nil,
		"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
}

func TestStatusChange_Admin_Deleted_Waiting(t *testing.T) {
	testcase := "st7adm5-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Deleted, status.Waiting,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Waiting, false)},
	)
}

func TestStatusChange_Admin_Deleted_Cancelled(t *testing.T) {
	testcase := "st7adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		status.Deleted, status.Cancelled,
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Cancelled, false)},
	)
}

// ban check

func TestStatusChange_Admin_New_Approved_Banned(t *testing.T) {
	tstStatusChange_Admin_Unavailable_Banned(t, "st0adm1ban-", status.New, status.Approved)
}

func TestStatusChange_Admin_Cancelled_Approved_Banned(t *testing.T) {
	tstStatusChange_Admin_Unavailable_Banned(t, "st6adm1ban-", status.Cancelled, status.Approved)
}

func TestStatusChange_Admin_Deleted_Approved_Banned(t *testing.T) {
	tstStatusChange_Admin_Unavailable_Banned(t, "st7adm1ban-", status.Deleted, status.Approved)
}

func TestStatusChange_Admin_New_Approved_Banned_WithSkip(t *testing.T) {
	tstStatusChange_Admin_Allow_Banned_WithSkip(t, "st0adm1bsk-", status.New, status.Approved)
}

func TestStatusChange_Admin_Cancelled_Approved_WithSkip(t *testing.T) {
	tstStatusChange_Admin_Allow_Banned_WithSkip(t, "st6adm1bsk-", status.Cancelled, status.Approved)
}

func TestStatusChange_Admin_Deleted_Approved_WithSkip(t *testing.T) {
	tstStatusChange_Admin_Allow_Banned_WithSkip(t, "st7adm1bsk-", status.Deleted, status.Approved)
}

// TODO transition to cancelled and deleted with more complicated dues / payment histories

// TODO guest handling

// -- resend status mail

func TestResendStatusMail_Anonymous_Deny(t *testing.T) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, "stml1anon", status.Approved)

	docs.When("when an anonymous user requests the status mail to be resent")
	response := tstPerformPostNoBody(loc+"/status/resend", tstNoToken())

	docs.Then("then the request is denied with the appropriate error")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func TestResendStatusMail_User_Deny(t *testing.T) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status approved")
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, "stml1user", status.Approved)

	docs.When("when they request their last status mail to be resent")
	response := tstPerformPostNoBody(loc+"/status/resend", tstValidUserToken(t, att.Id))

	docs.Then("then the request is denied with the appropriate error (this is not yet supported for any user)")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not authorized for this operation - the attempt has been logged")

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func TestResendStatusMail_Admin_NoMails(t *testing.T) {
	for n, targetStatus := range config.AllowedStatusValues() {
		if targetStatus == status.Deleted || targetStatus == status.CheckedIn || targetStatus == status.New {
			testname := fmt.Sprintf("TestResendStatusMail_Admin_NoMails_%s", targetStatus)
			t.Run(testname, func(t *testing.T) {
				tstResendStatusMail_Admin(t, fmt.Sprintf("stml%dadm", n),
					targetStatus,
					nil)
			})

		}
	}
}

func TestResendStatusMail_Admin_WithMail(t *testing.T) {
	for n, targetStatus := range config.AllowedStatusValues() {
		if targetStatus != status.Deleted && targetStatus != status.CheckedIn && targetStatus != status.New {
			testname := fmt.Sprintf("TestResendStatusMail_Admin_WithMail_%s", targetStatus)
			t.Run(testname, func(t *testing.T) {
				testcase := fmt.Sprintf("stml%dadm", n)
				statusMail := tstNewStatusMail(testcase, targetStatus, false)
				if targetStatus == status.Cancelled {
					statusMail.Variables["reason"] = "change to cancelled"
					statusMail.Variables["total_dues"] = "EUR 255.00" // correct here because we cancelled after paid - means: no refund
				}
				tstResendStatusMail_Admin(t, testcase,
					targetStatus,
					[]mailservice.MailSendDto{statusMail})
			})
		}
	}
}

// --- detail implementations for status mail resend tests ---

func tstResendStatusMail_Admin(t *testing.T, testcase string,
	theStatus status.Status,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(theStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, theStatus)
	// we (ab)use the payments-changed webhook to set all the cached attendee fields
	// (but note this may also change status and trigger mails)
	webhookResponse := tstPerformPost(loc+"/payments-changed", "", tstValidApiToken())
	require.True(t, http.StatusAccepted == webhookResponse.status || http.StatusNoContent == webhookResponse.status)
	// so now reset mail recording
	mailMock.Reset()

	docs.When("when an admin requests their last status mail to be resent")
	response := tstPerformPostNoBody(loc+"/status/resend", tstValidAdminToken(t))

	docs.Then("then the request is successful")
	require.Equal(t, http.StatusNoContent, response.status)

	docs.Then("and the appropriate email message was sent via the mail service")
	tstRequireMailRequests(t, expectedMailRequests)
}

// --- detail implementations for the status change tests ---

func tstStatusChange_Anonymous_Deny(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when an anonymous user tries to change the status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "you must be logged in for this operation")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Self_Deny(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they try to change the status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidUserToken(t, att.Id))

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not allowed to make this status transition - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Self_Unavailable(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they prematurely try to change their own status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidStaffToken(t, att.Id))

	docs.Then("then the request fails as conflict (409) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusConflict, "", "")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Self_Allow(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
	expectedTransactions []paymentservice.Transaction,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	token := tstValidStaffToken(t, 1)
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they change their own status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), token)

	docs.Then("then the request is successful and the status change has been done")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, newStatus)

	docs.Then("and the appropriate dues were booked in the payment service")
	require.Equal(t, len(expectedTransactions), len(paymentMock.Recording()))
	for i, expected := range expectedTransactions {
		actual := paymentMock.Recording()[i]
		require.EqualValues(t, expected, actual)
	}

	docs.Then("and the appropriate email messages were sent via the mail service")
	tstRequireMailRequests(t, expectedMailRequests)
}

func tstStatusChange_Other_Deny(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus) + " and a second user")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	token2 := tstValidUserToken(t, 101)

	docs.When("when the second user tries to change the first attendee's status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), token2)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not allowed to make this status transition - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Regdesk_Deny(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus) + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee tries to change the first attendee's status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not allowed to make this status transition - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Regdesk_Unavailable(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	message string, details string,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus) + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
	}
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee prematurely tries to change the first attendee's status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Then("then the request fails as conflict (409) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusConflict, message, details)

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Regdesk_Allow(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
	expectedTransactions []paymentservice.Transaction,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus) + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee changes the first attendee's status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Then("then the request is successful and the status change has been done")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, newStatus)

	docs.Then("and the appropriate dues were booked in the payment service")
	require.Equal(t, len(expectedTransactions), len(paymentMock.Recording()))
	for i, expected := range expectedTransactions {
		actual := paymentMock.Recording()[i]
		require.EqualValues(t, expected, actual)
	}

	docs.Then("and the appropriate email messages were sent via the mail service")
	tstRequireMailRequests(t, expectedMailRequests)
}

func tstStatusChange_Staff_Other_Deny(t *testing.T, testcase string, oldStatus status.Status, newStatus status.Status) {
	tstSetup(false, true, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus) + " and a second user who is staff")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	token := tstValidStaffToken(t, 202)

	docs.When("when the staffer tries to change the first attendee's status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned, because staff have no special privileges for status changes")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not allowed to make this status transition - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

// admins never get deny (403), but they can get "not possible right now" (409)

func tstStatusChange_Admin_Unavailable(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	message string, details string) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
	}

	docs.When("when an admin prematurely tries to change the status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request fails as conflict (409) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusConflict, message, details)

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Admin_Unavailable_Banned(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there is a ban rule")
	ban := tstBuildValidBanRule(testcase)
	ban.NicknamePattern = "^.*cheetah$"
	banResponse := tstPerformPost("/api/rest/v1/bans", tstRenderJson(ban), tstValidAdminToken(t))
	require.Equal(t, http.StatusCreated, banResponse.status, "failed to create ban rule")

	docs.Given("given an attendee in status " + string(oldStatus) + " who matches the rule")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when an admin tries to change the status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request fails as conflict (409) due to the ban")
	tstRequireErrorResponse(t, response, http.StatusConflict,
		"status.ban.match",
		"this attendee matches a ban rule and cannot be approved, please review and either cancel or set the skip_ban_check admin flag to allow approval")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Admin_Allow_Banned_WithSkip(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
) {
	tstSetup(true, false, true)
	defer tstShutdown()

	docs.Given("given there is a ban rule")
	ban := tstBuildValidBanRule(testcase)
	ban.NicknamePattern = "^.*cheetah$"
	banResponse := tstPerformPost("/api/rest/v1/bans", tstRenderJson(ban), tstValidAdminToken(t))
	require.Equal(t, http.StatusCreated, banResponse.status, "failed to create ban rule")

	docs.Given("given an attendee in status " + string(oldStatus) + " who matches the rule")
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, "approved")
	// manually progress from approved to oldStatus so no payments get created
	_ = database.GetRepository().AddStatusChange(context.Background(), tstCreateStatusChange(att.Id, oldStatus))

	docs.Given("given the admin flag skip_ban_check has been set by an admin")
	adminInfoBody := admin.AdminInfoDto{
		Flags: "skip_ban_check",
	}
	adminInfoResponse := tstPerformPut(loc+"/admin", tstRenderJson(adminInfoBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, adminInfoResponse.status, "unexpected http response status")

	docs.When("when an admin tries to change the status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request is successful and the status change has been done")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, newStatus)
}

func tstStatusChange_Admin_Allow_WithStatusAutoProgress(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status, targetStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedTransactions []paymentservice.Transaction,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
	}
	// may have sent mails, so let's reset the mail recording here
	mailMock.Reset()

	docs.When("when an admin changes their status to " + string(newStatus))
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request is successful and their status has been set to " + string(targetStatus))
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, targetStatus)

	docs.Then("and the appropriate dues were booked in the payment service")
	tstRequireTransactions(t, expectedTransactions)

	docs.Then("and the appropriate email messages were sent via the mail service")
	tstRequireMailRequests(t, expectedMailRequests)
}

func tstStatusChange_Admin_Allow(t *testing.T, testcase string,
	oldStatus status.Status, newStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedTransactions []paymentservice.Transaction,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstStatusChange_Admin_Allow_WithStatusAutoProgress(t, testcase,
		oldStatus, newStatus, newStatus,
		injectExtraTransactions, expectedTransactions, expectedMailRequests,
	)
}

func tstStatusChange_Admin_Allow_DeletedCanReregister(t *testing.T, testcase string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedTransactions []paymentservice.Transaction,
) {
	tstSetup(false, false, true)
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
	}

	docs.When("when an admin changes their status to deleted")
	body := status.StatusChangeDto{
		Status:  status.Deleted,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request is successful and the status change to deleted has been done")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, status.Deleted)

	docs.Then("and the appropriate dues were booked in the payment service")
	tstRequireTransactions(t, expectedTransactions)

	docs.Then("and no email messages were sent via the mail service (deleted does not get emails)")
	tstRequireMailRequests(t, nil)

	docs.Then("and the same user can successfully register again after the deletion")
	_, _ = tstRegisterAttendeeAndTransitionToStatus(t, testcase, "new")
}

// TODO test unbook unpaid dues on cancel (but not paid dues!), in order of invoicing (don't forget negative dues in history)

// TODO test downstream errors (502) by simulating errors in payment and mail service

// helper functions

func tstRequireAttendeeStatus(t *testing.T, expected status.Status, responseBody string) {
	statusDto := status.StatusDto{}
	tstParseJson(responseBody, &statusDto)

	expectedStatusDto := status.StatusDto{
		Status: expected,
	}
	require.EqualValues(t, expectedStatusDto, statusDto, "status did not match expected value")
}

func tstRegisterRegdeskAttendee(t *testing.T, testcase string) string {
	token := tstValidUserToken(t, 101)

	loc2, _ := tstRegisterAttendeeWithToken(t, testcase+"2nd", token)
	permBody := admin.AdminInfoDto{
		Permissions: "regdesk",
	}
	permissionResponse := tstPerformPut(loc2+"/admin", tstRenderJson(permBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, permissionResponse.status)

	return token
}

func tstUpdateCache(ctx context.Context, attid uint, du int64, pa int64, dd string) {
	a, _ := database.GetRepository().GetAttendeeById(ctx, attid)
	a.CacheTotalDues = du
	a.CachePaymentBalance = pa
	a.CacheDueDate = dd
	_ = database.GetRepository().UpdateAttendee(ctx, a)
}

func tstUpdateCacheRelative(ctx context.Context, attid uint, du int64, pa int64, dd string) {
	a, _ := database.GetRepository().GetAttendeeById(ctx, attid)
	a.CacheTotalDues = a.CacheTotalDues + du
	a.CachePaymentBalance = a.CachePaymentBalance + pa
	a.CacheDueDate = dd
	_ = database.GetRepository().UpdateAttendee(ctx, a)
}

func tstRegisterAttendeeAndTransitionToStatus(t *testing.T, testcase string, targetStatus status.Status) (location string, att attendee.AttendeeDto) {
	// this works in all configurations, and for status changes, it makes no difference if a user is staff
	token := tstValidStaffToken(t, 1)

	location, att = tstRegisterAttendeeWithToken(t, testcase, token)
	if targetStatus == status.New {
		return
	}

	ctx := context.Background()
	attid := att.Id

	// waiting
	if targetStatus == status.Waiting {
		_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.Waiting))
		return
	}

	// approved
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.Approved))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Due, 25500))
	tstUpdateCache(ctx, attid, 25500, 0, "2022-12-22")
	if targetStatus == status.Approved {
		return
	}

	if targetStatus == status.Deleted {
		_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.Deleted))
		_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Due, -25500))
		tstUpdateCache(ctx, attid, 0, 0, "2022-12-22")
		return
	}

	// partially paid
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.PartiallyPaid))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Payment, 15500))
	tstUpdateCache(ctx, attid, 25500, 15500, "2022-12-22")
	if targetStatus == status.PartiallyPaid {
		return
	}

	// paid
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.Paid))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Payment, 10000))
	tstUpdateCache(ctx, attid, 25500, 25500, "2022-12-22")
	if targetStatus == status.Paid {
		return
	}

	// checked in
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.CheckedIn))
	if targetStatus == status.CheckedIn {
		return
	}

	// cancelled
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, status.Cancelled))
	if targetStatus == status.Cancelled {
		return
	}

	// invalid status - error in test code
	t.FailNow()
	return
}

func tstCreateStatusChange(attid uint, status status.Status) *entity.StatusChange {
	return &entity.StatusChange{
		AttendeeId: attid,
		Status:     status,
		Comments:   fmt.Sprintf("change to %s", status),
	}
}

func tstCreateTransaction(attid uint, ty paymentservice.TransactionType, amount int64) paymentservice.Transaction {
	method := paymentservice.Internal
	dueDate := "2022-12-22"
	if ty == paymentservice.Payment {
		method = paymentservice.Credit
		dueDate = "1999-12-31"
	}
	return paymentservice.Transaction{
		TransactionIdentifier: "1234-1234abc",
		DebitorID:             attid,
		TransactionType:       ty,
		Method:                method,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: amount,
			VatRate:   19,
		},
		Status:        paymentservice.Valid,
		EffectiveDate: dueDate,
		DueDate:       dueDate,
		StatusHistory: nil,
	}
}

func tstCreateMatcherTransaction(attid uint, ty paymentservice.TransactionType, amount int64, comment string) paymentservice.Transaction {
	method := paymentservice.Internal
	reason := ""
	if ty == paymentservice.Payment {
		method = paymentservice.Credit
	} else {
		reason = tstGuessDuesReason(amount, comment)
	}
	return paymentservice.Transaction{
		TransactionIdentifier: "",
		DebitorID:             attid,
		TransactionType:       ty,
		Method:                method,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: amount,
			VatRate:   19,
		},
		Status:        paymentservice.Valid,
		EffectiveDate: "2022-12-08",
		DueDate:       "2022-12-22",
		StatusHistory: nil, // TODO
		Comment:       comment,
		Reason:        reason,
	}
}

func tstVerifyStatus(t *testing.T, loc string, expectedStatus status.Status) {
	response := tstPerformGet(loc+"/status", tstValidAdminToken(t))
	require.Equal(t, http.StatusOK, response.status)
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)
	require.Equal(t, expectedStatus, statusDto.Status)
}
