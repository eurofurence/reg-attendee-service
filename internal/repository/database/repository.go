package database

import (
	"rexis/rexis-go-attendee/internal/entity"
)

type Repository interface {
	AddAttendee(a entity.Attendee) (uint32, error)
	UpdateAttendee(a entity.Attendee) error
	GetAttendeeById(id uint32) (entity.Attendee, error)
}
