package attendeectl

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/stretchr/testify/mock"
	"os"
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
	config.LoadTestingConfigurationFromPathOrAbort("../../../../test/testconfig.yaml")
}

type MockAttendeeService struct {
	mock.Mock
}

var _ attendeesrv.AttendeeService = (*MockAttendeeService)(nil)

func (s *MockAttendeeService) NewAttendee(ctx context.Context) *entity.Attendee {
	return &entity.Attendee{}
}

func (s *MockAttendeeService) RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error) {
	return 0, errors.New("some error, this is a mock")
}

func (s *MockAttendeeService) GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error) {
	return &entity.Attendee{}, errors.New("some error, this is a mock")
}

func (s *MockAttendeeService) UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error {
	return errors.New("some error, this is a mock")
}

func (s *MockAttendeeService) GetAttendeeMaxId(ctx context.Context) (uint, error) {
	return 0, nil
}

func (s *MockAttendeeService) CanChangeChoiceTo(ctx context.Context, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error {
	return nil
}

func (s *MockAttendeeService) CanRegisterAtThisTime(ctx context.Context) error {
	return nil
}

func (s *MockAttendeeService) GetAdminInfo(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	return &entity.AdminInfo{}, nil
}

func (s *MockAttendeeService) UpdateAdminInfo(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo) error {
	return nil
}

func (s *MockAttendeeService) GetFullStatusHistory(ctx context.Context, attendee *entity.Attendee) ([]entity.StatusChange, error) {
	return []entity.StatusChange{}, nil
}

func (s *MockAttendeeService) UpdateDuesAndDoStatusChangeIfNeeded(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string, comments string) error {
	return nil
}

func (s *MockAttendeeService) StatusChangeAllowed(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error {
	return nil
}

func (s *MockAttendeeService) StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error {
	return nil
}

func tstSetupServiceMocks() {
	attendeeService = &MockAttendeeService{}
}
