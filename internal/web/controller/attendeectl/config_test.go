package attendeectl

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
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
	config.LoadTestingConfigurationFromPathOrAbort("../../../../test/testconfig-base.yaml")
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

func (s *MockAttendeeService) CanChangeEmailTo(ctx context.Context, originalEmail string, newEmail string) error {
	return nil
}

func (s *MockAttendeeService) CanChangeChoiceTo(ctx context.Context, what string, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error {
	return nil
}

func (s *MockAttendeeService) CanChangeChoiceToCurrentStatus(ctx context.Context, what string, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig, currentStatus status.Status) error {
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

func (s *MockAttendeeService) UpdateDuesAndDoStatusChangeIfNeeded(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status, statusComment string, overrideDuesComment string) error {
	return nil
}

func (s *MockAttendeeService) StatusChangeAllowed(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status) error {
	return nil
}

func (s *MockAttendeeService) StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status) error {
	return nil
}

func (s *MockAttendeeService) IsOwnerFor(ctx context.Context) ([]*entity.Attendee, error) {
	return make([]*entity.Attendee, 0), nil
}

func (s *MockAttendeeService) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) (*attendee.AttendeeSearchResultList, error) {
	return &attendee.AttendeeSearchResultList{
		Attendees: make([]attendee.AttendeeSearchResult, 0),
	}, nil
}

func (s *MockAttendeeService) NewBan(ctx context.Context) *entity.Ban {
	return &entity.Ban{}
}

func (s *MockAttendeeService) CreateBan(ctx context.Context, ban *entity.Ban) (uint, error) {
	return 1, nil
}

func (s *MockAttendeeService) UpdateBan(ctx context.Context, ban *entity.Ban) error {
	return nil
}

func (s *MockAttendeeService) DeleteBan(ctx context.Context, ban *entity.Ban) error {
	return nil
}

func (s *MockAttendeeService) GetBan(ctx context.Context, id uint) (*entity.Ban, error) {
	return &entity.Ban{}, nil
}

func (s *MockAttendeeService) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	return make([]*entity.Ban, 0), nil
}

func (s *MockAttendeeService) GetAdditionalInfo(ctx context.Context, attendeeId uint, area string) (string, error) {
	return "", nil
}

func (s *MockAttendeeService) WriteAdditionalInfo(ctx context.Context, attendeeId uint, area string, value string) error {
	return nil
}

func (s *MockAttendeeService) CanAccessAdditionalInfoArea(ctx context.Context, area string) (bool, error) {
	return false, nil
}

func tstSetupServiceMocks() {
	attendeeService = &MockAttendeeService{}
}
