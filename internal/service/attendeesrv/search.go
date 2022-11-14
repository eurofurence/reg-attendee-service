package attendeesrv

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"strings"
)

func (s *AttendeeServiceImplData) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) (*attendee.AttendeeSearchResultList, error) {
	atts, err := database.GetRepository().FindAttendees(ctx, criteria)
	return s.mapToAttendeeSearchResults(atts), err
}

func (s *AttendeeServiceImplData) mapToAttendeeSearchResults(atts []*entity.Attendee) *attendee.AttendeeSearchResultList {
	result := attendee.AttendeeSearchResultList{
		Attendees: make([]attendee.AttendeeSearchResult, len(atts)),
	}
	for i, att := range atts {
		result.Attendees[i] = s.mapToAttendeeSearchResult(att)
	}

	return &result
}

func (s *AttendeeServiceImplData) mapToAttendeeSearchResult(att *entity.Attendee) attendee.AttendeeSearchResult {
	// TODO field visibilities
	// TODO missing information - status
	// TODO missing information - dues
	return attendee.AttendeeSearchResult{
		Id:             att.ID,
		BadgeId:        s.badgeId(att.ID),
		Nickname:       &att.Nickname,
		FirstName:      &att.FirstName,
		LastName:       &att.LastName,
		Street:         &att.Street,
		Zip:            &att.Zip,
		City:           &att.City,
		Country:        &att.Country,
		CountryBadge:   &att.CountryBadge,
		State:          &att.State,
		Email:          &att.Email,
		Phone:          &att.Phone,
		Telegram:       &att.Telegram,
		Partner:        &att.Partner,
		Birthday:       &att.Birthday,
		Gender:         &att.Gender,
		Pronouns:       &att.Pronouns,
		TshirtSize:     &att.TshirtSize,
		Flags:          s.removeWrappingCommas(att.Flags),
		Options:        s.removeWrappingCommas(att.Options),
		Packages:       s.removeWrappingCommas(att.Packages),
		UserComments:   &att.UserComments,
		Status:         nil,
		TotalDues:      nil,
		PaymentBalance: nil,
		CurrentDues:    nil,
		DueDate:        nil,
	}
}

func (s *AttendeeServiceImplData) badgeId(id uint) *string {
	// TODO implement checksum character
	checksum := "Y"
	result := fmt.Sprintf("%d%s", id, checksum)
	return &result
}

func (s *AttendeeServiceImplData) removeWrappingCommas(v string) *string {
	v = strings.TrimPrefix(v, ",")
	v = strings.TrimSuffix(v, ",")
	return &v
}
