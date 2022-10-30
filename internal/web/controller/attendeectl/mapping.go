package attendeectl

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"strings"
)

func mapDtoToAttendee(dto *attendee.AttendeeDto, a *entity.Attendee) {
	// do not map id - instead load by ID from db or you'll introduce errors
	a.Nickname = dto.Nickname
	a.FirstName = dto.FirstName
	a.LastName = dto.LastName
	a.Street = dto.Street
	a.Zip = dto.Zip
	a.City = dto.City
	a.Country = dto.Country
	a.CountryBadge = dto.CountryBadge
	a.State = dto.State
	a.Email = dto.Email
	a.Phone = dto.Phone
	a.Telegram = dto.Telegram
	a.Partner = dto.Partner
	a.Birthday = dto.Birthday
	a.Gender = dto.Gender
	a.Pronouns = dto.Pronouns
	a.TshirtSize = dto.TshirtSize
	a.Flags = addWrappingCommas(dto.Flags)
	a.Packages = addWrappingCommas(dto.Packages)
	a.Options = addWrappingCommas(dto.Options)
	a.UserComments = dto.UserComments
}

func mapAttendeeToDto(a *entity.Attendee, dto *attendee.AttendeeDto) {
	// this cannot fail
	dto.Id = fmt.Sprint(a.ID)
	dto.Nickname = a.Nickname
	dto.FirstName = a.FirstName
	dto.LastName = a.LastName
	dto.Street = a.Street
	dto.Zip = a.Zip
	dto.City = a.City
	dto.Country = a.Country
	dto.CountryBadge = a.CountryBadge
	dto.State = a.State
	dto.Email = a.Email
	dto.Phone = a.Phone
	dto.Telegram = a.Telegram
	dto.Partner = a.Partner
	dto.Birthday = a.Birthday
	dto.Gender = a.Gender
	dto.Pronouns = a.Pronouns
	dto.TshirtSize = a.TshirtSize
	dto.Flags = removeWrappingCommas(a.Flags)
	dto.Packages = removeWrappingCommas(a.Packages)
	dto.Options = removeWrappingCommas(a.Options)
	dto.UserComments = a.UserComments
}

func removeWrappingCommas(v string) string {
	v = strings.TrimPrefix(v, ",")
	v = strings.TrimSuffix(v, ",")
	return v
}

func addWrappingCommas(v string) string {
	if !strings.HasPrefix(v, ",") {
		v = "," + v
	}
	if !strings.HasSuffix(v, ",") {
		v = v + ","
	}
	return v
}
