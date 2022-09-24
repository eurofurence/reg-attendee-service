package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web"
	"net/http/httptest"
)

// placing these here because they are package global

var (
	ts          *httptest.Server
	paymentMock paymentservice.Mock
	mailMock    mailservice.Mock
)

const tstDefaultConfigFile = "../../test/testconfig.yaml"
const tstDefaultConfigFileBeforeTarget = "../../test/testconfig-before-target.yaml"
const tstStaffregConfigFile = "../../test/testconfig-staffreg.yaml"

func tstSetup(configFilePath string) {
	tstSetupConfig(configFilePath)
	paymentMock = paymentservice.CreateMock()
	mailMock = mailservice.CreateMock()
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
	paymentMock.Reset()
	mailMock.Reset()
}
