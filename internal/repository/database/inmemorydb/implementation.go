package inmemorydb

import (
	"context"
	"fmt"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/dbrepo"
	"sync/atomic"
)

type InMemoryRepository struct {
	attendees map[uint]*entity.Attendee
	history map[uint]*entity.History
	idSequence uint32
}

func Create() dbrepo.Repository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) Open() {
	r.attendees = make(map[uint]*entity.Attendee)
	r.history = make(map[uint]*entity.History)
}

func (r *InMemoryRepository) Close() {
	r.attendees = nil
	r.history = nil
}

func (r *InMemoryRepository) Migrate() {
	// nothing to do
}

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
