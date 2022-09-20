package mailservice

import "context"

type Mock interface {
	MailService

	Reset()
	Recording() []TemplateRequestDto
	SimulateError(err error)
}

type MockImpl struct {
	recording     []TemplateRequestDto
	simulateError error
}

var (
	_ MailService = (*MockImpl)(nil)
	_ Mock        = (*MockImpl)(nil)
)

func NewMock() Mock {
	return &MockImpl{
		recording: make([]TemplateRequestDto, 0),
	}
}

func (m *MockImpl) SendEmail(ctx context.Context, request TemplateRequestDto) error {
	if m.simulateError != nil {
		return m.simulateError
	}
	m.recording = append(m.recording, request)
	return nil
}

// only used in tests

func (m *MockImpl) Reset() {
	m.recording = make([]TemplateRequestDto, 0)
	m.simulateError = nil
}

func (m *MockImpl) Recording() []TemplateRequestDto {
	return m.recording
}

func (m *MockImpl) SimulateError(err error) {
	m.simulateError = err
}
