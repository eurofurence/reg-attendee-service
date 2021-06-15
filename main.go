package main

import (
	"flag"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/web"
)

func main() {
	flag.Parse()
	config.StartupLoadConfiguration()
	database.Open()
	defer database.Close()
	database.MigrateIfSwitchedOn()
	server := web.Create()
	web.StartWebserverAndNeverReturn(server)
}
