package banctl

import (
	"context"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"net/url"
)

func validate(ctx context.Context, b *bans.BanRule, allowedId uint) url.Values {
	errs := url.Values{}

	if b.Id != "" && b.Id != fmt.Sprint(allowedId) {
		errs.Add("id", "id field must be empty or correctly assigned for incoming requests")
	}

	validation.CheckLength(&errs, 1, 255, "reason", b.Reason)
	validation.CheckValidRegexOrEmpty(&errs, "name_pattern", b.NamePattern)
	validation.CheckValidRegexOrEmpty(&errs, "nickname_pattern", b.NicknamePattern)
	validation.CheckValidRegexOrEmpty(&errs, "email_pattern", b.EmailPattern)

	if len(errs) != 0 {
		if config.LoggingSeverity() == "DEBUG" {
			logger := aulogging.Logger.Ctx(ctx).Debug()
			for key, val := range errs {
				logger.Printf("ban rule validation error for key %s: %s", key, val)
			}
		}
	}
	return errs
}
