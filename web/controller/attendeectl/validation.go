package attendeectl

import (
	"context"
	"fmt"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
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

const emailPattern = "^[^\\@\\s]+\\@[^\\@\\s]+$"

const countryPattern = "^[A-Z]{2}$"

var allowedGenders = [...]string{"male", "female", "other", "notprovided", ""}

func validate(ctx context.Context, a *attendee.AttendeeDto, trustedOriginalState *entity.Attendee) url.Values {
	errs := url.Values{}

	if a.Id != "" && a.Id != fmt.Sprint(trustedOriginalState.ID) {
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
	validation.CheckLength(&errs, 0, 80, "state", a.State)
	if validation.ViolatesPattern(countryPattern, a.Country) {
		errs.Add("country", "country field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)")
	}
	if validation.ViolatesPattern(countryPattern, a.CountryBadge) {
		errs.Add("country_badge", "country_badge field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)")
	}
	validation.CheckLength(&errs, 1, 200, "email", a.Email)
	if validation.ViolatesPattern(emailPattern, a.Email) {
		errs.Add("email", "email field is not plausible")
	}
	validation.CheckLength(&errs, 1, 32, "phone", a.Phone)
	if validation.ViolatesPattern("^(|@.+)$", a.Telegram) {
		errs.Add("telegram", "optional telegram field must contain your @username from telegram, or it can be left blank")
	}
	validation.CheckLength(&errs, 0, 80, "telegram", a.Telegram)
	if validation.InvalidISODate(a.Birthday) {
		errs.Add("birthday", "birthday field must be a valid ISO 8601 date (format yyyy-MM-dd)")
	} else if validation.DateNotInRangeInclusive(a.Birthday, config.EarliestBirthday(), config.LatestBirthday()) {
		errs.Add("birthday", "birthday must be no earlier than " + config.EarliestBirthday() + " and no later than " + config.LatestBirthday())
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

	// check permission to change flags, packages, options to their new values
	if err := attendeeService.CanChangeChoiceTo(ctx, trustedOriginalState.Flags, a.Flags, config.FlagsConfig()); err != nil {
		errs.Add( "flags", err.Error())
	}
	if err := attendeeService.CanChangeChoiceTo(ctx, trustedOriginalState.Packages, a.Packages, config.PackagesConfig()); err != nil {
		errs.Add( "packages", err.Error())
	}
	if err := attendeeService.CanChangeChoiceTo(ctx, trustedOriginalState.Options, a.Options, config.OptionsConfig()); err != nil {
		errs.Add( "options", err.Error())
	}

	if len(errs) != 0 {
		logger := logging.Ctx(ctx)
		if logger.IsDebugEnabled() {
			for key, val := range errs {
				logger.Debugf("attendee dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}
