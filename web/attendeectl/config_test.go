package attendeectl

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"os"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"testing"
)

// placing these here because they are package global

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
	tstSetupConfig()
	tstSetupServiceMocks()
}

func tstShutdown() {

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

type MockAttendeeService struct {
	mock.Mock
}

func (s *MockAttendeeService) NewAttendee() *entity.Attendee {
	return &entity.Attendee{}
}

func (s *MockAttendeeService) RegisterNewAttendee(attendee *entity.Attendee) (uint, error) {
	return 0, errors.New("some error, this is a mock")
}

func (s *MockAttendeeService) GetAttendee(id uint) (*entity.Attendee, error) {
	return &entity.Attendee{}, errors.New("some error, this is a mock")
}

func (s *MockAttendeeService) UpdateAttendee(attendee *entity.Attendee) error {
	return errors.New("some error, this is a mock")
}

func tstSetupServiceMocks() {
	attendeeService = &MockAttendeeService{}
}
