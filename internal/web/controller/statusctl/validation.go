package statusctl

import (
	"context"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"net/url"
	"strings"

	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

func validate(ctx context.Context, trustedOriginalStatus status.Status, s *status.StatusChangeDto) url.Values {
	errs := url.Values{}

	if validation.NotInAllowedValues(config.AllowedStatusValues(), s.Status) {
		errs.Add("status", "status must be one of "+strings.Join(convertStatusValuesSlice(config.AllowedStatusValues()), ","))
	}
	validation.CheckLength(&errs, 1, 256, "comment", s.Comment)

	if len(errs) != 0 {
		if config.LoggingSeverity() == "DEBUG" {
			logger := aulogging.Logger.Ctx(ctx).Debug()
			for key, val := range errs {
				logger.Printf("status change dto validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}

func convertStatusValuesSlice(input []status.Status) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = string(v)
	}
	return result
}
