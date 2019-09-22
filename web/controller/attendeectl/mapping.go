package attendeectl

import (
	"fmt"
	"github.com/jumpy-squirrel/rexis-go-attendee/api/v1/attendee"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
)

func mapDtoToAttendee(dto *attendee.AttendeeDto, a *entity.Attendee) error {
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
	a.Birthday = dto.Birthday
	a.Gender = dto.Gender
	a.TshirtSize = dto.TshirtSize
	a.Flags = dto.Flags
	a.Packages = dto.Packages
	a.Options = dto.Options
	a.UserComments = dto.UserComments
	return nil
}

func mapAttendeeToDto(a *entity.Attendee, dto *attendee.AttendeeDto)  {
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
	dto.Birthday = a.Birthday
	dto.Gender = a.Gender
	dto.TshirtSize = a.TshirtSize
	dto.Flags = a.Flags
	dto.Packages = a.Packages
	dto.Options = a.Options
	dto.UserComments = a.UserComments
}
