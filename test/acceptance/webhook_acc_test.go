package acceptance

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// --- error cases

func TestPaymentsChangedWebhook_InvalidAttendeeId(t *testing.T) {
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.When("when the payments-changed webhook is invoked with an invalid attendee id")
	response := tstPerformPost("/api/rest/v1/attendees/helloworld/payments-changed", "", tstValidApiToken())

	docs.Then("then the request fails with the expected error")
	require.Equal(t, http.StatusBadRequest, response.status)
}

func TestPaymentsChangedWebhook_NonexistentAttendee(t *testing.T) {
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.When("when the payments-changed webhook is invoked with an attendee id that does not exist")
	response := tstPerformPost("/api/rest/v1/attendees/42/payments-changed", "", tstValidApiToken())

	docs.Then("then the request fails with the expected error")
	require.Equal(t, http.StatusNotFound, response.status)
}

// --- webhook invocations by non-admin or with wrong api key never work for any situation (just testing a few examples)

func TestPaymentsChangedWebhook_Anonymous_Declined(t *testing.T) {
	testcase := "pc1anon2-"
	tstStatusChange_Webhook_Decline(t, testcase,
		tstNoToken(),
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 2040),
		},
		http.StatusUnauthorized,
	)
}

func TestPaymentsChangedWebhook_User_Declined(t *testing.T) {
	testcase := "pc1user2-"
	tstStatusChange_Webhook_Decline(t, testcase,
		tstValidUserToken(t, 101),
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 2040),
		},
		http.StatusForbidden,
	)
}

func TestPaymentsChangedWebhook_Staff_Declined(t *testing.T) {
	testcase := "pc1staff2-"
	tstStatusChange_Webhook_Decline(t, testcase,
		tstValidStaffToken(t, 202),
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 2040),
		},
		http.StatusForbidden,
	)
}

// --- webhook invocations by admin and with api key (everything else cannot be a success case)

var subcaseAdmOrApi = []string{
	"adm",
	"api",
}
var subcaseAdmOrApiTokens = []string{
	valid_JWT_is_admin,
	valid_Api_Token_Matches_Test_Configuration_Files,
}

func TestPaymentsChangedWebhook_New_NoNewPayments(t *testing.T) {
	testcase := "pc0a0-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.New,
		nil,
	)
}

func TestPaymentsChangedWebhook_New_WithPayments(t *testing.T) {
	testcase := "pc0a0p-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.New,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 10000),
		},
	)
}

func TestPaymentsChangedWebhook_Approved_NoNewPayments(t *testing.T) {
	testcase := "pc1a1-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Approved,
		nil,
		status.Approved,
		nil,
	)
}

func TestPaymentsChangedWebhook_Approved_PartiallyPaid(t *testing.T) {
	testcase := "pc1a2-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 2040),
		},
		status.PartiallyPaid,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.PartiallyPaid)},
	)
}

func TestPaymentsChangedWebhook_Approved_Paid_WithGraceAmount(t *testing.T) {
	testcase := "pc1a3-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 25400),
		},
		status.Paid,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid)},
	)
}

func TestPaymentsChangedWebhook_Approved_Paid_Overpaid(t *testing.T) {
	testcase := "pc1a3o-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Approved,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 27000),
		},
		status.Paid,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid)},
	)
}

func TestPaymentsChangedWebhook_PartiallyPaid_NoNewPayments(t *testing.T) {
	testcase := "pc2a2-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.PartiallyPaid,
		nil,
		status.PartiallyPaid,
		nil,
	)
}

func TestPaymentsChangedWebhook_PartiallyPaid_Approved(t *testing.T) {
	testcase := "pc2a1-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.PartiallyPaid,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, -15500),
		},
		status.Approved,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved)},
	)
}

func TestPaymentsChangedWebhook_PartiallyPaid_PartialRefund(t *testing.T) {
	testcase := "pc2a2p-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.PartiallyPaid,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, -5500),
		},
		status.PartiallyPaid,
		nil,
	)
}

func TestPaymentsChangedWebhook_PartiallyPaid_Paid(t *testing.T) {
	testcase := "pc2a3-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.PartiallyPaid,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 10000),
		},
		status.Paid,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Paid)},
	)
}

func TestPaymentsChangedWebhook_Paid_NoNewPayments(t *testing.T) {
	testcase := "pc3a3-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Paid,
		nil,
		status.Paid,
		nil,
	)
}

func TestPaymentsChangedWebhook_Paid_Approved(t *testing.T) {
	testcase := "pc3a1-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Paid,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, -25500),
		},
		status.Approved,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.Approved)},
	)
}

func TestPaymentsChangedWebhook_Paid_PartiallyPaid(t *testing.T) {
	testcase := "pc3a2-"
	tstStatusChange_Webhook_Success(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Paid,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, -15500),
		},
		status.PartiallyPaid,
		[]mailservice.MailSendDto{tstNewStatusMail(testcase, status.PartiallyPaid)},
	)
}

func TestPaymentsChangedWebhook_CheckedIn_NoNewPayments(t *testing.T) {
	testcase := "pc4a4-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.CheckedIn,
		nil,
	)
}

func TestPaymentsChangedWebhook_CheckedIn_EvenWithRefundsIgnored(t *testing.T) {
	testcase := "pc4a4p-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.CheckedIn,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, -10000),
		},
	)
}

