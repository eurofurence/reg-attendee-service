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
	Cash     PaymentMethod = "cash"
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

type StatusHistory struct {
	Status     TransactionStatus `json:"status"`
	Comment    string            `json:"comment"`
	ChangedBy  string            `json:"changed_by"`
	ChangeDate time.Time         `json:"change_date"`
}

type Amount struct {
	Currency  string  `json:"currency"`
	GrossCent int64   `json:"gross_cent"`
	VatRate   float64 `json:"vat_rate"`
}

type Transaction struct {
	DebitorID             uint              `json:"debitor_id"`
	TransactionIdentifier string            `json:"transaction_identifier"`
	TransactionType       TransactionType   `json:"transaction_type"`
	Method                PaymentMethod     `json:"method"`
	Amount                Amount            `json:"amount"`
	Comment               string            `json:"comment"`
	Status                TransactionStatus `json:"status"`
	EffectiveDate         string            `json:"effective_date"`
	DueDate               string            `json:"due_date"`
	CreationDate          time.Time         `json:"creation_date"`
	StatusHistory         []StatusHistory   `json:"status_history"`
}

type TransactionResponse struct {
	Payload []Transaction
}
