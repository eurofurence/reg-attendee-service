package attendeectl

import (
	"github.com/stretchr/testify/require"
	"rexis/rexis-go-attendee/api/v1/attendee"
	"rexis/rexis-go-attendee/docs"
	"rexis/rexis-go-attendee/internal/entity"
	"testing"
)

func TestMapping(t *testing.T) {
	docs.Description("mapping an attendee dto back and forth should result in the same data")
	attendeeDtoSource := tstCreateValidAttendee()
	attendeeEntity := entity.Attendee{}
	_ = mapDtoToAttendee(&attendeeDtoSource, &attendeeEntity)

	attendeeDtoResult := attendee.AttendeeDto{}
	mapAttendeeToDto(&attendeeEntity, &attendeeDtoResult)
	// id differences are ok because the field is only mapped one way, so overwrite with actual value
	attendeeDtoSource.Id = attendeeDtoResult.Id
	require.EqualValues(t, attendeeDtoSource, attendeeDtoResult, "unexpected difference after mapping back and forth")
}
