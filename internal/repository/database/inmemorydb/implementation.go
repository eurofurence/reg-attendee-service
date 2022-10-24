package inmemorydb

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"sync/atomic"
)

type InMemoryRepository struct {
	attendees     map[uint]*entity.Attendee
	adminInfo     map[uint]*entity.AdminInfo
	statusChanges map[uint][]entity.StatusChange
	history       map[uint]*entity.History
	idSequence    uint32
}

func Create() dbrepo.Repository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) Open() error {
	r.attendees = make(map[uint]*entity.Attendee)
	r.adminInfo = make(map[uint]*entity.AdminInfo)
	r.statusChanges = make(map[uint][]entity.StatusChange)
	r.history = make(map[uint]*entity.History)
	return nil
}

func (r *InMemoryRepository) Close() {
	r.attendees = nil
	r.adminInfo = nil
	r.statusChanges = nil
	r.history = nil
}

func (r *InMemoryRepository) Migrate() error {
	// nothing to do
	return nil
}

// --- attendee ---

func (r *InMemoryRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	a.ID = newId

	// copy the attendee, so later modifications won't also modify it in the simulated db
	copiedAttendee := *a
	r.attendees[newId] = &copiedAttendee
	return newId, nil
}

func (r *InMemoryRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	if _, ok := r.attendees[a.ID]; ok {
		// copy the attendee, so later modifications won't also modify it in the simulated db
		copiedAttendee := *a
		r.attendees[a.ID] = &copiedAttendee
		return nil
	} else {
		return fmt.Errorf("cannot update attendee %d - not present", a.ID)
	}
}

func (r *InMemoryRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	if att, ok := r.attendees[id]; ok {
		// copy the attendee, so later modifications won't also modify it in the simulated db
		copiedAttendee := *att
		return &copiedAttendee, nil
	} else {
		return &entity.Attendee{}, fmt.Errorf("cannot get attendee %d - not present", id)
	}
}

func (r *InMemoryRepository) CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error) {
	var count int64
	for _, v := range r.attendees {
		if nickname == v.Nickname && zip == v.Zip && email == v.Email {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryRepository) MaxAttendeeId(ctx context.Context) (uint, error) {
	var max uint
	for _, v := range r.attendees {
		if v.ID > max {
			max = v.ID
		}
	}
	return max, nil
}

// --- attendee search ---

// --- admin info ---

func (r *InMemoryRepository) GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	if ai, ok := r.adminInfo[attendeeId]; ok {
		// copy the info, so later modifications won't also modify it in the simulated db
		copiedAdminInfo := *ai
		return &copiedAdminInfo, nil
	} else {
		aiEmpty := entity.AdminInfo{}
		aiEmpty.ID = attendeeId
		return &aiEmpty, nil
	}
}

func (r *InMemoryRepository) WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error {
	if ai.ID == 0 {
		return fmt.Errorf("cannot save admin info for attendee ID 0")
	}

	copiedAdminInfo := *ai
	r.adminInfo[ai.ID] = &copiedAdminInfo
	return nil
}

// --- status changes ---

func (r *InMemoryRepository) GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error) {
	scEmpty := entity.StatusChange{
		AttendeeId: attendeeId,
		Status:     "new",
		Comments:   "",
	}
	if scList, ok := r.statusChanges[attendeeId]; ok {
		if len(scList) > 0 {
			sc := scList[len(scList)-1]
			return &sc, nil
		} else {
			return &scEmpty, nil
		}
	} else {
		return &scEmpty, nil
	}
}

func (r *InMemoryRepository) GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error) {
	if scList, ok := r.statusChanges[attendeeId]; ok {
		scListCopy := make([]entity.StatusChange, len(scList))
		for i := range scList {
			scListCopy[i] = scList[i]
		}
		return scListCopy, nil
	} else {
		return make([]entity.StatusChange, 0), nil
	}
}

func (r *InMemoryRepository) AddStatusChange(ctx context.Context, sc *entity.StatusChange) error {
	if scList, ok := r.statusChanges[sc.AttendeeId]; ok {
		scCopy := *sc
		r.statusChanges[sc.AttendeeId] = append(scList, scCopy)
	} else {
		scCopy := *sc
		r.statusChanges[sc.AttendeeId] = []entity.StatusChange{scCopy}
	}
	return nil
}

func (r *InMemoryRepository) FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error) {
	result := make([]*entity.Attendee, 0)
	for _, a := range r.attendees {
		if a.Identity == identity {
			result = append(result, a)
		}
	}
	return result, nil
}

// --- bans ---

func (r *InMemoryRepository) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	return make([]*entity.Ban, 0), errors.New("TODO - not implemented")
}

func (r *InMemoryRepository) GetBanById(ctx context.Context, id uint) (*entity.Ban, error) {
	return &entity.Ban{}, errors.New("TODO - not implemented")
}

func (r *InMemoryRepository) AddBan(ctx context.Context, b *entity.Ban) (uint, error) {
	return 0, errors.New("TODO - not implemented")
}

func (r *InMemoryRepository) UpdateBan(ctx context.Context, b *entity.Ban) error {
	return errors.New("TODO - not implemented")
}

// --- additional info ---

func (r *InMemoryRepository) GetAdditionalInfoFor(ctx context.Context, attendeeId uint, area string) (*entity.AdditionalInfo, error) {
	return &entity.AdditionalInfo{}, errors.New("TODO - not implemented")
}

func (r *InMemoryRepository) WriteAdditionalInfo(ctx context.Context, ad *entity.AdditionalInfo) error {
	return errors.New("TODO - not implemented")
}

// --- history ---

func (r *InMemoryRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	h.ID = newId
	r.history[newId] = h
	return nil
}

// only offered for testing, and only on the in memory db
func (r *InMemoryRepository) GetHistoryById(ctx context.Context, id uint) (*entity.History, error) {
	if h, ok := r.history[id]; ok {
		return h, nil
	} else {
		return &entity.History{}, fmt.Errorf("cannot get history entry %d - not present", id)
	}
}
