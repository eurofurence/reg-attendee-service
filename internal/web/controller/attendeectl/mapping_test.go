package attendeectl

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
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

func TestMapping_PackagesClassic(t *testing.T) {
	docs.Description("mapping an attendee dto back and forth should result in the same data - packages provided as comma separated list")
	attendeeDtoSource := tstCreateValidAttendee()
	keepPkgList := attendeeDtoSource.PackagesList
	attendeeDtoSource.PackagesList = nil
	attendeeEntity := entity.Attendee{}
	mapDtoToAttendee(&attendeeDtoSource, &attendeeEntity)

	attendeeDtoResult := attendee.AttendeeDto{}
	mapAttendeeToDto(&attendeeEntity, &attendeeDtoResult)
	// id differences are ok because the field is only mapped one way, so overwrite with actual value
	attendeeDtoSource.Id = attendeeDtoResult.Id
	// reverse mapping fills the package list, so restore it now
	attendeeDtoSource.PackagesList = keepPkgList
	require.EqualValues(t, attendeeDtoSource, attendeeDtoResult, "unexpected difference after mapping back and forth")
}

func TestMapping_PackagesAsList(t *testing.T) {
	docs.Description("mapping an attendee dto back and forth should result in the same data - packages provided in packages_list field")
	attendeeDtoSource := tstCreateValidAttendee()
	keepPkgClassic := attendeeDtoSource.Packages
	attendeeDtoSource.Packages = ""
	attendeeEntity := entity.Attendee{}
	mapDtoToAttendee(&attendeeDtoSource, &attendeeEntity)

	attendeeDtoResult := attendee.AttendeeDto{}
	mapAttendeeToDto(&attendeeEntity, &attendeeDtoResult)
	// id differences are ok because the field is only mapped one way, so overwrite with actual value
	attendeeDtoSource.Id = attendeeDtoResult.Id
	// reverse mapping fills both packages_list and packages, so restore it now
	attendeeDtoSource.Packages = keepPkgClassic
	require.EqualValues(t, attendeeDtoSource, attendeeDtoResult, "unexpected difference after mapping back and forth")
}

func TestMapping_DuplicatePackage(t *testing.T) {

}
