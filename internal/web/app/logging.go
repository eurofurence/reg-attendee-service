package app

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	auzerolog "github.com/StephanHCB/go-autumn-logging-zerolog"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/rs/zerolog"
)

func setupLogging(applicationName string, useEcsLogging bool) {
	aulogging.RequestIdRetriever = ctxvalues.RequestId
	if useEcsLogging {
		auzerolog.SetupJsonLogging(applicationName)
		zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z"
	} else {
		aulogging.DefaultRequestIdValue = "00000000"
		auzerolog.SetupPlaintextLogging()
	}
}

func setLoglevel(severity string) {
	switch severity {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
	}
}
