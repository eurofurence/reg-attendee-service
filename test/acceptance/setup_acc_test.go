package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/web"
	"net/http/httptest"
)

// placing these here because they are package global

var (
	ts *httptest.Server
)

const tstDefaultConfigFile = "../../test/testconfig.yaml"
const tstDefaultConfigFileBeforeTarget = "../../test/testconfig-before-target.yaml"
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
	router := web.Create()
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
