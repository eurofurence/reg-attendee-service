package acceptance

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/service/attendeesrv"
	"github.com/eurofurence/reg-attendee-service/internal/web/app"
	"net/http/httptest"
	"time"
)

// placing these here because they are package global

var (
	ts          *httptest.Server
	paymentMock paymentservice.Mock
	mailMock    mailservice.Mock
)

const tstDefaultConfigFileBeforeTarget = "../../test/testconfig-before-target.yaml"

func tstConfigFileIrrelevant() string {
	return tstConfigFile(true, false, true)
}

func tstConfigFile(needLogin bool, staffReg bool, afterTarget bool) string {
	path := "../../test/testconfig-"
	if needLogin {
		path += "needlogin"
	} else {
		path += "public"
	}
	if staffReg {
		if afterTarget {
			// after the staffreg target but before the normal target
			path += "-staffreg.yaml"
		} else {
			// neither public reg nor staff reg has started yet
			path += "-before-target-staffreg.yaml"
		}
		// we do not test the after both targets case separately because it is a low risk case
	} else {
		if afterTarget {
			// after main target, no staff reg configured
			path += ".yaml"
		} else {
			// no staff reg configured
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
	attSrv := attendeesrv.New()
	attSrv.(*attendeesrv.AttendeeServiceImplData).Now = func() time.Time {
		t, _ := time.Parse(config.IsoDateFormat, "2022-12-08")
		return t
	}
	router := app.CreateRouter(context.Background(), attSrv)
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
