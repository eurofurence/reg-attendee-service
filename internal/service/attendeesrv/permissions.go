package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
)

func (s *AttendeeServiceImplData) subjectHasAdminPermissionEntry(ctx context.Context, subject string, permission ...string) (bool, error) {
	if subject == "" {
		return false, errors.New("not a logged in user subject - this is an implementation error")
	}
	if len(permission) == 0 {
		return false, errors.New("must provide valid permissions - this is an implementation error")
	}

	// check that any of the registrations owned by subject have the regdesk permission
	ownedAttendees, err := database.GetRepository().FindByIdentity(ctx, subject)
	if err != nil {
		return false, err
	}
	for _, oa := range ownedAttendees {
		adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, oa.ID)
		if err != nil {
			return false, err
		}

		permissions := commaSeparatedStrToMap(adminInfo.Permissions, config.AllowedPermissions())

		for _, perm := range permission {
			if perm == "" {
				return false, errors.New("must provide valid permissions - this is an implementation error")
			}

			allowed, _ := permissions[perm]
			if allowed {
				return true, nil
			}
		}
	}

	return false, nil
}
