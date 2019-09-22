package attendeesrv

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database"
)

type AttendeeServiceImplData struct {
}

func (s *AttendeeServiceImplData) NewAttendee() *entity.Attendee {
	return &entity.Attendee{}
}

func (s *AttendeeServiceImplData) RegisterNewAttendee(attendee *entity.Attendee) (uint, error) {
	id, err := database.GetRepository().AddAttendee(attendee)
	return id, err
}

func (s *AttendeeServiceImplData) GetAttendee(id uint) (*entity.Attendee, error) {
	attendee, err := database.GetRepository().GetAttendeeById(id)
	return attendee, err
}

func (s *AttendeeServiceImplData) UpdateAttendee(attendee *entity.Attendee) error {
	err := database.GetRepository().UpdateAttendee(attendee)
	return err
}
