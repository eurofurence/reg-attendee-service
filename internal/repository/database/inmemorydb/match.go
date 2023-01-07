package inmemorydb

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"github.com/ryanuber/go-glob"
	"strings"
)

func (r *InMemoryRepository) matchesCriteria(conds *attendee.AttendeeSearchCriteria, a *entity.Attendee, adm *entity.AdminInfo, st *entity.StatusChange, addInf map[string]*entity.AdditionalInfo) bool {
	if conds != nil {
		if conds.MinId > 0 && a.ID < conds.MinId {
			return false
		}
		if conds.MaxId > 0 && a.ID > conds.MaxId {
			return false
		}
		for _, cond := range conds.MatchAny {
			if r.matches(&cond, a, adm, st, addInf) {
				return true
			}
		}
	}
	return false
}

func (r *InMemoryRepository) matches(cond *attendee.AttendeeSearchSingleCriterion, a *entity.Attendee, adm *entity.AdminInfo, st *entity.StatusChange, addInf map[string]*entity.AdditionalInfo) bool {
	return matchesUintSliceOrEmpty(cond.Ids, a.ID) &&
		matchesFullstringGlobOrEmpty(cond.Nickname, a.Nickname) &&
		matchesFullstringGlobOrEmpty(cond.Name, a.FirstName+" "+a.LastName) &&
		matchesSubstringGlobOrEmpty(cond.Address, a.Street+" "+a.Zip+" "+a.City+" "+a.State) &&
		matchesExactOrEmpty(cond.Country, a.Country) &&
		matchesSubstringGlobOrEmpty(cond.Email, a.Email) &&
		matchesSubstringGlobOrEmpty(cond.Telegram, a.Telegram) &&
		choiceMatch(cond.SpokenLanguages, a.SpokenLanguages) &&
		choiceMatch(cond.RegistrationLanguage, a.RegistrationLanguage) &&
		choiceMatch(cond.Flags, a.Flags, adm.Flags) &&
		choiceMatch(cond.Options, a.Options) &&
		choiceMatch(cond.Packages, a.Packages) &&
		matchesSubstringGlobOrEmpty(cond.UserComments, a.UserComments) &&
		matchesStatus(cond.Status, st.Status) &&
		choiceMatch(cond.Permissions, adm.Permissions) &&
		matchesSubstringGlobOrEmpty(cond.AdminComments, adm.AdminComments) &&
		matchesAddInfoPresence(cond.AddInfo, addInf) &&
		matchesOverdue(cond.AddInfo, a.CacheDueDate, r.Now().Format(config.IsoDateFormat))
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

func matchesAddInfoPresence(cond map[string]int8, values map[string]*entity.AdditionalInfo) bool {
	for key, wanted := range cond {
		if key != "overdue" {
			_, ok := values[key]
			if wanted == 0 && ok {
				return false
			}
			if wanted == 1 && !ok {
				return false
			}
		}
	}
	return true
}

func matchesOverdue(addInfoConds map[string]int8, dueDate string, currDate string) bool {
	cond, ok := addInfoConds["overdue"]
	if !ok {
		return true // no condition given
	}

	if cond == 0 {
		return dueDate >= currDate
	} else if cond == 1 {
		return dueDate < currDate
	} else {
		return false
	}
}
