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
	"strconv"
	"testing"
	"time"
)

// -------------------------------------------
// acceptance tests for the status subresource
// -------------------------------------------

// -- read status

func TestStatus_AnonDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	attendeeLocation, _ := tstRegisterAttendee(t, "stat1-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status")
	response := tstPerformGet(attendeeLocation+"/status", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")
}

func TestStatus_UserDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given two existing attendees")
	_, attendee1 := tstRegisterAttendee(t, "stat2a-")
	location2, _ := tstRegisterAttendee(t, "stat2b-")

	docs.Given("given the first attendee logs in and is a regular user")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access somebody else's status")
	response := tstPerformGet(location2+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatus_UserAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "stat3-")

	docs.Given("given the attendee logs in and is a regular user")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they access their own status")
	_ = tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model
}

func TestStatus_StaffDenyOther(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given two existing attendees")
	_, attendee1 := tstRegisterAttendee(t, "stat4a-")
	location2, _ := tstRegisterAttendee(t, "stat4b-")

	docs.Given("given the first attendee logs in and is staff")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access somebody else's status")
	response := tstPerformGet(location2+"/status", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatus_StaffAllowSelf(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee with no special privileges")
	location1, attendee1 := tstRegisterAttendee(t, "stat5-")

	docs.Given("given the attendee logs in")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they access their own status")
	_ = tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and status 'new' is returned")
	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model
}

func TestStatus_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat6-")

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they access the status for any attendee")
	response := tstPerformGet(location1+"/status", token)

	docs.Then("then the request is successful and the default status is returned")
	require.Equal(t, http.StatusOK, response.status, "unexpected http response status")
	tstRequireAttendeeStatus(t, "new", response.body)
}

func TestStatus_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
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
	tstSetup(tstDefaultConfigFile)
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
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, _ := tstRegisterAttendee(t, "stat20-")

	docs.Given("given an unauthenticated user")
	token := tstNoToken()

	docs.When("when they attempt to access the status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")
}

func TestStatusHistory_UserDeny(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an existing attendee")
	location1, attendee1 := tstRegisterAttendee(t, "stat21-")

	docs.Given("given a regular authenticated attendee")
	token := tstValidUserToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatusHistory_StaffDeny(t *testing.T) {
	docs.Given("given the configuration for staff registration")
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an authenticated staffer who has made a valid registration")
	location1, attendee1 := tstRegisterAttendee(t, "stat22-")
	token := tstValidStaffToken(t, attendee1.Id)

	docs.When("when they attempt to access their own or somebody else's status history")
	response := tstPerformGet(location1+"/status-history", token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")
}

func TestStatusHistory_AdminOk(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
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
			Status:    "new",
			Comment:   "registration",
		}},
	}
	require.EqualValues(t, expectedStatusHistory, statusHistoryDto, "status history did not match expected value")
}

func TestStatusHistory_InvalidId(t *testing.T) {
	docs.Given("given the configuration for standard registration")
	tstSetup(tstDefaultConfigFile)
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
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given a logged in admin")
	token := tstValidAdminToken(t)

	docs.When("when they try to access the status history for an attendee that does not exist")
	response := tstPerformGet("/api/rest/v1/attendees/42/status-history", token)

	docs.Then("then the request fails and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusNotFound, "attendee.id.notfound", url.Values{})
}

// --- status changes ---

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

// - staff (without regdesk permission they are no different from regular attendees) is always denied -

func TestStatusChange_Staff_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			testname := fmt.Sprintf("TestStatusChange_Staff_%s_%s", oldStatus, newStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Staff_Deny(t, fmt.Sprintf("st%dstaff%d-", o, n), oldStatus, newStatus)
			})
		}
	}
}

// - self can do self cancellation from new and approved, but nothing else -
// (note that received payments come in as admin requests either from the payment service or from an admin, so those aren't self reported)

func TestStatusChange_Self_New_Cancelled(t *testing.T) {
	tstStatusChange_Self_Allow(t, "st0self6-", "new", "cancelled")
	// TODO refund logic by self cancellation date
}

func TestStatusChange_Self_Approved_Cancelled(t *testing.T) {
	tstStatusChange_Self_Allow(t, "st1self6-", "approved", "cancelled")
}

