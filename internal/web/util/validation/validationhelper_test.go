package validation

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
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

func TestDateNotInRangeInclusive(t *testing.T) {
	docs.Description("verify that the dateNotInRange check works as expected")
	min := "1999-02-28"
	max := "2000-02-29"
	require.True(t, DateNotInRangeInclusive("1998-04-22", min, max))
	require.False(t, DateNotInRangeInclusive(min, min, max))
	require.False(t, DateNotInRangeInclusive("2000-01-01", min, max))
	require.False(t, DateNotInRangeInclusive(max, min, max))
	require.True(t, DateNotInRangeInclusive("2004-12-31", min, max))
}
