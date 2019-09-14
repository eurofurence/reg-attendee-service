package attendeectl

import (
	"os"
	"rexis/rexis-go-attendee/internal/repository/config"
	"testing"
)

// placing these here because they are package global

func TestMain(m *testing.M) {
	tstSetup()
	code := m.Run()
	tstShutdown()
	os.Exit(code)
}

func tstSetup() {
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

func tstShutdown() {

}
