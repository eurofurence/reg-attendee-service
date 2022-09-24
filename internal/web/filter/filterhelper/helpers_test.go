package filterhelper

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/repository/system"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTimeout(t *testing.T) {
	docs.Description("parseTimeout should correctly convert timeouts given as string to a Duration")
	require.Equal(t, int64(800000000), parseTimeout("800ms").Nanoseconds(), "unexpected conversion result")
}

func TestParseTimeoutInvalid(t *testing.T) {
	docs.Description("parseTimeout should log an error and os.Exit when given an invalid timeout")
	system.TestingMode = true
	oldcounter := system.TestingExitCounter
	parseTimeout("3.8e24lightyears")
	require.Equal(t, oldcounter+1, system.TestingExitCounter, "should have exited")
}
