package attendeectl

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/jumpy-squirrel/rexis-go-attendee/api/v1/attendee"
	"github.com/jumpy-squirrel/rexis-go-attendee/docs"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"net/url"
	"reflect"
	"testing"
)

func tstCreateValidAttendee() attendee.AttendeeDto {
	return attendee.AttendeeDto{
		Nickname:     "BlackCheetah",
		FirstName:    "Hans",
		LastName:     "Mustermann",
		Street:       "Teststraße 24",
		Zip:          "12345",
		City:         "Berlin",
		Country:      "DE",
		CountryBadge: "DE",
		State:        "Sachsen",
		Email:        "jsquirrel_github_9a6d@packetloss.de",
		Phone:        "+49-30-123",
		Telegram:     "@ihopethisuserdoesnotexist",
		Birthday:     "1998-11-23",
		Gender:       "other",
		Flags:        "anon,ev",
		Packages:     "room-none,attendance,stage,sponsor2",
		Options:      "music,suit",
		TshirtSize:   "XXL",
	}
}

func TestValidateSuccess(t *testing.T) {
	docs.Description("a valid attendee reports no validation errors")
	a := tstCreateValidAttendee()
	expected := url.Values{}
	performValidationTest(t, &a, expected, 0)
}

func TestValidateMissingInfo(t *testing.T) {
	docs.Description("an attendee with wrong and missing fields reports the expected validation errors")
	a := attendee.AttendeeDto{
		Country:      "meow",
		CountryBadge: "bark",
	}
	expected := url.Values{
		"birthday":      []string{"birthday field must be a valid ISO 8601 date (format yyyy-MM-dd)"},
		"city":          []string{"city field must be at least 1 and at most 80 characters long"},
		"country":       []string{"country field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)"},
		"country_badge": []string{"country_badge field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)"},
		"email":         []string{"email field must be at least 1 and at most 200 characters long", "email field is not plausible"},
		"first_name":    []string{"first_name field must be at least 1 and at most 80 characters long"},
		"last_name":     []string{"last_name field must be at least 1 and at most 80 characters long"},
		"nickname": []string{"nickname field must contain at least two letters, and contain no more than two non-letters",
			"nickname field must be at least 2 and at most 80 characters long"},
		"phone":  []string{"phone field must be at least 1 and at most 32 characters long"},
		"street": []string{"street field must be at least 1 and at most 120 characters long"},
		"zip":    []string{"zip field must be at least 1 and at most 20 characters long"},
	}
	performValidationTest(t, &a, expected, 0)
}

func TestValidateTooLong(t *testing.T) {
	docs.Description("an attendee with just barely too long field values reports the expected validation errors")
	a := tstCreateValidAttendee()
	a.Nickname = "ThisIsASuperLongNicknameWhichIsNotAllowedBecauseItWillNotFitOnTheBadgeAndAnywayWh"
	tooLong := "And this is a super long text that we will use to test for the length limits of the other fields. While we do this, " +
		"we will cut off at just the right place to make it 1 character too long. I hope this text is long enough in total!"
	a.City = tooLong[0:81]
	a.Email = tooLong[0:201]
	a.FirstName = tooLong[0:81]
	a.LastName = tooLong[0:81]
	a.Phone = tooLong[0:33]
	a.Street = tooLong[0:121]
	a.Zip = tooLong[0:21]

	expected := url.Values{
		"city":       []string{"city field must be at least 1 and at most 80 characters long"},
		"email":      []string{"email field must be at least 1 and at most 200 characters long", "email field is not plausible"},
		"first_name": []string{"first_name field must be at least 1 and at most 80 characters long"},
		"last_name":  []string{"last_name field must be at least 1 and at most 80 characters long"},
		"nickname":   []string{"nickname field must be at least 2 and at most 80 characters long"},
		"phone":      []string{"phone field must be at least 1 and at most 32 characters long"},
		"street":     []string{"street field must be at least 1 and at most 120 characters long"},
		"zip":        []string{"zip field must be at least 1 and at most 20 characters long"},
	}
	performValidationTest(t, &a, expected, 0)
}

func TestValidateNicknameOnlySpecials(t *testing.T) {
	docs.Description("an attendee with an invalid nickname reports a validation error")
	performNicknameValidationTest(t, "}:{")
}

func TestValidateNicknameTooManySpecials1(t *testing.T) {
	docs.Description("an attendee with an invalid nickname reports a validation error")
	performNicknameValidationTest(t, "}super:friendly{")
}

func TestValidateNicknameTooManySpecials2(t *testing.T) {
	docs.Description("an attendee with an invalid nickname reports a validation error")
	performNicknameValidationTest(t, "suPer8friendly99")
}

