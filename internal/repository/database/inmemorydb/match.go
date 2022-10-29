package inmemorydb

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
)

func matchesCriteria(conds *attendee.AttendeeSearchCriteria, a *entity.Attendee) bool {
	if conds != nil {
		if conds.MinId > 0 && a.ID < conds.MinId {
			return false
		}
		if conds.MaxId > 0 && a.ID > conds.MaxId {
			return false
		}
		for _, cond := range conds.MatchAny {
			if matches(&cond, a) {
				return true
			}
		}
	}
	return false
}

func matches(cond *attendee.AttendeeSearchSingleCriterion, a *entity.Attendee) bool {
	// TODO
	return false
}
