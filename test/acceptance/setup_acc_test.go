package acceptance

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database"
	"github.com/jumpy-squirrel/rexis-go-attendee/web"
	"net/http/httptest"
)

// placing these here because they are package global

var (
	ts *httptest.Server
)

const tstDefaultConfigFile =  "../../test/testconfig.yaml"
const tstStaffregConfigFile = "../../test/testconfig-staffreg.yaml"

func tstSetup(configFilePath string) {
	tstSetupConfig(configFilePath)
	tstSetupDatabase()
	tstSetupHttpTestServer()
}

func tstSetupConfig(configFilePath string) {
	config.LoadTestingConfigurationFromPathOrAbort(configFilePath)
}

func tstSetupHttpTestServer() {
	router := web.CreateRouter()
	ts = httptest.NewServer(router)
}

func tstSetupDatabase() {
	database.Open()
	config.EnableTestingMigrateDatabase()
	database.MigrateIfSwitchedOn()
}

func tstShutdown() {
	ts.Close()
	database.Close()
}
