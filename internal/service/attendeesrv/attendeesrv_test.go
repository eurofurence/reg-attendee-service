package attendeesrv

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/stretchr/testify/require"
	"testing"
)

var tstChoiceConfig = map[string]config.ChoiceConfig{
	"a": {
		MaxCount: 0, // interpreted as 1
	},
	"b": {
		MaxCount: 1,
	},
	"c": {
		MaxCount: 4,
	},
}

func TestChoiceStrToMap_Flags(t *testing.T) {
	actual := choiceStrToMap(",a,b,", tstChoiceConfig)
	expected := map[string]int{
		"a": 1,
		"b": 1,
		"c": 0,
	}
	require.EqualValues(t, expected, actual)
}

func TestChoiceStrToMap_Packages(t *testing.T) {
	actual := choiceStrToMap(",a:1,c,b,c:2,", tstChoiceConfig)
	expected := map[string]int{
		"a": 1,
		"b": 1,
		"c": 3,
	}
	require.EqualValues(t, expected, actual)
}
