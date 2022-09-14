package adminctl

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"net/url"
)

func validate(ctx context.Context, a *admin.AdminInfoDto, trustedOriginalState *entity.AdminInfo) url.Values {
	errs := url.Values{}

	// TODO
	// - ensure only admin only admin only, admin only visible flags come in
	// - ensure only valid permissions come in, values hardcoded probably ok

	if len(errs) != 0 {
		logger := logging.Ctx(ctx)
		if logger.IsDebugEnabled() {
			for key, val := range errs {
				logger.Debugf("adminInfo dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}
