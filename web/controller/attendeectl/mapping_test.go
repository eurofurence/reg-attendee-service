package attendeectl

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMapping(t *testing.T) {
	docs.Description("mapping an attendee dto back and forth should result in the same data")
	attendeeDtoSource := tstCreateValidAttendee()
	attendeeEntity := entity.Attendee{}
	mapDtoToAttendee(&attendeeDtoSource, &attendeeEntity)

	attendeeDtoResult := attendee.AttendeeDto{}
	mapAttendeeToDto(&attendeeEntity, &attendeeDtoResult)
	// id differences are ok because the field is only mapped one way, so overwrite with actual value
	attendeeDtoSource.Id = attendeeDtoResult.Id
	require.EqualValues(t, attendeeDtoSource, attendeeDtoResult, "unexpected difference after mapping back and forth")
}