func performNicknameValidationTest(t *testing.T, wrongNick string) {
	a := tstCreateValidAttendee()
	a.Nickname = wrongNick

	expected := url.Values{
		"nickname": []string{"nickname field must contain at least two letters, and contain no more than two non-letters"},
	}
	performValidationTest(t, &a, expected, 0)
}

func TestValidateBirthday1(t *testing.T) {
	docs.Description("an attendee with an invalid date of birth reports a validation error")
	performBirthdayValidationTest(t, "2022-02-29")
}

func TestValidateBirthday2(t *testing.T) {
	docs.Description("an attendee with an invalid date of birth reports a validation error")
	performBirthdayValidationTest(t, "completely-absurd-date")
}

func TestValidateBirthday3(t *testing.T) {
	docs.Description("an attendee with an invalid date of birth reports a validation error")
	performBirthdayValidationTest(t, "1914-13-48")
}

func performBirthdayValidationTest(t *testing.T, wrongDate string) {
	a := tstCreateValidAttendee()
	a.Birthday = wrongDate

	expected := url.Values{
		"birthday": []string{"birthday field must be a valid ISO 8601 date (format yyyy-MM-dd)"},
	}
	performValidationTest(t, &a, expected, 0)
}

func TestValidateChoiceFieldsAndId(t *testing.T) {
	docs.Description("an attendee with invalid values for the choice fields reports the expected validation errors")
	a := tstCreateValidAttendee()
	a.Id = "16"
	a.Gender = "348trhkuth4uihgkj4h89"
	a.Options = "music,awoo"
	a.Flags = "hc,noflag"
	a.Packages = "helicopterflight,boattour,room-none"
	a.TshirtSize = "micro"
	a.Telegram = "iforgotthe_at_atthebeginning"

	expected := url.Values{
		"gender":      []string{"optional gender field must be one of male, female, other, notprovided, or it can be left blank, which counts as notprovided"},
		"options":     []string{"options field must be a comma separated combination of any of anim,art,music,suit"},
		"flags":       []string{"flags field must be a comma separated combination of any of anon,ev,hc"},
		"packages":    []string{"packages field must be a comma separated combination of any of attendance,room-none,sponsor,sponsor2,stage"},
		"telegram":    []string{"optional telegram field must contain your @username from telegram, or it can be left blank"},
		"tshirt_size": []string{"optional tshirt_size field must be empty or one of XS,S,M,L,XL,XXL,XXXL,XXXXL"},
	}
	performValidationTest(t, &a, expected, 16)
}

func TestValidatePreventSettingIdField(t *testing.T) {
	docs.Description("an attendee must not attempt to set its id in the request body")
	a := tstCreateValidAttendee()
	a.Id = "4"

	expected := url.Values{
		"id": []string{"id field must be empty or correctly assigned for incoming requests"},
	}
	performValidationTest(t, &a, expected, 0)
}

func TestValidatePreventSettingIdFieldWrongValue(t *testing.T) {
	docs.Description("an attendee must not attempt to set its id in the request body")
	a := tstCreateValidAttendee()
	a.Id = "4"

	expected := url.Values{
		"id": []string{"id field must be empty or correctly assigned for incoming requests"},
	}
	performValidationTest(t, &a, expected, 16)
}

func TestValidateWrongEmailWhitespaceInUsername(t *testing.T) {
	docs.Description("an attendee with whitespace in the username part of the email address must be rejected")
	performEmailValidationTest(t, "white\tspace@mailinator.com")
}

func TestValidateWrongEmailWhitespaceInDomain(t *testing.T) {
	docs.Description("an attendee with whitespace in the domain part of the email address must be rejected")
	performEmailValidationTest(t, "whitespace@mailinator com")
}

func TestValidateWrongEmailMultipleAtSigns(t *testing.T) {
	docs.Description("an attendee with multiple @ signs in the email address must be rejected")
	performEmailValidationTest(t, "a@bb@ccc")
}

func performEmailValidationTest(t *testing.T, wrongEmail string) {
	a := tstCreateValidAttendee()
	a.Email = wrongEmail
	expected := url.Values{"email": []string{"email field is not plausible"}}
	performValidationTest(t, &a, expected, 0)
}

func performValidationTest(t *testing.T, a *attendee.AttendeeDto, expectedErrors url.Values, allowedId uint) {
	actualErrors := validate(context.TODO(), a, &entity.Attendee{Model: gorm.Model{ID: allowedId}})

	prettyprintedActualErrors, _ := json.MarshalIndent(actualErrors, "", "  ")
	prettyprintedExpectedErrors, _ := json.MarshalIndent(expectedErrors, "", "  ")

	if !reflect.DeepEqual(actualErrors, expectedErrors) {
		t.Errorf("Errors were not as expected.\nActual:\n%v\nExpected:\n%v\n", string(prettyprintedActualErrors), string(prettyprintedExpectedErrors))
	}
}
