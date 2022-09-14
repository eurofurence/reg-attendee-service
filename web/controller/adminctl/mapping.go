package adminctl

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
)

func mapDtoToAdminInfo(dto *admin.AdminInfoDto, a *entity.AdminInfo) {
	// this cannot currently fail
	a.Flags = dto.Flags
	a.Permissions = dto.Permissions
	a.AdminComments = dto.AdminComments
}

func mapAdminInfoToDto(a *entity.AdminInfo, dto *admin.AdminInfoDto) {
	// this cannot fail
	dto.Id = fmt.Sprint(a.ID)
	dto.Flags = a.Flags
	dto.Permissions = a.Permissions
	dto.AdminComments = a.AdminComments
}