func TestPaymentsChangedWebhook_Cancelled_NoNewPayments(t *testing.T) {
	testcase := "pc5a5-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Cancelled,
		nil,
	)
}

func TestPaymentsChangedWebhook_Cancelled_WithPayments(t *testing.T) {
	testcase := "pc5a5p-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Cancelled,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 10000),
		},
	)
}

func TestPaymentsChangedWebhook_Deleted_NoNewPayments(t *testing.T) {
	testcase := "pc6a6-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Deleted,
		nil,
	)
}

func TestPaymentsChangedWebhook_Deleted_WithPayments(t *testing.T) {
	testcase := "pc6a6p-"
	tstStatusChange_Webhook_Ignored(t, testcase,
		subcaseAdmOrApi,
		subcaseAdmOrApiTokens,
		status.Deleted,
		[]paymentservice.Transaction{
			tstCreateTransaction(1, paymentservice.Payment, 10000),
		},
	)
}

// --- detail implementations for the status change tests ---

var subcaseNameMap = map[string]string{
	"adm": "Admin",
	"api": "Api_Key",
}

func tstStatusChange_Webhook_Success(t *testing.T, testcase string, subcases []string, tokens []string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedNewStatus status.Status,
	expectedMailRequests []mailservice.MailSendDto,
) {
	for i, subcase := range subcases {
		t.Run(subcaseNameMap[subcase], func(t2 *testing.T) {
			tstStatusChange_Webhook_Success_WithToken(t2, testcase+subcase, tokens[i],
				oldStatus, injectExtraTransactions, expectedNewStatus, expectedMailRequests)
		})
	}
}

func tstStatusChange_Webhook_Success_WithToken(t *testing.T, testcase string,
	token string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedNewStatus status.Status,
	expectedMailRequests []mailservice.MailSendDto,
) {
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	sum := 0.0
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
		sum += float64(tx.Amount.GrossCent) / 100.0
	}
	if sum != 0.0 {
		docs.Given(fmt.Sprintf("given extra payments of %.2f", sum))
	}

	docs.When("when the payments-changed webhook is invoked")
	response := tstPerformPost(loc+"/payments-changed", "", token)

	docs.Then("then the request is successfully processed (204)")
	require.Equal(t, http.StatusNoContent, response.status)

	docs.Then("and the resulting attendee status is " + string(expectedNewStatus) + " as expected")
	tstVerifyStatus(t, loc, expectedNewStatus)

	docs.Then("and no additional transactions were booked in the payment service")
	require.Equal(t, 0, len(paymentMock.Recording()))

	docs.Then("and the appropriate email messages were sent via the mail service")
	tstRequireMailRequests(t, expectedMailRequests)
}

func tstStatusChange_Webhook_Ignored(t *testing.T, testcase string, subcases []string, tokens []string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
) {
	for i, subcase := range subcases {
		t.Run(subcaseNameMap[subcase], func(t2 *testing.T) {
			tstStatusChange_Webhook_Ignored_WithToken(t2, testcase+subcase, tokens[i],
				oldStatus, injectExtraTransactions)
		})
	}
}

func tstStatusChange_Webhook_Ignored_WithToken(t *testing.T, testcase string,
	token string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
) {
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	sum := 0.0
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
		sum += float64(tx.Amount.GrossCent) / 100.0
	}
	if sum != 0.0 {
		docs.Given(fmt.Sprintf("given extra payments of %.2f", sum))
	}

	docs.When("when the payments-changed webhook is invoked")
	response := tstPerformPost(loc+"/payments-changed", "", token)

	docs.Then("then the request is successfully accepted (202), but no processing took place")
	require.Equal(t, http.StatusAccepted, response.status)

	docs.Then("and attendee status remains unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no additional transactions were booked in the payment service")
	require.Equal(t, 0, len(paymentMock.Recording()))

	docs.Then("and no email messages were sent via the mail service")
	require.Equal(t, 0, len(mailMock.Recording()))
}

func tstStatusChange_Webhook_Decline(t *testing.T, testcase string,
	token string,
	oldStatus status.Status,
	injectExtraTransactions []paymentservice.Transaction,
	expectedHttpStatus int,
) {
	tstSetup(tstConfigFile(false, false, true))
	defer tstShutdown()

	docs.Given("given an attendee in status " + string(oldStatus))
	loc, _ := tstRegisterAttendeeAndTransitionToStatus(t, testcase, oldStatus)

	sum := 0.0
	for _, tx := range injectExtraTransactions {
		_ = paymentMock.InjectTransaction(context.Background(), tx)
		sum += float64(tx.Amount.GrossCent) / 100.0
	}
	if sum != 0.0 {
		docs.Given(fmt.Sprintf("given extra payments of %.2f", sum))
	}

	docs.When("when the payments-changed webhook is invoked without proper authorization")
	response := tstPerformPost(loc+"/payments-changed", "", token)

	docs.Then("then the request is denied")
	require.Equal(t, expectedHttpStatus, response.status)

	docs.Then("and attendee status remains unchanged")
	tstVerifyStatus(t, loc, oldStatus)

	docs.Then("and no additional transactions were booked in the payment service")
	require.Equal(t, 0, len(paymentMock.Recording()))

	docs.Then("and no email messages were sent via the mail service")
	require.Equal(t, 0, len(mailMock.Recording()))
}
