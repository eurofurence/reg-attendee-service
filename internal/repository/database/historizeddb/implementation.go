package historizeddb

import (
	"context"
	"errors"
	"github.com/d4l3k/messagediff"
	_ "github.com/d4l3k/messagediff"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"gorm.io/gorm"
)

type HistorizingRepository struct {
	wrappedRepository dbrepo.Repository
}

func Create(wrappedRepository dbrepo.Repository) dbrepo.Repository {
	return &HistorizingRepository{wrappedRepository: wrappedRepository}
}

func (r *HistorizingRepository) Open() error {
	return r.wrappedRepository.Open()
}

func (r *HistorizingRepository) Close() {
	r.wrappedRepository.Close()
}

func (r *HistorizingRepository) Migrate() error {
	return r.wrappedRepository.Migrate()
}

// --- attendee ---

func (r *HistorizingRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	return r.wrappedRepository.AddAttendee(ctx, a)
}

func (r *HistorizingRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	oldVersion, err := r.wrappedRepository.GetAttendeeById(ctx, a.ID)
	if err != nil {
		return err
	}

	// hide always present diff in times
	oldVersion.CreatedAt = a.CreatedAt
	oldVersion.UpdatedAt = a.UpdatedAt

	histEntry := diffReverse(ctx, oldVersion, a, "Attendee", a.ID)

	err = r.wrappedRepository.RecordHistory(ctx, histEntry)
	if err != nil {
		return err
	}

	return r.wrappedRepository.UpdateAttendee(ctx, a)
}

func (r *HistorizingRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	return r.wrappedRepository.GetAttendeeById(ctx, id)
}

func (r *HistorizingRepository) CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error) {
	return r.wrappedRepository.CountAttendeesByNicknameZipEmail(ctx, nickname, zip, email)
}

func (r *HistorizingRepository) CountAttendeesByIdentity(ctx context.Context, identity string) (int64, error) {
	return r.wrappedRepository.CountAttendeesByIdentity(ctx, identity)
}

func (r *HistorizingRepository) MaxAttendeeId(ctx context.Context) (uint, error) {
	return r.wrappedRepository.MaxAttendeeId(ctx)
}

// --- attendee search ---

func (r *HistorizingRepository) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) ([]*entity.AttendeeQueryResult, error) {
	return r.wrappedRepository.FindAttendees(ctx, criteria)
}

// --- admin info ---

func (r *HistorizingRepository) GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	return r.wrappedRepository.GetAdminInfoByAttendeeId(ctx, attendeeId)
}

func (r *HistorizingRepository) WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error {
	oldVersion, err := r.wrappedRepository.GetAdminInfoByAttendeeId(ctx, ai.ID)
	if err != nil {
		return err
	}

	// hide always present diff in times
	oldVersion.CreatedAt = ai.CreatedAt
	oldVersion.UpdatedAt = ai.UpdatedAt

	histEntry := diffReverse(ctx, oldVersion, ai, "AdminInfo", ai.ID)

	err = r.wrappedRepository.RecordHistory(ctx, histEntry)
	if err != nil {
		return err
	}

	return r.wrappedRepository.WriteAdminInfo(ctx, ai)
}

// --- status changes ---

func (r *HistorizingRepository) GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error) {
	return r.wrappedRepository.GetLatestStatusChangeByAttendeeId(ctx, attendeeId)
}

func (r *HistorizingRepository) GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error) {
	return r.wrappedRepository.GetStatusChangesByAttendeeId(ctx, attendeeId)
}

func (r *HistorizingRepository) AddStatusChange(ctx context.Context, sc *entity.StatusChange) error {
	// status changes are only appended, so we don't need history
	return r.wrappedRepository.AddStatusChange(ctx, sc)
}

func (r *HistorizingRepository) FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error) {
	return r.wrappedRepository.FindByIdentity(ctx, identity)
}

// --- bans ---

func (r *HistorizingRepository) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	return r.wrappedRepository.GetAllBans(ctx)
}

func (r *HistorizingRepository) GetBanById(ctx context.Context, id uint) (*entity.Ban, error) {
	return r.wrappedRepository.GetBanById(ctx, id)
}

func (r *HistorizingRepository) AddBan(ctx context.Context, b *entity.Ban) (uint, error) {
	return r.wrappedRepository.AddBan(ctx, b)
}

func (r *HistorizingRepository) UpdateBan(ctx context.Context, b *entity.Ban) error {
	oldVersion, err := r.wrappedRepository.GetBanById(ctx, b.ID)
	if err != nil {
		return err
	}

	// hide always present diff in times
	oldVersion.CreatedAt = b.CreatedAt
	oldVersion.UpdatedAt = b.UpdatedAt

	histEntry := diffReverse(ctx, oldVersion, b, "Ban", b.ID)

	err = r.wrappedRepository.RecordHistory(ctx, histEntry)
	if err != nil {
		return err
	}

	return r.wrappedRepository.UpdateBan(ctx, b)
}

func (r *HistorizingRepository) DeleteBan(ctx context.Context, b *entity.Ban) error {
	_, err := r.wrappedRepository.GetBanById(ctx, b.ID)
	if err != nil {
		return err
	}

	histEntry := &entity.History{
		Entity:    "Ban",
		EntityId:  b.ID,
		RequestId: ctxvalues.RequestId(ctx),
		Identity:  ctxvalues.Subject(ctx),
		Diff:      "<deleted>",
	}

	err = r.wrappedRepository.RecordHistory(ctx, histEntry)
	if err != nil {
		return err
	}

	return r.wrappedRepository.DeleteBan(ctx, b)
}

// --- additional info ---

func (r *HistorizingRepository) GetAdditionalInfoFor(ctx context.Context, attendeeId uint, area string) (*entity.AdditionalInfo, error) {
	return r.wrappedRepository.GetAdditionalInfoFor(ctx, attendeeId, area)
}

func (r *HistorizingRepository) WriteAdditionalInfo(ctx context.Context, ad *entity.AdditionalInfo) error {
	oldVersion, err := r.wrappedRepository.GetAdditionalInfoFor(ctx, ad.AttendeeId, ad.Area)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// acceptable situation - first entry needs no history
			return r.wrappedRepository.WriteAdditionalInfo(ctx, ad)
		} else {
			return err
		}
	}

	histEntry := diffReverse(ctx, oldVersion, ad, "AdditionalInfo", ad.ID)

	err = r.wrappedRepository.RecordHistory(ctx, histEntry)
	if err != nil {
		return err
	}

	return r.wrappedRepository.WriteAdditionalInfo(ctx, ad)
}

// --- history ---

// it is an error to call this from the outside. From the inside use wrappedRepository.RecordHistory to bypass the error
func (r *HistorizingRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	return errors.New("not allowed to directly manipulate history")
}

// we diff reverse so the OLD value is printed in the diffs. The new value is in the database now.
func diffReverse[T any](ctx context.Context, oldVersion *T, newVersion *T, entityName string, entityID uint) *entity.History {
	histEntry := &entity.History{
		Entity:    entityName,
		EntityId:  entityID,
		RequestId: ctxvalues.RequestId(ctx),
		Identity:  ctxvalues.Subject(ctx),
	}
	diff, _ := messagediff.PrettyDiff(newVersion, oldVersion)
	histEntry.Diff = diff
	return histEntry
}
