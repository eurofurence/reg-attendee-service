package dbrepo

import (
	"context"

	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
)

type Repository interface {
	Open() error
	Close()
	Migrate() error

	AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error)
	UpdateAttendee(ctx context.Context, a *entity.Attendee) error
	GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error)
	SoftDeleteAttendeeById(ctx context.Context, id uint) error
	UndeleteAttendeeById(ctx context.Context, id uint) error

	CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error)
	CountAttendeesByIdentity(ctx context.Context, identity string) (int64, error)
	MaxAttendeeId(ctx context.Context) (uint, error)

	GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error)
	WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error

	// GetLatestStatusChangeByAttendeeId returns the latest status change entry for the given attendee id.
	//
	// If none is in the database, returns a blank (unsaved) change with status new.
	GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error)
	GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error)
	AddStatusChange(ctx context.Context, sc *entity.StatusChange) error

	FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) ([]*entity.AttendeeQueryResult, error)
	FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error)

	GetAllBans(ctx context.Context) ([]*entity.Ban, error)
	GetBanById(ctx context.Context, id uint) (*entity.Ban, error)
	AddBan(ctx context.Context, b *entity.Ban) (uint, error)
	UpdateBan(ctx context.Context, b *entity.Ban) error
	DeleteBan(ctx context.Context, b *entity.Ban) error

	GetAllAdditionalInfoForArea(ctx context.Context, area string) ([]*entity.AdditionalInfo, error)
	GetAdditionalInfoFor(ctx context.Context, attendeeId uint, area string) (*entity.AdditionalInfo, error)
	WriteAdditionalInfo(ctx context.Context, ad *entity.AdditionalInfo) error

	// CreateCount will create a row in the count table, unless a count for its area and name already exists.
	//
	// If a count for the area and name already exists, the passed in initial counts are ignored and the database
	// row remains untouched.
	//
	// In all cases, the current count is read and returned, which may or may not match the count passed in for creation.
	CreateCount(initial *entity.Count) (*entity.Count, error)

	// AddCount updates the existing counts for an area and name. The update is made as an atomic operation.
	// The counts in delta are allowed to be negative.
	//
	// If no count row exists, this will fail. You should have called CreateCount during database migration.
	//
	// In all cases, the current counts are returned, which may include concurrent updates from other threads,
	// as a read operation is done separately after the update.
	//
	// Note: count updates are not historized.
	AddCount(ctx context.Context, delta *entity.Count) (*entity.Count, error)

	// ResetCount overwrites the current counts for an area and nome with the given values.
	ResetCount(ctx context.Context, overwrite *entity.Count) error

	// GetCount obtains the current count for area and name.
	GetCount(ctx context.Context, area string, name string) (*entity.Count, error)

	RecordHistory(ctx context.Context, h *entity.History) error
}
