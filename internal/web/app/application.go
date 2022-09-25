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

	if err := config.StartupLoadConfiguration(); err != nil {
		return 1
	}

	database.Open()
	defer database.Close()
	database.MigrateIfSwitchedOn()

	paymentservice.Create()
	mailservice.Create()

	if err := runServerWithGracefulShutdown(); err != nil {
		return 2
	}

	return 0
}
