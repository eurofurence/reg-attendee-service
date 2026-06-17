package attendeesrv

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
)

func (s *AttendeeServiceImplData) GetCurrentUserPermissions(ctx context.Context) (*attendee.UserPermissionsDto, error) {
	// collect OIDC groups from context
	groups := make([]string, 0)
	ctxMap := ctx.Value(ctxvalues.ContextMap)
	if ctxMap != nil {
		for k, v := range ctxMap.(map[string]string) {
			if strings.HasPrefix(k, ctxvalues.ContextAuthorizedAs+"-") && v != "" {
				groups = append(groups, v)
			}
		}
	}
	sort.Strings(groups)

	// collect permissions from all registrations owned by the current user
	subject := ctxvalues.Subject(ctx)
	permissionsSet := make(map[string]bool)
	if subject != "" {
		ownedAttendees, err := database.GetRepository().FindByIdentity(ctx, subject)
		if err != nil {
			return nil, err
		}
		allowedPerms := config.AllowedPermissions()
		for _, oa := range ownedAttendees {
			adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, oa.ID)
			if err != nil {
				return nil, err
			}
			for perm, granted := range commaSeparatedStrToMap(adminInfo.Permissions, allowedPerms) {
				if granted {
					permissionsSet[perm] = true
				}
			}
		}
	}

	permissions := make([]string, 0, len(permissionsSet))
	for perm := range permissionsSet {
		permissions = append(permissions, perm)
	}
	sort.Strings(permissions)

	return &attendee.UserPermissionsDto{
		Groups:      groups,
		Permissions: permissions,
	}, nil
}

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
