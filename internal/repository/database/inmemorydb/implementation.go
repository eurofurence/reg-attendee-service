package inmemorydb

import (
	"context"
	"fmt"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"sync/atomic"
)

type InMemoryRepository struct {
	attendees map[uint]*entity.Attendee
	idSequence uint32
}

func (r *InMemoryRepository) Open() {
	r.attendees = make(map[uint]*entity.Attendee)
}

func (r *InMemoryRepository) Close() {
	r.attendees = nil
}

func (r *InMemoryRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	newId := uint(atomic.AddUint32(&r.idSequence, 1))
	a.ID = newId
	r.attendees[newId] = a
	return newId, nil
}

func (r *InMemoryRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	if _, ok := r.attendees[a.ID]; ok {
		r.attendees[a.ID] = a
		return nil
	} else {
		return fmt.Errorf("cannot update attendee %d - not present", a.ID)
	}
}

func (r *InMemoryRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	if att, ok := r.attendees[id]; ok {
		return att, nil
	} else {
		return &entity.Attendee{}, fmt.Errorf("cannot get attendee %d - not present", id)
	}
}
