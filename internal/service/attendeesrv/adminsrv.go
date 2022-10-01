package attendeesrv

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
)

func (s *AttendeeServiceImplData) GetAdminInfo(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	// admin authorization is checked in the controller
	// presence of attendeeId is checked in the controller
	adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, attendeeId)
	return adminInfo, err
}

func (s *AttendeeServiceImplData) UpdateAdminInfo(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo) error {
	// admin authorization is checked in the controller
	// presence of attendeeId is checked in the controller
	// controller has called GetAdminInfo before this, so we know ID is set

	err := database.GetRepository().WriteAdminInfo(ctx, adminInfo)
	if err != nil {
		return err
	}

	statusHistory, err := s.GetFullStatusHistory(ctx, attendee)
	if err != nil {
		return err
	}
	currentStatus := statusHistory[len(statusHistory)-1].Status

	// setting admin flags such as guest may change dues, and change status
	subject := ctxvalues.Subject(ctx)
	err = s.UpdateDuesAndDoStatusChangeIfNeeded(ctx, attendee, currentStatus, currentStatus, fmt.Sprintf("admin info update by %s", subject))
	if err != nil {
		return err
	}

	return nil
}
