package acceptance

import (
	"net/http/httptest"
	"os"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database"
	"github.com/jumpy-squirrel/rexis-go-attendee/web"
	"testing"
)

// placing these here because they are package global

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
	tstSetupDatabase()
	tstSetupHttpTestServer()
}

func tstSetupConfig() {
	config.LoadTestingConfigurationFromPathOrAbort("../../test/testconfig.yaml")
}

func tstSetupHttpTestServer() {
	router := web.CreateRouter()
	ts = httptest.NewServer(router)
}

func tstSetupDatabase() {
	database.Open()
}

func tstShutdown() {
	ts.Close()
	database.Close()
}
