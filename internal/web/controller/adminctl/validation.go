package adminctl

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"net/url"
)

func validate(ctx context.Context, a *admin.AdminInfoDto, trustedOriginalState *entity.AdminInfo) url.Values {
	errs := url.Values{}

	if a.Id != 0 && a.Id != trustedOriginalState.ID {
		errs.Add("id", "id field must be empty or correctly assigned for incoming requests")
	}

	validation.CheckCombinationOfAllowedValues(&errs, []string{"admin", "regdesk", "sponsordesk", "view", "stats", "announce", "export_conbook"}, "permissions", a.Permissions)

	validation.CheckCombinationOfAllowedValues(&errs, config.AllowedFlagsAdminOnly(), "flags", a.Flags)
	if err := attendeeService.CanChangeChoiceTo(ctx, trustedOriginalState.Flags, a.Flags, config.FlagsConfigAdminOnly()); err != nil {
		errs.Add("flags", err.Error())
	}

	if len(errs) != 0 {
		if config.LoggingSeverity() == "DEBUG" {
			logger := aulogging.Logger.Ctx(ctx).Debug()
			for key, val := range errs {
				logger.Printf("adminInfo dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}
