package database

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
)

type Repository interface {
	Open()
	Close()

	AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error)
	UpdateAttendee(ctx context.Context, a *entity.Attendee) error
	GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error)
}
