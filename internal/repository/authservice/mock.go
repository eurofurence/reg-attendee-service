package authservice

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
)

type Mock interface {
	AuthService

	Reset()
	Enable()
	Recording() []string
	SimulateGetError(err error)
	SetupResponse(idToken string, acToken string, response UserInfoResponse)
}

type MockImpl struct {
	responses        map[string]UserInfoResponse
	recording        []string
	simulateGetError error
	simulatorEnabled bool
}

var (
	_ AuthService = (*MockImpl)(nil)
	_ Mock        = (*MockImpl)(nil)
)

func newMock() Mock {
	return &MockImpl{
		responses: make(map[string]UserInfoResponse),
		recording: make([]string, 0),
	}
}

func (m *MockImpl) UserInfo(ctx context.Context) (UserInfoResponse, error) {
	if m.simulatorEnabled {
		key := fmt.Sprintf("userinfo %s %s", ctxvalues.IdToken(ctx), ctxvalues.AccessToken(ctx))
		m.recording = append(m.recording, key)

		if m.simulateGetError != nil {
			return UserInfoResponse{}, m.simulateGetError
		}
		response, ok := m.responses[key]
		if !ok {
			return UserInfoResponse{}, UnauthorizedError
		}

		return response, nil
	} else {
		return UserInfoResponse{}, DownstreamError
	}
}

func (m *MockImpl) IsEnabled() bool {
	return m.simulatorEnabled
}

// only used in tests

func (m *MockImpl) Reset() {
	m.recording = make([]string, 0)
	m.simulateGetError = nil
	m.simulatorEnabled = false
}

func (m *MockImpl) Enable() {
	m.simulatorEnabled = true
}

func (m *MockImpl) Recording() []string {
	return m.recording
}

func (m *MockImpl) SimulateGetError(err error) {
	m.simulateGetError = err
}

func (m *MockImpl) SetupResponse(idToken string, acToken string, response UserInfoResponse) {
	key := fmt.Sprintf("userinfo %s %s", idToken, acToken)
	m.responses[key] = response
}
