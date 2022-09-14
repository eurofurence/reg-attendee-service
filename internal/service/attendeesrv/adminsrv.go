package attendeesrv

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
)

func (s *AttendeeServiceImplData) GetAdminInfo(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	// admin authorization is checked in the controller
	// presence of attendeeId is checked in the controller
	adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, attendeeId)
	return adminInfo, err
}

func (s *AttendeeServiceImplData) UpdateAdminInfo(ctx context.Context, adminInfo *entity.AdminInfo) error {
	// admin authorization is checked in the controller
	// presence of attendeeId is checked in the controller
	// controller has called GetAdminInfo before this, so we know ID is set
	err := database.GetRepository().WriteAdminInfo(ctx, adminInfo)
	return err
}