func TestStatusChange_Self_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			if (oldStatus == "new" || oldStatus == "approved") && newStatus == "cancelled" {
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

// - an attendee with regdesk permission can check paid people in, but can do nothing else -

func TestStatusChange_Regdesk_Paid_CheckedIn(t *testing.T) {
	tstStatusChange_Regdesk_Allow(t, "st3regdsk4-", "paid", "checked in")
}

func TestStatusChange_Regdesk_Any_Any(t *testing.T) {
	for o, oldStatus := range config.AllowedStatusValues() {
		for n, newStatus := range config.AllowedStatusValues() {
			if oldStatus == "paid" && newStatus == "checked in" {
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
			tstStatusChange_Admin_Unavailable(t, fmt.Sprintf("st%dadm%d-", b, b), bothStatus, bothStatus,
				"status.unchanged.invalid", "old and new status are the same")
		})
	}
}

func TestStatusChange_Admin_New_Approved(t *testing.T) {
	testcase := "st0adm1-"
	tstStatusChange_Admin_Allow(t, testcase,
		"new", "approved",
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(25500, "dues adjustment due to change in status or selected packages")},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "approved")},
	)
}

func TestStatusChange_Admin_New_Cancelled(t *testing.T) {
	testcase := "st0adm5-"
	tstStatusChange_Admin_Allow(t, testcase,
		"new", "cancelled",
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "cancelled")},
	)
}

func TestStatusChange_Admin_New_Deleted(t *testing.T) {
	testcase := "st0adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		"new", "deleted",
		nil,
		[]paymentservice.Transaction{},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "deleted")},
	)
}

func TestStatusChange_Admin_New_Any(t *testing.T) {
	for n, targetStatus := range config.AllowedStatusValues() {
		if targetStatus == "partially paid" || targetStatus == "paid" || targetStatus == "checked in" {
			testname := fmt.Sprintf("TestStatusChange_Admin_%s_%s", "new", targetStatus)
			t.Run(testname, func(t *testing.T) {
				tstStatusChange_Admin_Unavailable(t, fmt.Sprintf("st%dadm%d-", 0, n), "new", targetStatus,
					"status.use.approved", "please change status to approved, this will automatically advance to (partially) paid as appropriate")
			})

		}
	}
}

func TestStatusChange_Admin_Approved_New(t *testing.T) {
	testcase := "st1adm0-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "new",
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "remove dues balance - status changed to new")},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "new")},
	)
}

func TestStatusChange_Admin_Approved_PartiallyPaid(t *testing.T) {
	testcase := "st1adm2-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "partially paid",
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 2040)},
		[]paymentservice.Transaction{},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "partially paid")},
	)
}

func TestStatusChange_Admin_Approved_Paid_WithGraceAmount(t *testing.T) {
	testcase := "st1adm3-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "paid",
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25400)},
		[]paymentservice.Transaction{},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "paid")},
	)
}

func TestStatusChange_Admin_Approved_CheckedIn(t *testing.T) {
	testcase := "st1adm4-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "checked in",
		[]paymentservice.Transaction{tstCreateTransaction(1, paymentservice.Payment, 25500)},
		[]paymentservice.Transaction{},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "checked in")},
	)
}

func TestStatusChange_Admin_Approved_Cancelled(t *testing.T) {
	testcase := "st1adm5-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "cancelled",
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "void unpaid dues on cancel")},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "cancelled")},
	)
}

func TestStatusChange_Admin_Approved_Deleted(t *testing.T) {
	testcase := "st1adm6-"
	tstStatusChange_Admin_Allow(t, testcase,
		"approved", "deleted",
		nil,
		[]paymentservice.Transaction{tstValidAttendeeDues(-25500, "remove dues balance - status changed to deleted")},
		[]mailservice.TemplateRequestDto{tstNewStatusMail(testcase, "deleted")},
	)
}

// ...

// TODO transitions with other payment states (so far we're only testing one example each)
// TODO transition to cancelled and deleted with more complicated dues / payment histories

// TODO ban check

// TODO guest handling

// --- detail implementations for the status change tests ---

