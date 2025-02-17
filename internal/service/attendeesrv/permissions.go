package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
)

func (s *AttendeeServiceImplData) subjectHasAreaPermissionEntry(ctx context.Context, subject string, areas ...string) (bool, error) {
	if subject == "" {
		return false, errors.New("not a logged in user subject - this is an implementation error")
	}
	if len(areas) == 0 {
		return false, errors.New("must provide valid areas - this is an implementation error")
	}

	// check that any of the registrations owned by subject have one of the permissions configured for any of the areas
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

		for _, area := range areas {
			if area == "" {
				return false, errors.New("must provide valid area - this is an implementation error")
			}

			conf := config.AdditionalInfoConfiguration(area)
			for _, perm := range conf.Permissions {
				allowed, _ := permissions[perm]
				if allowed {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (s *AttendeeServiceImplData) subjectHasDirectPermissionEntry(ctx context.Context, subject string, grantingPermissions ...string) (bool, error) {
	if subject == "" {
		return false, errors.New("not a logged in user subject - this is an implementation error")
	}
	if len(grantingPermissions) == 0 {
		return false, errors.New("must provide valid permissions - this is an implementation error")
	}

	// check that any of the registrations owned by subject has one of the granting permissions
	ownedAttendees, err := database.GetRepository().FindByIdentity(ctx, subject)
	if err != nil {
		return false, err
	}
	for _, oa := range ownedAttendees {
		adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, oa.ID)
		if err != nil {
			return false, err
		}

		grantedPermissions := commaSeparatedStrToMap(adminInfo.Permissions, config.AllowedPermissions())

		for _, perm := range grantingPermissions {
			if perm == "" {
				return false, errors.New("must provide valid permission - this is an implementation error")
			}

			allowed, _ := grantedPermissions[perm]
			if allowed {
				return true, nil
			}
		}
	}

	return false, nil
}
