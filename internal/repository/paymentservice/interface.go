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

type TransactionType int

const (
	Due TransactionType = iota
	Payment
)

type PaymentMethod int

const (
	Credit PaymentMethod = iota
	Paypal
	Transfer
	Internal
	Gift
)

type TransactionStatus int

const (
	Pending TransactionStatus = iota
	Tentative
	Valid
	Deleted
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
	DebitorID     uint // TODO either this is a string everywhere or a uint everywhere
	Type          TransactionType
	Method        PaymentMethod
	Amount        Amount
	Comment       string
	Status        TransactionStatus
	EffectiveDate string
	DueDate       time.Time
	Deletion      *Deletion
}
