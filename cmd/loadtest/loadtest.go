package main

import (
	"github.com/eurofurence/reg-attendee-service/internal/web/app"
	"os"
)

func main() {
	os.Exit(app.New().Loadtest())
}
