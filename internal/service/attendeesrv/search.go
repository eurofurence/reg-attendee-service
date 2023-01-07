package attendeesrv

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"strings"
)

func (s *AttendeeServiceImplData) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) (*attendee.AttendeeSearchResultList, error) {
	atts, err := database.GetRepository().FindAttendees(ctx, criteria)
	return s.mapToAttendeeSearchResults(atts, criteria.FillFields), err
}

func (s *AttendeeServiceImplData) mapToAttendeeSearchResults(atts []*entity.AttendeeQueryResult, fillFields []string) *attendee.AttendeeSearchResultList {
	result := attendee.AttendeeSearchResultList{
		Attendees: make([]attendee.AttendeeSearchResult, len(atts)),
	}
	for i, att := range atts {
		result.Attendees[i] = s.mapToAttendeeSearchResult(att, fillFields)
	}

	return &result
}

func (s *AttendeeServiceImplData) mapToAttendeeSearchResult(att *entity.AttendeeQueryResult, fillFields []string) attendee.AttendeeSearchResult {
	if len(fillFields) == 0 {
		fillFields = []string{"nickname", "name", "country", "spoken_languages", "email", "telegram", "birthday", "pronouns",
			"tshirt_size", "flags", "options", "packages", "user_comments", "status",
			"total_dues", "payment_balance", "current_dues", "due_date", "registered", "admin_comments"}
	}

	var currentDues = att.CacheTotalDues - att.CachePaymentBalance
	var registered = att.CreatedAt.Format(config.IsoDateFormat)
	return attendee.AttendeeSearchResult{
		Id:                   att.ID,
		BadgeId:              s.badgeId(att.ID),
		Nickname:             contains(p(att.Nickname), fillFields, "all", "nickname"),
		FirstName:            contains(p(att.FirstName), fillFields, "all", "name", "first_name"),
		LastName:             contains(p(att.LastName), fillFields, "all", "name", "last_name"),
		Street:               contains(p(att.Street), fillFields, "all", "address", "street"),
		Zip:                  contains(p(att.Zip), fillFields, "all", "address", "zip"),
		City:                 contains(p(att.City), fillFields, "all", "address", "city"),
		Country:              contains(p(att.Country), fillFields, "all", "address", "country"),
		State:                contains(p(att.State), fillFields, "all", "address", "state"),
		Email:                contains(p(att.Email), fillFields, "all", "contact", "email"),
		Phone:                contains(p(att.Phone), fillFields, "all", "contact", "phone"),
		Telegram:             contains(p(att.Telegram), fillFields, "all", "contact", "telegram"),
		Partner:              contains(n(att.Partner), fillFields, "all", "partner"),
		Birthday:             contains(p(att.Birthday), fillFields, "all", "birthday"),
		Gender:               contains(n(att.Gender), fillFields, "all", "gender"),
		Pronouns:             contains(n(att.Pronouns), fillFields, "all", "pronouns"),
		TshirtSize:           contains(n(att.TshirtSize), fillFields, "all", "tshirt_size"),
		SpokenLanguages:      contains(p(removeWrappingCommas(att.SpokenLanguages)), fillFields, "all", "contact", "spoken_languages"),
		RegistrationLanguage: contains(p(removeWrappingCommas(att.RegistrationLanguage)), fillFields, "all", "configuration", "registration_language"),
		Flags:                contains(p(removeWrappingCommasJoin(att.Flags, att.AdminFlags)), fillFields, "all", "configuration", "flags"),
		Options:              contains(p(removeWrappingCommas(att.Options)), fillFields, "all", "configuration", "options"),
		Packages:             contains(p(removeWrappingCommas(att.Packages)), fillFields, "all", "configuration", "packages"),
		UserComments:         contains(n(att.UserComments), fillFields, "all", "user_comments"),
		Status:               contains(&att.Status, fillFields, "all", "status"),
		TotalDues:            contains(&att.CacheTotalDues, fillFields, "all", "balances", "total_dues"),
		PaymentBalance:       contains(&att.CachePaymentBalance, fillFields, "all", "balances", "payment_balance"),
		CurrentDues:          contains(&currentDues, fillFields, "all", "balances", "current_dues"),
		DueDate:              contains(n(att.CacheDueDate), fillFields, "all", "balances", "due_date"),
		Registered:           contains(n(registered), fillFields, "all", "registered"),
		AdminComments:        contains(n(att.AdminComments), fillFields, "all", "admin_comments"),
	}
}

func (s *AttendeeServiceImplData) badgeId(id uint) *string {
	// TODO implement checksum character
	checksum := "Y"
	result := fmt.Sprintf("%d%s", id, checksum)
	return &result
}

func removeWrappingCommas(v string) string {
	v = strings.TrimPrefix(v, ",")
	v = strings.TrimSuffix(v, ",")
	return v
}

func removeWrappingCommasJoin(v1 string, v2 string) string {
	v1 = removeWrappingCommas(v1)
	v2 = removeWrappingCommas(v2)
	v := v1 + "," + v2
	return removeWrappingCommas(v)
}

// n formats an optional field (rendered as missing if unset)
func n(v string) *string {
	if v == "" {
		return nil
	} else {
		return &v
	}
}

// p formats a mandatory field (rendered as an empty string if unset, which should never happen anyway)
func p(v string) *string {
	return &v
}

func contains[T any](v *T, selected []string, matches ...string) *T {
	for _, m := range matches {
		for _, s := range selected {
			if m == s {
				return v
			}
		}
	}
	return nil
}
