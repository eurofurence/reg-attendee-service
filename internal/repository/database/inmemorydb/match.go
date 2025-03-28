package inmemorydb

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"github.com/ryanuber/go-glob"
	"slices"
	"strconv"
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
		matchesOverdue(cond.AddInfo, a.CacheDueDate, r.Now().Format(config.IsoDateFormat), st.Status) &&
		matchesIsoDateRange(cond.BirthdayFrom, cond.BirthdayTo, a.Birthday) &&
		matchesIdentitySubjects(cond.IdentitySubjects, a.Identity)
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

func choiceMatch(cond map[string]int8, selectedValues ...string) bool {
	chosen := choiceCountMap(selectedValues...)

	for k, v := range cond {
		count, _ := chosen[k]
		if v == 1 && count == 0 {
			return false
		}
		if v == 0 && count > 0 {
			return false
		}
	}
	return true
}

// choiceCountMap allows passing in multiple dbRepresentations that are
// combined into a single count map.
//
// Used to combine flags and admin flags into a single map.
//
// Each parameter is a comma separated list of choice names, possibly followed
// by :count, where count is a positive integer. If the :count postfix is missing,
// it is treated as a count of 1.
//
// The :count postfix is currently only in use for packages.
func choiceCountMap(dbRepresentations ...string) map[string]int {
	result := make(map[string]int)

	for _, dbRepr := range dbRepresentations {
		value := strings.TrimPrefix(dbRepr, ",")
		value = strings.TrimSuffix(value, ",")

		chosen := strings.Split(value, ",")

		for _, entry := range chosen {
			if entry != "" {
				nameAndPossiblyCount := strings.Split(entry, ":")
				name := nameAndPossiblyCount[0]
				count := 1
				if len(nameAndPossiblyCount) > 1 {
					var err error
					count, err = strconv.Atoi(nameAndPossiblyCount[1])
					if err != nil {
						aulogging.Logger.NoCtx().Warn().Printf("encountered invalid choice entry '%s' in database - ignoring", entry)
						continue
					}
				}
				currentCount, _ := result[name]
				result[name] = currentCount + count
			}
		}
	}

	return result
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

func matchesOverdue(addInfoConds map[string]int8, dueDate string, currDate string, currStatus status.Status) bool {
	cond, ok := addInfoConds["overdue"]
	if !ok {
		return true // no condition given
	}

	if dueDate == "" {
		return false // not in correct status anyway
	}

	if cond == 0 {
		return dueDate >= currDate && (currStatus == status.Approved || currStatus == status.PartiallyPaid)
	} else if cond == 1 {
		return dueDate < currDate && (currStatus == status.Approved || currStatus == status.PartiallyPaid)
	} else {
		return false
	}
}

func matchesIsoDateRange(condFrom string, condTo string, value string) bool {
	if value != "" && condFrom != "" {
		if value < condFrom {
			return false
		}
	}
	if value != "" && condTo != "" {
		if value > condTo {
			return false
		}
	}
	return true
}

func matchesIdentitySubjects(cond []string, value string) bool {
	if len(cond) > 8 {
		cond = cond[:8]
	}
	return len(cond) == 0 || slices.Contains(cond, value)
}
