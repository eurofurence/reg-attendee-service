package app

import (
	"flag"
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
	flag.Parse()
	config.StartupLoadConfiguration()

	database.Open()
	defer database.Close()
	database.MigrateIfSwitchedOn()

	paymentservice.Create()
	mailservice.Create()

	server := CreateRouter()
	err := StartWebserver(server)
	if err != nil {
		return 1
	}
	return 0
}
