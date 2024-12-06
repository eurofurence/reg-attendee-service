package attendeectl

import (
	"context"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

const emailPattern = "^[^\\@\\s]+\\@[^\\@\\s]+$"

const countryPattern = "^[A-Z]{2}$"

var allowedGenders = [...]string{"male", "female", "other", "notprovided", ""}

func validateCountry(ctx context.Context, country string) bool {
	for _, c := range config.AllowedCountries() {
		if c == country {
			return true
		}
	}
	return false
}

func validateNickname(errs *url.Values, nickname string) {
	countAlphanumeric := 0
	countNonAlphanumeric := 0

	for _, nickRune := range nickname {
		if unicode.IsDigit(nickRune) || unicode.IsLetter(nickRune) {
			countAlphanumeric++
		} else if unicode.IsSpace(nickRune) {
			// spaces neither count towards alphanumerics nor the non-alphanumeric count
		} else {
			countNonAlphanumeric++
		}
	}

	if countAlphanumeric < 1 {
		errs.Add("nickname", "nickname field must contain at least one alphanumeric character")
	}

	if countNonAlphanumeric > 2 {
		errs.Add("nickname", "nickname field must not contain more than two non-alphanumeric characters (not counting spaces)")
	}

	validation.CheckLength(errs, 1, 80, "nickname", nickname)
}

func validate(ctx context.Context, a *attendee.AttendeeDto, trustedOriginalState *entity.Attendee, currentStatus status.Status) url.Values {
	errs := url.Values{}

	if a.Id != 0 && a.Id != trustedOriginalState.ID {
		errs.Add("id", "id field must be empty or correctly assigned for incoming requests")
	}
	validateNickname(&errs, a.Nickname)
	validation.CheckLength(&errs, 1, 80, "first_name", a.FirstName)
	validation.CheckLength(&errs, 1, 80, "last_name", a.LastName)
	validation.CheckLength(&errs, 1, 120, "street", a.Street)
	validation.CheckLength(&errs, 1, 20, "zip", a.Zip)
	validation.CheckLength(&errs, 1, 80, "city", a.City)
	validation.CheckLength(&errs, 0, 80, "state", a.State)
	validation.CheckLength(&errs, 0, 80, "partner", a.Partner)
	if validation.ViolatesPattern(countryPattern, a.Country) || !validateCountry(ctx, a.Country) {
		errs.Add("country", "country field must contain a 2 letter upper case ISO-3166-1 country code (Alpha-2 code, see https://en.wikipedia.org/wiki/ISO_3166-1)")
	}
	validation.CheckLength(&errs, 1, 200, "email", a.Email)
	if validation.ViolatesPattern(emailPattern, a.Email) {
		errs.Add("email", "email field is not plausible, must match "+emailPattern)
	}
	validation.CheckLength(&errs, 1, 32, "phone", a.Phone)
	if validation.ViolatesPattern("^(|@.+)$", a.Telegram) {
		errs.Add("telegram", "optional telegram field must contain your @username from telegram, or it can be left blank")
	}
	validation.CheckLength(&errs, 0, 80, "telegram", a.Telegram)
	validation.CheckLength(&errs, 0, 40, "pronouns", a.Pronouns)
	if validation.InvalidISODate(a.Birthday) {
		errs.Add("birthday", "birthday field must be a valid ISO 8601 date (format yyyy-MM-dd)")
	} else if validation.DateNotInRangeInclusive(a.Birthday, config.EarliestBirthday(), config.LatestBirthday()) {
		errs.Add("birthday", "birthday must be no earlier than "+config.EarliestBirthday()+" and no later than "+config.LatestBirthday())
	}
	if validation.NotInAllowedValues(allowedGenders[:], a.Gender) {
		errs.Add("gender", "optional gender field must be one of male, female, other, notprovided, or it can be left blank, which counts as notprovided")
	}
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedSpokenLanguages(), "spoken_languages", a.SpokenLanguages)
	if validation.NotInAllowedValues(config.AllowedRegistrationLanguages(), a.RegistrationLanguage) {
		errs.Add("registration_language", "registration_language field must be one of "+strings.Join(config.AllowedRegistrationLanguages(), ",")+" or it can be left blank, which counts as "+config.DefaultRegistrationLanguage())
	}
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedFlagsNoAdmin(), "flags", a.Flags)
	checkPackagesValid(&errs, config.PackagesConfig(), a.Packages)
	checkPackagesListValid(&errs, config.PackagesConfig(), a.PackagesList)
	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedOptions(), "options", a.Options)
	if a.TshirtSize != "" && validation.NotInAllowedValues(config.AllowedTshirtSizes(), a.TshirtSize) {
		errs.Add("tshirt_size", "optional tshirt_size field must be empty or one of "+strings.Join(config.AllowedTshirtSizes(), ","))
	}

	// check permission to change flags, packages, options, email to their new values
	if err := attendeeService.CanChangeChoiceTo(ctx, "flag", trustedOriginalState.Flags, a.Flags, config.FlagsConfigNoAdmin()); err != nil {
		errs.Add("flags", err.Error())
	}
	if err := attendeeService.CanChangeChoiceToCurrentStatus(ctx, "package", trustedOriginalState.Packages, a.Packages, config.PackagesConfig(), currentStatus); err != nil {
		errs.Add("packages", err.Error())
	}
	if err := attendeeService.CanChangeChoiceTo(ctx, "option", trustedOriginalState.Options, a.Options, config.OptionsConfig()); err != nil {
		errs.Add("options", err.Error())
	}
	if err := attendeeService.CanChangeEmailTo(ctx, trustedOriginalState.Email, a.Email); err != nil {
		errs.Add("email", err.Error())
	}

	if err := attendeeService.CanRegisterAtThisTime(ctx); err != nil {
		errs.Add("timing", err.Error())
	}

	if len(errs) != 0 {
		if config.LoggingSeverity() == "DEBUG" {
			logger := aulogging.Logger.Ctx(ctx).Debug()
			for key, val := range errs {
				logger.Printf("attendee dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}

func validateDueDateChange(ctx context.Context, d *attendee.DueDate, trustedOriginalState *entity.Attendee) url.Values {
	errs := url.Values{}

	if validation.InvalidISODate(d.DueDate) {
		errs.Add("due_date", "due date field must be a valid ISO date, e.g. 2023-08-17")
	}

	if len(errs) != 0 {
		if config.LoggingSeverity() == "DEBUG" {
			logger := aulogging.Logger.Ctx(ctx).Debug()
			for key, val := range errs {
				logger.Printf("attendee dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}

func checkPackagesList(errs *url.Values, cfg map[string]config.ChoiceConfig, key string, pkgList []attendee.PackageState) {
	namesOk := true

	for _, v := range pkgList {
		c, ok := cfg[v.Name]
		if !ok {
			namesOk = false
		} else {
			if v.Count > c.MaxCount {
				errs.Add(key, fmt.Sprintf("package %s occurs too many times, can occur at most %d times", v.Name, c.MaxCount))
			}
		}
	}

	if !namesOk {
		allowedCommaSeparated := strings.Join(sortedKeys(cfg), ",")
		if key == "packages" {
			// different error message for packages field
			errs.Add(key, fmt.Sprintf("%s field must be a comma separated combination of any of %s", key, allowedCommaSeparated))
		} else {
			errs.Add(key, fmt.Sprintf("%s can only contain package names %s", key, allowedCommaSeparated))
		}
	}
}

func checkPackagesListValid(errs *url.Values, cfg map[string]config.ChoiceConfig, pkgList []attendee.PackageState) {
	checkPackagesList(errs, cfg, "packages_list", pkgList)
}

func checkPackagesValid(errs *url.Values, cfg map[string]config.ChoiceConfig, commaSeparatedValue string) {
	asList := packageListFromCommaSeparated(commaSeparatedValue)
	checkPackagesList(errs, cfg, "packages", asList)
}

func sortedKeys(choiceMap map[string]config.ChoiceConfig) []string {
	keys := make([]string, len(choiceMap))
	i := 0
	for k := range choiceMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
