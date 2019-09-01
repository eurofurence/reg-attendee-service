package main

import (
	"flag"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/web"
)

func main() {
	flag.Parse()
	config.StartupLoadConfiguration()
	web.StartWebserverAndNeverReturn()
}
