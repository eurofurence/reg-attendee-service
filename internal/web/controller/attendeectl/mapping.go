package attendeectl

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
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
	a.State = dto.State
	a.Email = dto.Email
	a.Phone = dto.Phone
	a.Telegram = dto.Telegram
	a.Partner = dto.Partner
	a.Birthday = dto.Birthday
	if dto.Gender != "" {
		a.Gender = dto.Gender
	} else {
		a.Gender = "notprovided"
	}
	a.Pronouns = dto.Pronouns
	a.TshirtSize = dto.TshirtSize
	a.SpokenLanguages = addWrappingCommas(dto.SpokenLanguages)
	if dto.RegistrationLanguage != "" {
		a.RegistrationLanguage = addWrappingCommas(dto.RegistrationLanguage)
	} else {
		a.RegistrationLanguage = addWrappingCommas(config.DefaultRegistrationLanguage())
	}
	a.Flags = addWrappingCommas(dto.Flags)
	a.Packages = addWrappingCommas(dto.Packages)
	a.Options = addWrappingCommas(dto.Options)
	a.UserComments = dto.UserComments
}

func mapAttendeeToDto(a *entity.Attendee, dto *attendee.AttendeeDto) {
	// this cannot fail
	dto.Id = a.ID
	dto.Nickname = a.Nickname
	dto.FirstName = a.FirstName
	dto.LastName = a.LastName
	dto.Street = a.Street
	dto.Zip = a.Zip
	dto.City = a.City
	dto.Country = a.Country
	dto.State = a.State
	dto.Email = a.Email
	dto.Phone = a.Phone
	dto.Telegram = a.Telegram
	dto.Partner = a.Partner
	dto.Birthday = a.Birthday
	dto.Gender = a.Gender
	dto.Pronouns = a.Pronouns
	dto.TshirtSize = a.TshirtSize
	dto.SpokenLanguages = removeWrappingCommas(a.SpokenLanguages)
	dto.RegistrationLanguage = removeWrappingCommas(a.RegistrationLanguage)
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

func commaSeparatedContains(commaSeparated string, singleValue string) bool {
	list := strings.Split(removeWrappingCommas(commaSeparated), ",")
	return sliceContains(list, singleValue)
}

func sliceContains(slice []string, singleValue string) bool {
	for _, e := range slice {
		if e == singleValue {
			return true
		}
	}
	return false
}
