package inmemorydb

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"github.com/ryanuber/go-glob"
	"strings"
)

func matchesCriteria(conds *attendee.AttendeeSearchCriteria, a *entity.Attendee, adm *entity.AdminInfo, st *entity.StatusChange) bool {
	if conds != nil {
		if conds.MinId > 0 && a.ID < conds.MinId {
			return false
		}
		if conds.MaxId > 0 && a.ID > conds.MaxId {
			return false
		}
		for _, cond := range conds.MatchAny {
			if matches(&cond, a, adm, st) {
				return true
			}
		}
	}
	return false
}

func matches(cond *attendee.AttendeeSearchSingleCriterion, a *entity.Attendee, adm *entity.AdminInfo, st *entity.StatusChange) bool {
	return matchesUintSliceOrEmpty(cond.Ids, a.ID) &&
		matchesFullstringGlobOrEmpty(cond.Nickname, a.Nickname) &&
		matchesFullstringGlobOrEmpty(cond.Name, a.FirstName+" "+a.LastName) &&
		matchesSubstringGlobOrEmpty(cond.Address, a.Street+" "+a.Zip+" "+a.City+" "+a.State) &&
		matchesExactOrEmpty(cond.Country, a.Country) &&
		matchesExactOrEmpty(cond.CountryBadge, a.CountryBadge) &&
		matchesSubstringGlobOrEmpty(cond.Email, a.Email) &&
		matchesSubstringGlobOrEmpty(cond.Telegram, a.Telegram) &&
		choiceMatch(cond.Flags, a.Flags, adm.Flags) &&
		choiceMatch(cond.Options, a.Options) &&
		choiceMatch(cond.Packages, a.Packages) &&
		matchesSubstringGlobOrEmpty(cond.UserComments, a.UserComments) &&
		matchesStatus(cond.Status, st.Status) &&
		choiceMatch(cond.Permissions, adm.Permissions) &&
		matchesSubstringGlobOrEmpty(cond.AdminComments, adm.AdminComments)
}

func matchesUintSliceOrEmpty(cond []uint, value uint) bool {
	return len(cond) == 0 || validation.SliceContains(cond, value)
}

func matchesFullstringGlobOrEmpty(cond string, value string) bool {
	return cond == "" || glob.Glob(cond, value)
}

func matchesSubstringGlobOrEmpty(cond string, value string) bool {
	return cond == "" || glob.Glob("*"+cond+"*", value)
}

func matchesExactOrEmpty(cond string, value string) bool {
	return cond == "" || cond == value
}

func choiceMatch(cond map[string]int8, rawValues ...string) bool {
	combined := ""
	for _, rawValue := range rawValues {
		value := strings.TrimPrefix(rawValue, ",")
		value = strings.TrimSuffix(value, ",")
		combined = combined + value + ","
	}
	combined = strings.TrimSuffix(combined, ",")

	chosen := strings.Split(combined, ",")

	for k, v := range cond {
		contained := validation.SliceContains(chosen, k)
		if v == 1 && !contained {
			return false
		}
		if v == 0 && contained {
			return false
		}
	}
	return true
}

func matchesStatus(wanted []status.Status, value status.Status) bool {
	if len(wanted) == 0 {
		// default: all except deleted
		wanted = []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Waiting, status.Cancelled}
	}
	for _, w := range wanted {
		if value == w {
			return true
		}
	}
	return false
}
