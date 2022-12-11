package mailservice

import "context"

type Mock interface {
	MailService

	Reset()
	Recording() []MailSendDto
	SimulateError(err error)
}

type MockImpl struct {
	recording     []MailSendDto
	simulateError error
}

var (
	_ MailService = (*MockImpl)(nil)
	_ Mock        = (*MockImpl)(nil)
)

func newMock() Mock {
	return &MockImpl{
		recording: make([]MailSendDto, 0),
	}
}

func (m *MockImpl) SendEmail(ctx context.Context, request MailSendDto) error {
	if m.simulateError != nil {
		return m.simulateError
	}
	m.recording = append(m.recording, request)
	return nil
}

// only used in tests

func (m *MockImpl) Reset() {
	m.recording = make([]MailSendDto, 0)
	m.simulateError = nil
}

func (m *MockImpl) Recording() []MailSendDto {
	return m.recording
}

func (m *MockImpl) SimulateError(err error) {
	m.simulateError = err
}
