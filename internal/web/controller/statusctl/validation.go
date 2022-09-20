package statusctl

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"net/url"
	"strings"

	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
)

func validate(ctx context.Context, trustedOriginalStatus string, s *status.StatusChangeDto) url.Values {
	errs := url.Values{}

	if validation.NotInAllowedValues(config.AllowedStatusValues(), s.Status) {
		errs.Add("status", "status must be one of "+strings.Join(config.AllowedStatusValues(), ","))
	}
	validation.CheckLength(&errs, 1, 256, "comment", s.Comment)

	if len(errs) != 0 {
		logger := logging.Ctx(ctx)
		if logger.IsDebugEnabled() {
			for key, val := range errs {
				logger.Debugf("status change dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}
