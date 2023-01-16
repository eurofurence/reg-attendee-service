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

const tstDefaultConfigFile = "../../test/testconfig-base.yaml"

func tstSetupConfigIrrelevant() {
	tstSetup(true, false, true)
}

func tstAdjustConfig(needLogin bool, earlyReg bool, afterTarget bool) {
	conf := config.Configuration()
	if needLogin {
		conf.Service.Name += " With Login"
		conf.Security.RequireLogin = true

		// just so we cover these code paths somewhere:
		conf.Security.Cors.DisableCors = true
		conf.Security.Cors.AllowOrigin = "http://localhost:8000"
	} else {
		conf.Service.Name += " No Login"
	}

	if earlyReg {
		if afterTarget {
			conf.Service.Name += " After Early Reg Started But Before Normal Reg"
			conf.Security.Oidc.EarlyReg = "staff"
			// public reg has not started yet, so only staff may register
			conf.GoLive.StartIsoDatetime = "2050-11-28T20:00:00+01:00"
			conf.GoLive.EarlyRegStartIsoDatetime = "2019-10-31T20:00:00+01:00"
		} else {
			conf.Service.Name += " Before Early Reg"
			conf.Security.Oidc.EarlyReg = "staff"
			// neither public reg nor staff reg has started yet
			conf.GoLive.StartIsoDatetime = "2050-11-28T20:00:00+01:00"
			conf.GoLive.EarlyRegStartIsoDatetime = "2050-10-31T20:00:00+01:00"
		}
		// we do not test the after both targets case separately because it is a low risk case
	} else {
		if afterTarget {
			conf.Service.Name += " After Target No Early Reg Configured"

			// just so we cover these code paths somewhere:
			// allow deselecting, so we can test at-least-one-mandatory
			att := conf.Choices.Packages["attendance"]
			att.ReadOnly = false
			conf.Choices.Packages["attendance"] = att
			sta := conf.Choices.Packages["stage"]
			sta.ReadOnly = false
			conf.Choices.Packages["stage"] = sta
		} else {
			// no staff reg configured, and public reg has not started yet
			conf.Service.Name += " Before Target No Early Reg Configured"
			conf.GoLive.StartIsoDatetime = "2050-11-28T20:00:00+01:00"
		}
	}
}

func tstSetup(needLogin bool, staffReg bool, afterTarget bool) {
	tstSetupConfig(tstDefaultConfigFile)
	tstAdjustConfig(needLogin, staffReg, afterTarget)
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
