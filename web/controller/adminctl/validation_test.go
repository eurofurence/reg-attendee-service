package adminctl

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
)

func tstCreateValidAdminInfo() admin.AdminInfoDto {
	return admin.AdminInfoDto{
		Id:            "42",
		Flags:         "staff,banned",
		Permissions:   "regdesk,readonly",
		AdminComments: "some admin comment",
	}
}

// TODO test validation
