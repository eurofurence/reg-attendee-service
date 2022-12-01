package paymentservice

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
)

type Mock interface {
	PaymentService

	InjectTransaction(ctx context.Context, transaction Transaction) error
	Reset()
	Recording() []Transaction
	SimulateGetError(err error)
	SimulateAddError(err error)
}

type MockImpl struct {
	data             map[uint][]Transaction
	recording        []Transaction
	simulateGetError error
	simulateAddError error
}

var (
	_ PaymentService = (*MockImpl)(nil)
	_ Mock           = (*MockImpl)(nil)
)

func newMock() Mock {
	return &MockImpl{
		data:      make(map[uint][]Transaction),
		recording: make([]Transaction, 0),
	}
}

func (m *MockImpl) GetTransactions(ctx context.Context, debitorId uint) ([]Transaction, error) {
	result := make([]Transaction, 0)
	if m.simulateGetError != nil {
		return result, m.simulateGetError
	}

	transactions, ok := m.data[debitorId]
	if !ok {
		return result, NoSuchDebitor404Error
	}

	return transactions, nil
}

func (m *MockImpl) AddTransaction(ctx context.Context, transaction Transaction) error {
	if m.simulateAddError != nil {
		return m.simulateAddError
	}

	_ = m.InjectTransaction(ctx, transaction)
	m.recording = append(m.recording, transaction)

	aulogging.Logger.Ctx(ctx).Info().Printf("add transaction debitor %d type %s status %s method %s for %0.2f %s", transaction.DebitorID, transaction.Type, transaction.Status, transaction.Method, float64(transaction.Amount.GrossCent)/100.0, transaction.Amount.Currency)

	return nil
}

// only used in tests

func (m *MockImpl) Reset() {
	m.recording = make([]Transaction, 0)
	m.simulateGetError = nil
	m.simulateAddError = nil
}

func (m *MockImpl) Recording() []Transaction {
	return m.recording
}

func (m *MockImpl) SimulateGetError(err error) {
	m.simulateGetError = err
}

func (m *MockImpl) SimulateAddError(err error) {
	m.simulateAddError = err
}

func (m *MockImpl) InjectTransaction(_ context.Context, transaction Transaction) error {
	existingTransactions, ok := m.data[transaction.DebitorID]
	if !ok {
		existingTransactions = make([]Transaction, 0)
	}

	transactions := append(existingTransactions, transaction)
	m.data[transaction.DebitorID] = transactions

	return nil
}
