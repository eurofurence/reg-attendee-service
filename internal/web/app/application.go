package app

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
)

type Application interface {
	Run() int
}

type Impl struct{}

func New() Application {
	return &Impl{}
}

func (i *Impl) Run() int {
	config.ParseCommandLineFlags()
	setupLogging("attendee-service", config.UseEcsLogging())

	if err := config.StartupLoadConfiguration(); err != nil {
		return 1
	}
	setLoglevel(config.LoggingSeverity())

	if err := database.Open(); err != nil {
		return 1
	}
	defer database.Close()
	if err := database.MigrateIfSwitchedOn(); err != nil {
		return 1
	}

	if err := paymentservice.Create(); err != nil {
		return 1
	}
	if err := mailservice.Create(); err != nil {
		return 1
	}

	if err := runServerWithGracefulShutdown(); err != nil {
		return 2
	}

	return 0
}
