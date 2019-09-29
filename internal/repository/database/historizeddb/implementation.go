package historizeddb

import (
	"context"
	"errors"
	"github.com/d4l3k/messagediff"
	_ "github.com/d4l3k/messagediff"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/dbrepo"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxvalues"
)

type HistorizingRepository struct {
	wrappedRepository dbrepo.Repository
}

func Create(wrappedRepository dbrepo.Repository) dbrepo.Repository {
	return &HistorizingRepository{wrappedRepository: wrappedRepository}
}

func (r *HistorizingRepository) Open() {
	r.wrappedRepository.Open()
}

func (r *HistorizingRepository) Close() {
	r.wrappedRepository.Close()
}

func (r *HistorizingRepository) Migrate() {
	r.wrappedRepository.Migrate()
}

func (r *HistorizingRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	return r.wrappedRepository.AddAttendee(ctx, a)
}

// we diff reverse so the OLD value is printed in the diffs. The new value is in the database now.
func attendeeDiffReverse(ctx context.Context, oldVersion *entity.Attendee, newVersion *entity.Attendee) *entity.History {
	histEntry := &entity.History{
		Entity:    "Attendee",
		EntityId:  newVersion.ID,
		RequestId: ctxvalues.RequestId(ctx),
		UserId:    0, // TODO: we don't really have user ids yet
	}
	diff, _ := messagediff.PrettyDiff(newVersion, oldVersion)
	histEntry.Diff = diff
	return histEntry
}

func (r *HistorizingRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	oldVersion, err := r.wrappedRepository.GetAttendeeById(ctx, a.ID)
	if err != nil {
		return err
	}

	histEntry := attendeeDiffReverse(ctx, oldVersion, a)

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

// it is an error to call this from the outside. From the inside use wrappedRepository.RecordHistory to bypass the error
func (r *HistorizingRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	return errors.New("not allowed to directly manipulate history")
}
