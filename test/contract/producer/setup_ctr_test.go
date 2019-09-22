package producer

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/web"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/controller/attendeectl"
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
	config.LoadTestingConfigurationFromPathOrAbort("../../../test/testconfig.yaml")
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

func tstSetupServiceMocks() {
	attendeectl.OverrideAttendeeService(&MockAttendeeService{})
}
