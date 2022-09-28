package acceptance

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/app"
	"net/http/httptest"
)

// placing these here because they are package global

var (
	ts          *httptest.Server
	paymentMock paymentservice.Mock
	mailMock    mailservice.Mock
)

const tstDefaultConfigFileBeforeTarget = "../../test/testconfig-before-target.yaml"

func tstConfigFile(needLogin bool, staffReg bool, afterTarget bool) string {
	path := "../../test/testconfig-"
	if needLogin {
		path += "needlogin"
	} else {
		path += "public"
	}
	if afterTarget {
		path += ".yaml"
	} else {
		if staffReg {
			path += "-staffreg.yaml"
		} else {
			path += "-before-target.yaml"
		}
	}
	return path
}

func tstSetup(configFilePath string) {
	tstSetupConfig(configFilePath)
	paymentMock = paymentservice.CreateMock()
	mailMock = mailservice.CreateMock()
	tstSetupDatabase()
	tstSetupHttpTestServer()
}

func tstSetupConfig(configFilePath string) {
	aulogging.SetupNoLoggerForTesting()
	config.LoadTestingConfigurationFromPathOrAbort(configFilePath)
}

func tstSetupHttpTestServer() {
	router := app.CreateRouter(context.Background())
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
