package main

import (
	"flag"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database"
	"github.com/jumpy-squirrel/rexis-go-attendee/web"
)

func main() {
	flag.Parse()
	config.StartupLoadConfiguration()
	database.Open()
	defer database.Close()
	web.StartWebserverAndNeverReturn()
}
