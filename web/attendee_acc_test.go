package web

import (
	"rexis/rexis-go-attendee/docs"
	"testing"
)

// see config in setup_all_test.go

func TestCreateNewAttendee(t *testing.T) {
	docs.Given("given an unauthenticated user")
	docs.When( "when they create a new attendee")
	docs.Then( "then the attendee is successfully created")
}
