package attendeesrv

import "github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"

type AttendeeService interface {
	NewAttendee() *entity.Attendee
	RegisterNewAttendee(attendee *entity.Attendee) (uint, error)
	GetAttendee(id uint) (*entity.Attendee, error)
	UpdateAttendee(attendee *entity.Attendee) error
}