func tstStatusChange_Anonymous_Deny(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when an anonymous user tries to change the status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstNoToken())

	docs.Then("then the request is denied as unauthenticated (401) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusUnauthorized, "auth.unauthorized", "missing Authorization header with bearer token")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Self_Deny(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they try to change the status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidUserToken(t, att.Id))

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Self_Unavailable(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they prematurely try to change the status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	_ = tstPerformPost(loc+"/status", tstRenderJson(body), tstValidUserToken(t, att.Id))

	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model

	docs.Then("then the request fails as conflict (409) and the appropriate error is returned")
	// TODO

	docs.Then("and the status is unchanged")
	// TODO

	docs.Then("and no dues or payment changes have been recorded")
	// TODO

	docs.Then("and no email messages have been sent")
	// TODO
}

func tstStatusChange_Self_Allow(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, att := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when they change their own status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	_ = tstPerformPost(loc+"/status", tstRenderJson(body), tstValidUserToken(t, att.Id))

	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model

	docs.Then("then the request is successful and the status change has been done")
	// TODO

	docs.Then("and the appropriate dues were booked in the payment-service")
	// TODO - pass in expected as parameter and record in mock

	docs.Then("and the appropriate email messages were sent via the mail-service")
	// TODO - pass in expected as parameter and record in mock
}

func tstStatusChange_Other_Deny(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus + " and a second user")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	_, att2 := tstRegisterAttendee(t, testcase+"second")

	docs.When("when the second user tries to change the first attendee's status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidUserToken(t, att2.Id))

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Regdesk_Deny(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee tries to change the first attendee's status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

func tstStatusChange_Regdesk_Unavailable(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee prematurely tries to change the first attendee's status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	_ = tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model

	docs.Then("then the request fails as conflict (409) and the appropriate error is returned")
	// TODO

	docs.Then("and the status is unchanged")
	// TODO

	docs.Then("and no dues or payment changes have been recorded")
	// TODO

	docs.Then("and no email messages have been sent")
	// TODO
}

func tstStatusChange_Regdesk_Allow(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus + " and a second attendee with the regdesk permission")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	regdeskUserToken := tstRegisterRegdeskAttendee(t, testcase)

	docs.When("when the regdesk attendee changes the first attendee's status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	_ = tstPerformPost(loc+"/status", tstRenderJson(body), regdeskUserToken)

	docs.Limitation("the current fixed-token security model cannot check which user is logged in. Once implemented this should be successful!")
	// TODO implement this part when working on security model

	docs.Then("then the request is successful and the status change has been done")
	// TODO

	docs.Then("and the appropriate dues were booked in the payment-service")
	// TODO - pass in expected as parameter and record in mock

	docs.Then("and the appropriate email messages were sent via the mail-service")
	// TODO - pass in expected as parameter and record in mock
}

func tstStatusChange_Staff_Deny(t *testing.T, testcase string, oldStatus string, newStatus string) {
	tstSetup(tstStaffregConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus + " and a second user who is staff")
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	_, att2 := tstRegisterAttendee(t, testcase+"second")
	token := tstValidStaffToken(t, att2.Id)

	docs.When("when the staffer tries to change the first attendee's status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), token)

	docs.Then("then the request is denied as unauthorized (403) and the appropriate error is returned")
	tstRequireErrorResponse(t, response, http.StatusForbidden, "auth.forbidden", "you are not unauthorized for this operation - the attempt has been logged")

	docs.Then("and the status is unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no dues or payment changes have been recorded")
	require.Empty(t, paymentMock.Recording())

	docs.Then("and no email messages have been sent")
	require.Empty(t, mailMock.Recording())
}

// admins never get deny (403), but they can get "not possible right now" (409)

func tstStatusChange_Admin_Unavailable(t *testing.T, testcase string, oldStatus string, newStatus string, message string, details string) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	docs.When("when an admin prematurely tries to change the status to " + newStatus)
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

func tstStatusChange_Admin_Allow(t *testing.T, testcase string,
	oldStatus string, newStatus string,
	injectExtraTransactions []paymentservice.Transaction,
	expectedTransactions []paymentservice.Transaction,
	expectedMailRequests []mailservice.TemplateRequestDto,
) {
	tstSetup(tstDefaultConfigFile)
	defer tstShutdown()

	docs.Given("given an attendee in status " + oldStatus)
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
	}

	docs.When("when an admin changes their status to " + newStatus)
	body := status.StatusChangeDto{
		Status:  newStatus,
		Comment: testcase,
	}
	response := tstPerformPost(loc+"/status", tstRenderJson(body), tstValidAdminToken(t))

	docs.Then("then the request is successful and the status change has been done")
	require.Equal(t, http.StatusNoContent, response.status)
	tstVerifyStatus(t, loc, newStatus)

	docs.Then("and the appropriate dues were booked in the payment service")
	require.EqualValues(t, expectedTransactions, paymentMock.Recording())

	docs.Then("and the appropriate email messages were sent via the mail service")
	require.Equal(t, len(expectedMailRequests), len(mailMock.Recording()))
	for i, expected := range expectedMailRequests {
		actual := mailMock.Recording()[i]
		require.Contains(t, actual.Email, expected.Email)
		actual.Email = expected.Email
		require.EqualValues(t, expected, actual)
	}
}

// TODO test unbook unpaid dues on cancel (but not paid dues!), in order of invoicing (don't forget negative dues in history)

// TODO test invalid values, attendee id, invalid body etc. with admin

// TODO test downstream errors (502) by simulating errors in payment and mail service

// helper functions

func tstRequireAttendeeStatus(t *testing.T, expected string, responseBody string) {
	statusDto := status.StatusDto{}
	tstParseJson(responseBody, &statusDto)

	expectedStatusDto := status.StatusDto{
		Status: expected,
	}
	require.EqualValues(t, expectedStatusDto, statusDto, "status did not match expected value")
}

func tstRegisterRegdeskAttendee(t *testing.T, testcase string) (token string) {
	loc2, att2 := tstRegisterAttendee(t, testcase+"second")
	permBody := admin.AdminInfoDto{
		Permissions: "regdesk",
	}
	permissionResponse := tstPerformPut(loc2+"/admin", tstRenderJson(permBody), tstValidAdminToken(t))
	require.Equal(t, http.StatusNoContent, permissionResponse.status)

	return tstValidUserToken(t, att2.Id)
}

func tstRegisterAttendeeAndTransitionToStatus(t *testing.T, testcase string, status string) (location string, att attendee.AttendeeDto) {
	location, att = tstRegisterAttendee(t, testcase)
	if status == "new" {
		return
	}

	ctx := context.Background()
	attid, _ := strconv.Atoi(att.Id)

	// approved
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "approved"))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Due, 25500))
	if status == "approved" {
		return
	}

	if status == "deleted" {
		_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "deleted"))
		return
	}

	// partially paid
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "partially paid"))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Payment, 15500))
	if status == "partially paid" {
		return
	}

	// paid
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "paid"))
	_ = paymentMock.InjectTransaction(ctx, tstCreateTransaction(attid, paymentservice.Payment, 10000))
	if status == "paid" {
		return
	}

	// checked in
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "checked in"))
	if status == "checked in" {
		return
	}

	// cancelled
	_ = database.GetRepository().AddStatusChange(ctx, tstCreateStatusChange(attid, "cancelled"))
	if status == "cancelled" {
		return
	}

	// invalid status - error in test code
	t.FailNow()
	return
}

func tstCreateStatusChange(attid int, status string) *entity.StatusChange {
	return &entity.StatusChange{
		AttendeeId: uint(attid),
		Status:     status,
	}
}

func tstCreateTransaction(attid int, ty paymentservice.TransactionType, amount int64) paymentservice.Transaction {
	method := paymentservice.Internal
	if ty == paymentservice.Payment {
		method = paymentservice.Credit
	}
	return paymentservice.Transaction{
		ID:        "1234-1234abc",
		DebitorID: uint(attid),
		Type:      ty,
		Method:    method,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: amount,
			VatRate:   19,
		},
		Status:        paymentservice.Valid,
		EffectiveDate: "1999-12-31",
		DueDate:       time.Now(),
		Deletion:      nil,
	}
}

func tstVerifyStatus(t *testing.T, loc string, expectedStatus string) {
	response := tstPerformGet(loc+"/status", tstValidAdminToken(t))
	require.Equal(t, http.StatusOK, response.status)
	statusDto := status.StatusDto{}
	tstParseJson(response.body, &statusDto)
	require.Equal(t, expectedStatus, statusDto.Status)
}
