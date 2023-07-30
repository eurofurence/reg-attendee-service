package inmemorydb

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMatchesIsoDateRange(t *testing.T) {
	require.True(t, matchesIsoDateRange("", "", ""))
	require.True(t, matchesIsoDateRange("", "", "1972-10-24"))
	require.True(t, matchesIsoDateRange("1976-10-22", "1977-01-01", ""))
	require.True(t, matchesIsoDateRange("1976-10-22", "1977-01-01", "1976-12-24"))
	require.True(t, matchesIsoDateRange("1976-10-22", "1977-01-01", "1976-10-22")) // left inclusive
	require.True(t, matchesIsoDateRange("1976-10-22", "1977-01-01", "1977-01-01")) // right inclusive
	require.False(t, matchesIsoDateRange("1976-10-22", "1977-01-01", "1972-12-24"))
	require.False(t, matchesIsoDateRange("1976-10-22", "1977-01-01", "1979-12-24"))
}
