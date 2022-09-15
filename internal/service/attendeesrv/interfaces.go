package attendeesrv

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

type AttendeeService interface {
	NewAttendee(ctx context.Context) *entity.Attendee
	RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error)
	GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error)
	UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error
	GetAttendeeMaxId(ctx context.Context) (uint, error)

	CanRegisterAtThisTime(ctx context.Context) error

	CanChangeChoiceTo(ctx context.Context, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error

	GetAdminInfo(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error)
	UpdateAdminInfo(ctx context.Context, adminInfo *entity.AdminInfo) error

	GetFullStatusHistory(ctx context.Context, attendee *entity.Attendee) ([]entity.StatusChange, error)
	RequestStatusChange(ctx context.Context, attendee *entity.Attendee, newStatus string, comments string) error
}
