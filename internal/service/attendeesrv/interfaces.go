package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

type AttendeeService interface {
	// NewAttendee creates an empty (unsaved) attendee, without an assigned badge number (aka. ID).
	//
	// Mostly useful for filling in values and passing it to RegisterNewAttendee.
	NewAttendee(ctx context.Context) *entity.Attendee

	// RegisterNewAttendee saves a previously unsaved attendee, assigning them a badge number.
	RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error)
	GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error)
	UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error

	// GetAttendeeMaxId returns the highest assigned badge number.
	GetAttendeeMaxId(ctx context.Context) (uint, error)

	CanRegisterAtThisTime(ctx context.Context) error

	CanChangeChoiceTo(ctx context.Context, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error

	GetAdminInfo(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error)
	UpdateAdminInfo(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo) error

	GetFullStatusHistory(ctx context.Context, attendee *entity.Attendee) ([]entity.StatusChange, error)
	// UpdateDuesAndDoStatusChangeIfNeeded updates dues (depending on newStatus) and records a status change if appropriate.
	//
	// If newStatus is one of approved/partially paid/paid, the actual status value written may be any of these three.
	// This is because depending on package and flag changes (guests attend for free!), the dues may change, and
	// so paid may turn into partially paid etc.
	UpdateDuesAndDoStatusChangeIfNeeded(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string, comments string) error
	StatusChangeAllowed(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error
	StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error

	// IsOwnerFor returns the list of attendees (registrations) that are owned by the currently logged
	// in user account.
	//
	// Unless an admin has made changes to the database, this essentially means their registration was made
	// using this account.
	IsOwnerFor(ctx context.Context) ([]*entity.Attendee, error)
}

var (
	SameStatusError          = errors.New("old and new status are the same")
	InsufficientPaymentError = errors.New("payment amount not sufficient")
	HasPaymentBalanceError   = errors.New("there is a non-zero payment balance, please use partially paid, or refund")
	CannotDeleteError        = errors.New("cannot delete attendee for legal reasons (there were payments or invoices)")
	GoToApprovedFirst        = errors.New("please change status to approved, this will automatically advance to (partially) paid as appropriate")
	UnknownStatusError       = errors.New("unknown status value - this is a programming error")
)
