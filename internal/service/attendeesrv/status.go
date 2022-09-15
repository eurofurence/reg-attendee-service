package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"gorm.io/gorm"
)

func (s *AttendeeServiceImplData) GetFullStatusHistory(ctx context.Context, attendee *entity.Attendee) ([]entity.StatusChange, error) {
	// controller checks permissions

	result := make([]entity.StatusChange, 0)
	if attendee.ID == 0 {
		return result, errors.New("invalid attendee missing id, please read full dataset from the database - this is an implementation error")
	}

	fromDb, err := database.GetRepository().GetStatusChangesByAttendeeId(ctx, attendee.ID)
	if err != nil {
		return result, err
	}

	// first status entry comes from registration time, not stored in db for performance reasons during initial reg
	result = append(result, entity.StatusChange{
		Model: gorm.Model{
			CreatedAt: attendee.CreatedAt,
		},
		AttendeeId: attendee.ID,
		Status:     "new",
		Comments:   "registration",
	})

	for _, change := range fromDb {
		result = append(result, change)
	}

	return result, nil
}

func (s *AttendeeServiceImplData) RequestStatusChange(ctx context.Context, attendee *entity.Attendee, newStatus string, comments string) error {
	//TODO implement me
	panic("implement me")
}
