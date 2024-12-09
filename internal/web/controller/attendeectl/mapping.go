package attendeectl

import (
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"sort"
	"strconv"
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
	a.Packages = packagesFromDto(dto.Packages, dto.PackagesList)
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
	dto.Packages = packagesFromEntity(a.Packages)
	dto.PackagesList = packagesListFromEntity(a.Packages)
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

// packages conversion from attendee entity (NOT dto)
//
// Note: here, the :count postfixes are supported.

func packagesFromEntity(entityPackages string) string {
	asList := packagesListFromEntity(entityPackages)

	var result []string
	for _, entry := range asList {
		for i := 0; i < entry.Count; i++ {
			result = append(result, entry.Name)
		}
	}
	return strings.Join(result, ",")
}

func packagesListFromEntity(entityPackages string) []attendee.PackageState {
	unwrapped := removeWrappingCommas(entityPackages)

	result := make([]attendee.PackageState, 0)

	if unwrapped == "" {
		return result
	}

	counts := make(map[string]int)
	entries := strings.Split(unwrapped, ",")
	for _, entry := range entries {
		nameAndPossiblyCount := strings.Split(entry, ":")
		name := nameAndPossiblyCount[0]
		count := 1
		if len(nameAndPossiblyCount) > 1 {
			var err error
			count, err = strconv.Atoi(nameAndPossiblyCount[1])
			if err != nil {
				aulogging.Logger.NoCtx().Warn().Printf("encountered invalid choice entry '%s' in database - ignoring (please fix!)", entry)
				continue
			}
		}

		currentCount, _ := counts[name]
		counts[name] = currentCount + count
	}

	for name, count := range counts {
		result = append(result, attendee.PackageState{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// packages conversion from dto
//
// Note: in dtos, the :count postfixes are not supported (internal database format only)

func packagesFromDto(commaSeparated string, asList []attendee.PackageState) string {
	alwaysAsList := packagesListWithPrecedence(commaSeparated, asList)

	var result strings.Builder
	result.WriteString(",")
	for _, item := range alwaysAsList {
		if item.Count == 0 {
			// item.Count = 0 should be interpreted as 1 (allows omitting Count in requests)
			item.Count = 1
		}

		result.WriteString(fmt.Sprintf("%s:%d,", item.Name, item.Count))
	}
	return result.String()
}

func packagesListWithPrecedence(commaSeparated string, asList []attendee.PackageState) []attendee.PackageState {
	// only use commaSeparated if no list is supplied
	if len(asList) == 0 {
		return packageListFromCommaSeparated(commaSeparated)
	} else {
		return asList
	}
}

// packageListFromCommaSeparated takes a comma separated list, without leading and trailing commas, and converts
// it into a sorted package_list
//
// This version does not support :count postfixes, that only occurs in entities.
func packageListFromCommaSeparated(commaSeparatedValue string) []attendee.PackageState {
	result := make([]attendee.PackageState, 0)

	if commaSeparatedValue == "" {
		return result
	}

	counts := make(map[string]int)
	chosenValues := strings.Split(commaSeparatedValue, ",")
	for _, v := range chosenValues {
		currentCount, ok := counts[v]
		if !ok {
			counts[v] = 1
		} else {
			counts[v] = currentCount + 1
		}
	}

	for name, count := range counts {
		result = append(result, attendee.PackageState{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}
