package attendeesrv

import (
	"context"
	"errors"
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
	DoStatusChange(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string, comments string) error
	StatusChangeAllowed(ctx context.Context, oldStatus string, newStatus string) error
	StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error
}

var (
	SameStatusError          = errors.New("old and new status are the same")
	InsufficientPaymentError = errors.New("payment amount not sufficient")
	HasPaymentBalanceError   = errors.New("there is a non-zero payment balance, please use partially paid, or refund")
	CannotDeleteError        = errors.New("cannot delete attendee for legal reasons (there were payments or invoices)")
	UnknownStatusError       = errors.New("unknown status value - this is a programming error")
)
