package dbrepo

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
)

type Repository interface {
	Open() error
	Close()
	Migrate() error

	AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error)
	UpdateAttendee(ctx context.Context, a *entity.Attendee) error
	GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error)
	CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error)
	MaxAttendeeId(ctx context.Context) (uint, error)

	GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error)
	WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error

	// GetLatestStatusChangeByAttendeeId returns the latest status change entry for the given attendee id.
	//
	// If none is in the database, returns a blank (unsaved) change with status "new".
	GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error)
	GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error)
	AddStatusChange(ctx context.Context, sc *entity.StatusChange) error

	FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error)

	RecordHistory(ctx context.Context, h *entity.History) error
}
