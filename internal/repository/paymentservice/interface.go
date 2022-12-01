package paymentservice

import (
	"context"
	"errors"
	"time"
)

type PaymentService interface {
	GetTransactions(ctx context.Context, debitorId uint) ([]Transaction, error)
	AddTransaction(ctx context.Context, transaction Transaction) error
}

var (
	NoSuchDebitor404Error = errors.New("debitor id not found")
	DownstreamError       = errors.New("downstream unavailable - see log for details")
)

type TransactionType string

const (
	Due     TransactionType = "due"
	Payment TransactionType = "payment"
)

type PaymentMethod string

const (
	Credit   PaymentMethod = "credit"
	Paypal   PaymentMethod = "paypal"
	Transfer PaymentMethod = "transfer"
	Internal PaymentMethod = "internal"
	Gift     PaymentMethod = "gift"
)

type TransactionStatus string

const (
	Tentative TransactionStatus = "tentative"
	Pending   TransactionStatus = "pending"
	Valid     TransactionStatus = "valid"
	Deleted   TransactionStatus = "deleted"
)

type Deletion struct {
	PreviousStatus TransactionStatus
	Comment        string
	DeletedBy      string
	Date           time.Time
}

type Amount struct {
	Currency  string
	GrossCent int64
	VatRate   float64
}

type Transaction struct {
	ID            string
	DebitorID     uint
	Type          TransactionType
	Method        PaymentMethod
	Amount        Amount
	Comment       string
	Status        TransactionStatus
	EffectiveDate string
	DueDate       time.Time
	Deletion      *Deletion
}

type TransactionResponse struct {
	Payload []Transaction
}
