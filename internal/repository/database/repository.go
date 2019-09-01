package database

import (
	"rexis/rexis-go-attendee/internal/entity"
)

type Repository interface {
	Open()
	Close()

	AddAttendee(a *entity.Attendee) (uint, error)
	UpdateAttendee(a *entity.Attendee) error
	GetAttendeeById(id uint) (*entity.Attendee, error)
}
