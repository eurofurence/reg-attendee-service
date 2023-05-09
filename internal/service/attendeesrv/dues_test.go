package attendeesrv

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBalances_NoTransactions(t *testing.T) {
	docs.Description("balance and due date correct for: no transactions")
	txs := []paymentservice.Transaction{}
	tstBalancesTestcase(t, txs, 0, 0, 0, "")
}

func TestBalances_OneDuesTransaction(t *testing.T) {
	docs.Description("balance and due date correct for: one dues transaction")
	txs := []paymentservice.Transaction{
		tstTx(paymentservice.Due, paymentservice.Valid, 12000, "2023-01-24", "2023-02-05"),
	}
	tstBalancesTestcase(t, txs, 12000, 0, 0, "2023-02-05")
}

func TestBalances_TwoDuesTransaction(t *testing.T) {
	docs.Description("balance and due date correct for: two dues transaction")
	txs := []paymentservice.Transaction{
		tstTx(paymentservice.Due, paymentservice.Valid, 12000, "2023-01-24", "2023-02-05"),
		tstTx(paymentservice.Due, paymentservice.Valid, 10000, "2023-01-31", "2023-02-18"),
	}
	tstBalancesTestcase(t, txs, 22000, 0, 0, "2023-02-05")
}

func TestBalances_FirstPaidTransaction(t *testing.T) {
	docs.Description("balance and due date correct for: one dues transaction")
	txs := []paymentservice.Transaction{
		tstTx(paymentservice.Due, paymentservice.Valid, 12000, "2023-01-24", "2023-02-05"),
		tstTx(paymentservice.Payment, paymentservice.Valid, 12000, "2023-01-25", "ignoreme"),
	}
	tstBalancesTestcase(t, txs, 12000, 12000, 0, "2023-02-05")
}

func TestBalances_FirstPaidSecondDueTransaction(t *testing.T) {
	docs.Description("balance and due date correct for: one dues transaction")
	txs := []paymentservice.Transaction{
		tstTx(paymentservice.Due, paymentservice.Valid, 12000, "2023-01-24", "2023-02-05"),
		tstTx(paymentservice.Due, paymentservice.Valid, 10000, "2023-01-31", "2023-02-18"),
		tstTx(paymentservice.Payment, paymentservice.Valid, 12000, "2023-01-25", "ignoreme"),
	}
	tstBalancesTestcase(t, txs, 22000, 12000, 0, "2023-02-18")
}

func TestBalances_LongerChain(t *testing.T) {
	docs.Description("balance and due date correct for: lots of transactions")
	txs := []paymentservice.Transaction{
		tstTx(paymentservice.Due, paymentservice.Valid, 12000, "2023-01-24", "2023-02-05"),
		tstTx(paymentservice.Due, paymentservice.Valid, 10000, "2023-01-31", "2023-02-13"),
		tstTx(paymentservice.Payment, paymentservice.Valid, 10000, "2023-01-25", "ignoreme"),
		tstTx(paymentservice.Due, paymentservice.Valid, 8000, "2023-02-04", "2023-02-15"), // payments not sufficient for this due amount, hence this sets the due date
		tstTx(paymentservice.Due, paymentservice.Valid, -8000, "2023-02-05", "2023-02-16"),
		tstTx(paymentservice.Due, paymentservice.Valid, 7000, "2023-02-10", "2023-03-01"),
		tstTx(paymentservice.Due, paymentservice.Valid, 11000, "2023-02-20", "2023-03-11"),
		tstTx(paymentservice.Payment, paymentservice.Valid, 15000, "2023-02-27", "ignoreme"),
		tstTx(paymentservice.Payment, paymentservice.Tentative, 15000, "2023-03-01", "ignoreme"),
	}
	tstBalancesTestcase(t, txs, 40000, 25000, 15000, "2023-02-15")
}

// --- helpers ---

func tstBalancesTestcase(t *testing.T,
	transactionHistory []paymentservice.Transaction,
	expectedValidDues int64, expectedValidPayments int64, expectedOpenPayments int64, expectedDueDate string) {

	cut := New().(*AttendeeServiceImplData)

	validDues, validPayments, openPayments, dueDate := cut.balances(transactionHistory)
	require.Equal(t, expectedValidDues, validDues)
	require.Equal(t, expectedValidPayments, validPayments)
	require.Equal(t, expectedOpenPayments, openPayments)
	require.Equal(t, expectedDueDate, dueDate)
}

func tstTx(ty paymentservice.TransactionType, st paymentservice.TransactionStatus, am int64, ed string, dd string) paymentservice.Transaction {
	return paymentservice.Transaction{
		DebitorID:             42,
		TransactionIdentifier: "EFTST-000042-1231-235959-1234",
		TransactionType:       ty,
		Method:                paymentservice.Credit,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: am,
			VatRate:   19.0,
		},
		Comment:       "ignored",
		Status:        st,
		EffectiveDate: ed,
		DueDate:       dd,
		CreationDate:  time.Now(),
		StatusHistory: nil,
	}
}
