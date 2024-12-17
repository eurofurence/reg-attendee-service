package attendeesrv

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"sort"
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
			"total_dues", "payment_balance", "current_dues", "due_date", "registered", "admin_comments", "avatar"}
	}

	var currentDues = att.CacheTotalDues - att.CachePaymentBalance
	var registered = att.CreatedAt.Format(config.IsoDateFormat)
	spokenLanguages := removeWrappingCommas(att.SpokenLanguages)
	mergedFlags := removeWrappingCommasJoin(att.Flags, att.AdminFlags)
	options := removeWrappingCommas(att.Options)
	packagesList := sortedPackageListFromCommaSeparatedWithCounts(removeWrappingCommas(att.Packages))
	packages := packagesFromPackagesList(packagesList)
	identity := ""
	if att.Status != status.Deleted {
		identity = att.Identity
	}
	avatar := att.Avatar
	if avatar != "" {
		avatar = config.AvatarBaseUrl() + avatar
	}
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
		SpokenLanguages:      contains(p(spokenLanguages), fillFields, "all", "contact", "spoken_languages"),
		SpokenLanguagesList:  containsSlice(listFromCommaSeparated(spokenLanguages), fillFields, "all", "contact", "spoken_languages"),
		RegistrationLanguage: contains(p(removeWrappingCommas(att.RegistrationLanguage)), fillFields, "all", "configuration", "registration_language"),
		Flags:                contains(p(mergedFlags), fillFields, "all", "configuration", "flags"),
		FlagsList:            containsSlice(sortedListFromCommaSeparated(mergedFlags), fillFields, "all", "configuration", "flags"),
		Options:              contains(p(options), fillFields, "all", "configuration", "options"),
		OptionsList:          containsSlice(sortedListFromCommaSeparated(options), fillFields, "all", "configuration", "options"),
		Packages:             contains(p(packages), fillFields, "all", "configuration", "packages"),
		PackagesList:         containsSlice(packagesList, fillFields, "all", "configuration", "packages"),
		UserComments:         contains(n(att.UserComments), fillFields, "all", "user_comments"),
		Status:               contains(&att.Status, fillFields, "all", "status"),
		TotalDues:            contains(&att.CacheTotalDues, fillFields, "all", "balances", "total_dues"),
		PaymentBalance:       contains(&att.CachePaymentBalance, fillFields, "all", "balances", "payment_balance"),
		CurrentDues:          contains(&currentDues, fillFields, "all", "balances", "current_dues"),
		DueDate:              contains(n(att.CacheDueDate), fillFields, "all", "balances", "due_date"),
		Registered:           contains(n(registered), fillFields, "all", "registered"),
		AdminComments:        contains(n(att.AdminComments), fillFields, "all", "admin_comments"),
		IdentitySubject:      contains(n(identity), fillFields, "all", "identity_subject"),
		Avatar:               contains(n(avatar), fillFields, "all", "avatar"),
	}
}

var checksumLetters = strings.Split("FJQCEKNTWLVGYHSZXDBUARP", "") // 23 letters (prime)

var checksumWeights = [5]int{3, 7, 11, 13, 17}

func calculateChecksum(id int) string {
	sum := 0
	place := 0
	for id > 0 && place < 5 {
		digit := id % 10
		sum += digit * checksumWeights[place]
		id /= 10
		place++
	}
	idx := sum % len(checksumLetters)

	return checksumLetters[idx]
}

func (s *AttendeeServiceImplData) badgeId(id uint) *string {
	checksum := calculateChecksum(int(id))
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

func listFromCommaSeparated(v string) []string {
	if v == "" {
		return nil
	}

	result := strings.Split(v, ",")
	return result
}

func sortedListFromCommaSeparated(v string) []string {
	if v == "" {
		return nil
	}

	result := strings.Split(v, ",")
	sort.Strings(result)
	return result
}

func sortedPackageListFromCommaSeparatedWithCounts(v string) []attendee.PackageState {
	if v == "" {
		return nil
	}

	result := make([]attendee.PackageState, 0)

	pkgMap := choiceStrToMapWithoutChecks(v)
	for name, count := range pkgMap {
		if count > 0 {
			result = append(result, attendee.PackageState{
				Name:  name,
				Count: count,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func packagesFromPackagesList(asList []attendee.PackageState) string {
	var result []string
	for _, entry := range asList {
		for i := 0; i < entry.Count; i++ {
			result = append(result, entry.Name)
		}
	}
	return strings.Join(result, ",")
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

func containsSlice[T any](v []T, selected []string, matches ...string) []T {
	for _, m := range matches {
		for _, s := range selected {
			if m == s {
				return v
			}
		}
	}
	return nil
}
