package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func ViolatesPattern(pattern string, value string) bool {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return true
	}
	return !matched
}

func CheckLength(errs *url.Values, min int, max int, key string, value string) {
	if len(value) < min || len(value) > max {
		errs.Add(key, fmt.Sprintf("%s field must be at least %d and at most %d characters long", key, min, max))
	}
}

func CheckIntValueRange(errs *url.Values, min int, max int, key string, value int) {
	if value < min || value > max {
		errs.Add(key, fmt.Sprintf("%s field must be an integer at least %d and at most %d", key, min, max))
	}
}

func CheckValidRegexOrEmpty(errs *url.Values, key string, value string) {
	if value != "" {
		_, err := regexp.Compile(value)
		if err != nil {
			errs.Add(key, fmt.Sprintf("%s field must be empty or contain a valid regular expression: %s", key, err.Error()))
		}
	}
}

const isoDateFormat = "2006-01-02"

func InvalidISODate(value string) bool {
	_, err := time.Parse(isoDateFormat, value)
	return err != nil
}

func DateNotInRangeInclusive(value string, earliest string, latest string) bool {
	val, _ := time.Parse(isoDateFormat, value)
	min, _ := time.Parse(isoDateFormat, earliest)
	max, _ := time.Parse(isoDateFormat, latest)
	return val.Before(min) || val.After(max)
}

func NotInAllowedValues(allowed []string, value string) bool {
	return !SliceContains(allowed, value)
}

func CheckCombinationOfAllowedValues(errs *url.Values, allowed []string, key string, commaSeparatedValue string) {
	if commaSeparatedValue == "" {
		return
	}

	chosenValues := strings.Split(commaSeparatedValue, ",")
	ok := true
	for _, v := range chosenValues {
		if NotInAllowedValues(allowed, v) {
			ok = false
		}
	}

	if !ok {
		allowedCommaSeparated := strings.Join(allowed, ",")
		errs.Add(key, fmt.Sprintf("%s field must be a comma separated combination of any of %s", key, allowedCommaSeparated))
	}
}

func SliceContains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
