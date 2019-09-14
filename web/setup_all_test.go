package web

import (
	"net/http/httptest"
	"os"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/internal/repository/database"
	"testing"
)

// placing these here because they are package global

var (
	ts *httptest.Server
)

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
	tstSetupConfig()
	tstSetupDatabase()
	tstSetupHttpTestServer()
}

func tstSetupConfig() {
	yaml := "" +
		`choices:
  flags:
    hc:
      description: 'blah'
    anon:
      description: 'blah'
    ev:  
      description: 'blah'
  packages:
    room-none:
      description: 'blah'
    attendance:
      description: 'blah'
    stage:
      description: 'blah'
    sponsor:
      description: 'blah'
    sponsor2:
      description: 'blah'
  options:
    art:
      description: 'blah'
    anim:
      description: 'blah'
    music:
      description: 'blah'
    suit:
      description: 'blah'
`
	err := config.InitializeConfiguration(yaml)
	if err != nil {
		os.Exit(1)
	}
}

func tstSetupHttpTestServer() {
	router := CreateRouter()
	ts = httptest.NewServer(router)
}

func tstSetupDatabase() {
	database.Open()
}

func tstShutdown() {
	ts.Close()
	database.Close()
}
