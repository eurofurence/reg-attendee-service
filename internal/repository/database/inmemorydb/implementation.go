package inmemorydb

import (
	"fmt"
	"rexis/rexis-go-attendee/internal/entity"
	"sync/atomic"
)

type InMemoryRepository struct {
	attendees map[uint32]entity.Attendee
	idSequence uint32
}

func (r *InMemoryRepository) AddAttendee(a entity.Attendee) (uint32, error) {
	newId := atomic.AddUint32(&r.idSequence, 1)
	a.Id = newId
	r.attendees[newId] = a
	return newId, nil
}

func (r *InMemoryRepository) UpdateAttendee(a entity.Attendee) error {
	if _, ok := r.attendees[a.Id]; ok {
		r.attendees[a.Id] = a
		return nil
	} else {
		return fmt.Errorf("cannot update attendee %d - not present", a.Id)
	}
}

func (r *InMemoryRepository) GetAttendeeById(id uint32) (entity.Attendee, error) {
	if att, ok := r.attendees[id]; ok {
		return att, nil
	} else {
		return entity.Attendee{}, fmt.Errorf("cannot get attendee %d - not present", id)
	}
}
