package producer

import (
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"os"
	"rexis/rexis-go-attendee/internal/entity"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/web"
	"rexis/rexis-go-attendee/web/attendeectl"
	"testing"
)

var (
	ts *httptest.Server
)

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
	tstSetupConfig()
	tstSetupServiceMocks()
	tstSetupHttpTestServer()
}

func tstSetupConfig() {
	yaml := "" +
		`choices:
  flags:
    hc:
      description: 'blah'
    anon:
      description: 'blah'
    ev:  
      description: 'blah'
  packages:
    room-none:
      description: 'blah'
    attendance:
      description: 'blah'
    stage:
      description: 'blah'
    sponsor:
      description: 'blah'
    sponsor2:
      description: 'blah'
  options:
    art:
      description: 'blah'
    anim:
      description: 'blah'
    music:
      description: 'blah'
    suit:
      description: 'blah'
`
	err := config.InitializeConfiguration(yaml)
	if err != nil {
		os.Exit(1)
	}
}

func tstSetupHttpTestServer() {
	router := web.CreateRouter()
	ts = httptest.NewServer(router)
}

func tstShutdown() {
	ts.Close()
}

type MockAttendeeService struct {
	mock.Mock
}

func (s *MockAttendeeService) NewAttendee() *entity.Attendee {
	return &entity.Attendee{}
}

func (s *MockAttendeeService) RegisterNewAttendee(attendee *entity.Attendee) (uint, error) {
	// TODO use mock to verify data for contract tests
	return 1, nil
}

func (s *MockAttendeeService) GetAttendee(id uint) (*entity.Attendee, error) {
	// TODO when writing a contract test, put matching response data here
	return &entity.Attendee{}, nil
}

func (s *MockAttendeeService) UpdateAttendee(attendee *entity.Attendee) error {
	return nil
}

func tstSetupServiceMocks() {
	attendeectl.OverrideAttendeeService(&MockAttendeeService{})
}
