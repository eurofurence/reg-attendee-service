package attendeesrv

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
)

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
	allowed, err := s.subjectHasAdminPermissionEntry(ctx, loggedInSubject, area...)
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
