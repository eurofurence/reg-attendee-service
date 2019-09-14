package validation

import (
	"rexis/rexis-go-attendee/docs"
	"testing"
)

// here we only add tests for code that isn't already covered by the tests of validation.go

func TestViolatesPatternReportsViolationWithErrorInPattern(t *testing.T) {
	docs.Description("ViolatesPattern should return true if there is a syntax error in the pattern")
	wrongPattern := "^blabla($"
	value := "hello"
	if !ViolatesPattern(wrongPattern, value) {
		t.Errorf("ViolatesPattern did not return true when the pattern contains a syntax error")
	}
}
