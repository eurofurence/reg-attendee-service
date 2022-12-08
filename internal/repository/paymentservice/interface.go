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
	ID                    string
	DebitorID             uint              `json:"debitor_id"`
	TransactionIdentifier string            `json:"transaction_identifier"`
	Type                  TransactionType   `json:"transaciont_type"`
	Method                PaymentMethod     `json:"method"`
	Amount                Amount            `json:"amount"`
	Comment               string            `json:"comment"`
	Status                TransactionStatus `json:"status"`
	EffectiveDate         string            `json:"effective_date"`
	DueDate               string            `json:"due_date"`
	CreationDate          time.Time         `json:"creation_date"`
	Deletion              *Deletion
}

type TransactionResponse struct {
	Payload []Transaction
}
