package attendeesrv

import "rexis/rexis-go-attendee/internal/entity"

type AttendeeService interface {
	NewAttendee() *entity.Attendee
	RegisterNewAttendee(attendee *entity.Attendee) (uint, error)
	GetAttendee(id uint) (*entity.Attendee, error)
	UpdateAttendee(attendee *entity.Attendee) error
}

