package adminctl

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"strings"
)

func mapDtoToAdminInfo(dto *admin.AdminInfoDto, a *entity.AdminInfo) {
	// this cannot currently fail
	a.Flags = addWrappingCommas(dto.Flags)
	a.Permissions = addWrappingCommas(dto.Permissions)
	a.AdminComments = dto.AdminComments
}

func mapAdminInfoToDto(a *entity.AdminInfo, dto *admin.AdminInfoDto) {
	// this cannot fail
	dto.Id = a.ID
	dto.Flags = removeWrappingCommas(a.Flags)
	dto.Permissions = removeWrappingCommas(a.Permissions)
	dto.AdminComments = a.AdminComments
}

func removeWrappingCommas(v string) string {
	v = strings.TrimPrefix(v, ",")
	v = strings.TrimSuffix(v, ",")
	return v
}

func addWrappingCommas(v string) string {
	if !strings.HasPrefix(v, ",") {
		v = "," + v
	}
	if !strings.HasSuffix(v, ",") {
		v = v + ","
	}
	return v
}
