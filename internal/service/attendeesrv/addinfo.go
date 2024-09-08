package attendeesrv

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
)

func (s *AttendeeServiceImplData) GetFullAdditionalInfoArea(ctx context.Context, area string) (map[string]string, error) {
	entries, err := database.GetRepository().GetAllAdditionalInfoForArea(ctx, area)
	if err != nil {
		return make(map[string]string), err
	}
	result := make(map[string]string)
	for _, entry := range entries {
		if entry != nil {
			key := fmt.Sprintf("%d", entry.AttendeeId)
			result[key] = entry.JsonValue
		}
	}
	return result, nil
}

func (s *AttendeeServiceImplData) GetAdditionalInfo(ctx context.Context, attendeeId uint, area string) (string, error) {
	existing, err := database.GetRepository().GetAdditionalInfoFor(ctx, attendeeId, area)
	return existing.JsonValue, err
}

func (s *AttendeeServiceImplData) WriteAdditionalInfo(ctx context.Context, attendeeId uint, area string, value string) error {
	existing, err := database.GetRepository().GetAdditionalInfoFor(ctx, attendeeId, area)
	if err != nil {
		return err
	}

	existing.JsonValue = value

	return database.GetRepository().WriteAdditionalInfo(ctx, existing)
}

func (s *AttendeeServiceImplData) CanAccessAdditionalInfoArea(ctx context.Context, area ...string) (bool, error) {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
		return true, nil
	}

	loggedInSubject := ctxvalues.Subject(ctx)
	allowed, err := s.subjectHasAreaPermissionEntry(ctx, loggedInSubject, area...)
	return allowed, err
}

func (s *AttendeeServiceImplData) CanAccessOwnAdditionalInfoArea(ctx context.Context, attendeeId uint, wantWriteAccess bool, area string) (bool, error) {
	att, err := database.GetRepository().GetAttendeeById(ctx, attendeeId)
	if err != nil {
		// attendee does not exist is checked later in order to not expose information
		return false, nil
	}

	loggedInSubject := ctxvalues.Subject(ctx)
	if loggedInSubject != "" && loggedInSubject == att.Identity {
		conf := config.AdditionalInfoConfiguration(area)
		if wantWriteAccess && conf.SelfWrite {
			return true, nil
		}
		if !wantWriteAccess && conf.SelfRead {
			return true, nil
		}
	}

	return false, nil
}

func (s *AttendeeServiceImplData) CanUseFindAttendee(ctx context.Context) (bool, error) {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
		return true, nil
	}

	permissions := config.PermissionsAllowingFindAttendees()
	loggedInSubject := ctxvalues.Subject(ctx)
	allowed, err := s.subjectHasDirectPermissionEntry(ctx, loggedInSubject, permissions...)
	return allowed, err
}
