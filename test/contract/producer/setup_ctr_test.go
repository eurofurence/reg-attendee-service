package producer

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/app"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/adminctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/attendeectl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/banctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/statusctl"
	"github.com/stretchr/testify/mock"
	"net/http/httptest"
	"os"
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
	aulogging.SetupNoLoggerForTesting()
	config.LoadTestingConfigurationFromPathOrAbort("../../../test/testconfig-public.yaml")
}

func tstSetupHttpTestServer() {
	router := app.CreateRouter(context.Background())
	ts = httptest.NewServer(router)
}

func tstShutdown() {
	ts.Close()
}

type MockAttendeeService struct {
	mock.Mock
}

var _ attendeesrv.AttendeeService = (*MockAttendeeService)(nil)

func (s *MockAttendeeService) NewAttendee(ctx context.Context) *entity.Attendee {
	return &entity.Attendee{}
}

func (s *MockAttendeeService) RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error) {
	// TODO use mock to verify data for contract tests
	return 1, nil
}

func (s *MockAttendeeService) GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error) {
	// TODO when writing a contract test, put matching response data here
	return &entity.Attendee{}, nil
}

func (s *MockAttendeeService) UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error {
	return nil
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

func (s *MockAttendeeService) GetBan(ctx context.Context, id uint) (*entity.Ban, error) {
	return &entity.Ban{}, nil
}

func (s *MockAttendeeService) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	return make([]*entity.Ban, 0), nil
}

func tstSetupServiceMocks() {
	attendeeServiceMock := MockAttendeeService{}
	attendeectl.OverrideAttendeeService(&attendeeServiceMock)
	adminctl.OverrideAttendeeService(&attendeeServiceMock)
	statusctl.OverrideAttendeeService(&attendeeServiceMock)
	banctl.OverrideAttendeeService(&attendeeServiceMock)
}
