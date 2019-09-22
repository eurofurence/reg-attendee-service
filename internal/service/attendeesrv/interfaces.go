package attendeesrv

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
)

type AttendeeService interface {
	NewAttendee(ctx context.Context) *entity.Attendee
	RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error)
	GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error)
	UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error
}

