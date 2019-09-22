package attendeectl

import (
	"net/url"
	"github.com/jumpy-squirrel/rexis-go-attendee/api/v1/attendee"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/util/validation"
	"strings"
)

const nicknamePattern = "^(" +
// cases where the non-letters are separated by at least one letter
// (non letters are optional, so these also cover all cases of 0 or 1 non-letter)
	"[A-Za-z]+[^A-Za-z]?[A-Za-z]+[^A-Za-z]?[A-Za-z]*" +
	"|[^A-Za-z]?[A-Za-z]+[^A-Za-z]?[A-Za-z]+" +
	"|[^A-Za-z]?[A-Za-z][A-Za-z]+[^A-Za-z]?" +
// cases where the non-letters stand together
	"|[^A-Za-z]{1,2}[A-Za-z][A-Za-z]+" +
	"|[A-Za-z]+[^A-Za-z]{1,2}[A-Za-z]+" +
	"|[A-Za-z][A-Za-z]+[^A-Za-z]{1,2}" +
	")$"

const countryPattern = "^[A-Z]{2}$"

var allowedGenders = [...]string{"male", "female", "other", "notprovided", ""}

func validate(a *attendee.AttendeeDto, allowedId string) url.Values {
	errs := url.Values{}

	if a.Id != "" && a.Id != allowedId {
		errs.Add("id", "id field must be empty or correctly assigned for incoming requests")
	}
	if validation.ViolatesPattern(nicknamePattern, a.Nickname) {
		errs.Add("nickname", "nickname field must contain at least two letters, and contain no more than two non-letters")
	}
	validation.CheckLength(&errs, 2, 80, "nickname", a.Nickname)
	validation.CheckLength(&errs, 1, 80, "first_name", a.FirstName)
	validation.CheckLength(&errs, 1, 80, "last_name", a.LastName)
	validation.CheckLength(&errs, 1, 120, "street", a.Street)
	validation.CheckLength(&errs, 1, 20, "zip", a.Zip)
	validation.CheckLength(&errs, 1, 80, "city", a.City)
	if validation.ViolatesPattern(countryPattern, a.Country) {
		errs.Add("country", "country field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)")
	}
	if validation.ViolatesPattern(countryPattern, a.CountryBadge) {
		errs.Add("country_badge", "country_badge field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)")
	}
	validation.CheckLength(&errs, 1, 200, "email", a.Email)
	validation.CheckLength(&errs, 1, 32, "phone", a.Phone)
	if validation.ViolatesPattern("^(|@.+)$", a.Telegram) {
		errs.Add("telegram", "optional telegram field must contain your @username from telegram, or it can be left blank")
	}
	if validation.InvalidISODate(a.Birthday) {
		errs.Add("birthday", "birthday field must be a valid ISO 8601 date (format yyyy-MM-dd)")
	}
	if validation.NotInAllowedValues(allowedGenders[:], a.Gender) {
		errs.Add("gender", "optional gender field must be one of male, female, other, notprovided, or it can be left blank, which counts as notprovided")
	}
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedFlags(), "flags", a.Flags)
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedPackages(), "packages", a.Packages)
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedOptions(), "options", a.Options)
	if a.TshirtSize != "" && validation.NotInAllowedValues(config.AllowedTshirtSizes(), a.TshirtSize) {
		errs.Add("tshirt_size", "optional tshirt_size field must be empty or one of " + strings.Join(config.AllowedTshirtSizes(), ","))
	}

	return errs
}
