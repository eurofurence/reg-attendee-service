package main

import (
	"flag"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web"
)

func main() {
	flag.Parse()
	config.StartupLoadConfiguration()
	database.Open()
	defer database.Close()
	database.MigrateIfSwitchedOn()
	paymentservice.Create()
	mailservice.Create()
	server := web.Create()
	web.StartWebserverAndNeverReturn(server)
}
