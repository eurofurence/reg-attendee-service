package system

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"os"
)

var TestingMode = false
var TestingExitCounter = 0

func Exit(code int) {
	if TestingMode {
		TestingExitCounter++
		aulogging.Logger.NoCtx().Info().Printf("would now os.exit with code %v", code)
	} else {
		os.Exit(code)
	}
}
