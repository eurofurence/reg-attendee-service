package config

import (
	"crypto/rsa"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/validation"
	"github.com/golang-jwt/jwt/v4"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

func setConfigurationDefaults(c *Application) {
	if c.Service.Name == "" {
		c.Service.Name = "Registration Attendee Service"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.ReadTimeout <= 0 {
		c.Server.ReadTimeout = 5
	}
	if c.Server.WriteTimeout <= 0 {
		c.Server.WriteTimeout = 5
	}
	if c.Server.IdleTimeout <= 0 {
		c.Server.IdleTimeout = 5
	}
	if c.Logging.Severity == "" {
		c.Logging.Severity = "INFO"
	}
	if c.Logging.Style == "" {
		c.Logging.Style = ECS
	}
	if c.Currency == "" {
		c.Currency = "EUR"
	}
	if c.VatPercent == 0 {
		c.VatPercent = 19.0
	}
	if c.Database.Use == "" {
		c.Database.Use = "inmemory"
	}
	if c.Security.Cors.AllowOrigin == "" {
		c.Security.Cors.AllowOrigin = "*"
	}
	for name, pkgConf := range c.Choices.Packages {
		if pkgConf.MaxCount == 0 {
			pkgConf.MaxCount = 1
			c.Choices.Packages[name] = pkgConf
		}
	}
	if len(c.SpokenLanguages) == 0 {
		c.SpokenLanguages = []string{"en-US"}
	}
	if len(c.RegistrationLanguages) == 0 {
		c.RegistrationLanguages = []string{"en-US"}
	}
	if c.Dues.DueDays == 0 {
		c.Dues.DueDays = 14
	}
	if len(c.Security.FindApiAccess.Permissions) == 0 {
		c.Security.FindApiAccess.Permissions = []string{"regdesk", "sponsordesk"}
	}
}

const (
	envDbPassword = "REG_SECRET_DB_PASSWORD"
	envApiToken   = "REG_SECRET_API_TOKEN"
)

func applyEnvVarOverrides(c *Application) {
	if dbPassword := os.Getenv(envDbPassword); dbPassword != "" {
		c.Database.Password = dbPassword
	}
	if apiToken := os.Getenv(envApiToken); apiToken != "" {
		c.Security.Fixed.Api = apiToken
	}
}

const portPattern = "^[1-9][0-9]{0,4}$"

func validateServerConfiguration(errs url.Values, c ServerConfig) {
	if validation.ViolatesPattern(portPattern, c.Port) {
		errs.Add("server.port", "must be a number between 1 and 65535")
	}
	validation.CheckIntValueRange(&errs, 1, 300, "server.read_timeout_seconds", c.ReadTimeout)
	validation.CheckIntValueRange(&errs, 1, 300, "server.write_timeout_seconds", c.WriteTimeout)
	validation.CheckIntValueRange(&errs, 1, 300, "server.idle_timeout_seconds", c.IdleTimeout)
}

var allowedSeverities = []string{"DEBUG", "INFO", "WARN", "ERROR"}

func validateLoggingConfiguration(errs url.Values, c LoggingConfig) {
	if validation.NotInAllowedValues(allowedSeverities[:], c.Severity) {
		errs.Add("logging.severity", "must be one of DEBUG, INFO, WARN, ERROR")
	}
}

func validateSecurityConfiguration(errs url.Values, c SecurityConfig) {
	validation.CheckLength(&errs, 16, 256, "security.fixed.api", c.Fixed.Api)
	validation.CheckLength(&errs, 1, 256, "security.oidc.admin_group", c.Oidc.AdminGroup)

	parsedKeySet = make([]*rsa.PublicKey, 0)
	for i, keyStr := range c.Oidc.TokenPublicKeysPEM {
		publicKeyPtr, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
		if err != nil {
			errs.Add(fmt.Sprintf("security.oidc.token_public_keys_PEM[%d]", i), fmt.Sprintf("failed to parse RSA public key in PEM format: %s", err.Error()))
		} else {
			parsedKeySet = append(parsedKeySet, publicKeyPtr)
		}
	}

	for _, perm := range c.FindApiAccess.Permissions {
		if validation.ViolatesPattern(attendeePermissionPattern, perm) {
			errs.Add("security.find_api_access.permissions", "permissions for find api access must match [a-z]+, no other characters allowed")
		}
	}
}

var allowedDatabases = []DatabaseType{Mysql, Inmemory}

func validateDatabaseConfiguration(errs url.Values, c DatabaseConfig) {
	if validation.NotInAllowedValues(allowedDatabases[:], c.Use) {
		errs.Add("database.use", "must be one of mysql, inmemory")
	}
	if c.Use == Mysql {
		validation.CheckLength(&errs, 1, 256, "database.username", c.Username)
		validation.CheckLength(&errs, 1, 256, "database.password", c.Password)
		validation.CheckLength(&errs, 1, 256, "database.database", c.Database)
	}
}

func validateBirthdayConfiguration(errs url.Values, c BirthdayConfig) {
	if validation.InvalidISODate(c.Earliest) {
		errs.Add("birthday.earliest", "invalid earliest birthday, must be specified as an ISO Date, as in 1901-01-01")
	}
	if validation.InvalidISODate(c.Latest) {
		errs.Add("birthday.latest", "invalid latest birthday, must be specified as an ISO Date, as in 2019-08-24. It is acceptable to specify the last day of the convention, if you wish to allow any underage participants. Otherwise use the first day, 18 years ago.")
	}
}

const keyPattern = "^[a-zA-Z0-9_-]+$"

func validateFlagsConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.flags."+k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.flags."+k+".description", v.Description)
		checkConstraints(errs, c, "choices.flags", k, v.Constraint, v.ConstraintMsg)
		if v.AdminOnly && v.ReadOnly {
			errs.Add("choices.flags."+k+".admin", "a flag cannot both be admin_only and read_only")
		}
		if v.AdminOnly && v.Default {
			errs.Add("choices.flags."+k+".default", "a flag cannot both be admin_only and default to on")
		}
	}
}

func validatePackagesConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.packages."+k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.packages."+k+".description", v.Description)
		checkConstraints(errs, c, "choices.packages", k, v.Constraint, v.ConstraintMsg)
		if v.AdminOnly {
			errs.Add("choices.packages."+k+".admin", "packages cannot be admin_only (they cost money). Try read_only instead.")
		}
		if len(v.AllowedCounts) > 0 {
			if v.MaxCount <= 1 {
				errs.Add("choices.packages."+k+".allowed_counts", "can only list allowed counts if max_count is set to at least 2")
			} else if slices.Max(v.AllowedCounts) > v.MaxCount {
				errs.Add("choices.packages."+k+".allowed_counts", "maximum allowed_counts value cannot be larger than max_count for package")
			}
		}
	}
}

func validateOptionsConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.options."+k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.options."+k+".description", v.Description)
		checkConstraints(errs, c, "choices.options", k, v.Constraint, v.ConstraintMsg)
		if v.AdminOnly {
			errs.Add("choices.options."+k+".admin", "options cannot be admin_only (they represent user choices).")
		}
		if v.ReadOnly {
			errs.Add("choices.options."+k+".readonly", "options cannot be read_only (they represent user choices).")
		}
	}
}

func checkConstraints(errs url.Values, c map[string]ChoiceConfig, keyPrefix string, key string, constraint string, constraintMsg string) {
	if constraint != "" {
		constraints := strings.Split(constraint, ",")
		for _, cn := range constraints {
			choiceKey := cn
			if strings.HasPrefix(cn, "!") {
				choiceKey = strings.TrimPrefix(cn, "!")
			}
			if _, ok := c[choiceKey]; !ok {
				errs.Add(keyPrefix+"."+key+".constraint", "invalid key in constraint, references nonexistent entry")
			} else {
				if c[choiceKey].AdminOnly != c[key].AdminOnly {
					errs.Add(keyPrefix+"."+key+".constraint", "invalid key in constraint, references across admin only and non-admin only")
				}
			}
			if choiceKey == key {
				errs.Add(keyPrefix+"."+key+".constraint", "invalid self referential constraint")
			}
			validation.CheckLength(&errs, 1, 256, keyPrefix+"."+key+".constraint_msg", constraintMsg)
		}
	}
}

func validateRegistrationStartTime(errs url.Values, c GoLiveConfig, s SecurityConfig) {
	normal, err := time.Parse(StartTimeFormat, c.StartIsoDatetime)
	if err != nil {
		errs.Add("go_live.start_iso_datetime", "invalid date/time format, use ISO with numeric timezone as in "+StartTimeFormat)
	}

	if c.EarlyRegStartIsoDatetime != "" {
		early, err := time.Parse(StartTimeFormat, c.EarlyRegStartIsoDatetime)
		if err != nil {
			errs.Add("go_live.early_reg_start_iso_datetime", "invalid date/time format, use ISO with numeric timezone as in "+StartTimeFormat)
		}

		if normal.Before(early) {
			errs.Add("go_live.early_reg_start_iso_datetime", "if supplied, must be earlier than go_live.start_iso_datetime")
		}

		if s.Oidc.EarlyRegGroup == "" {
			errs.Add("go_live.early_reg_start_iso_datetime", "if supplied, must also supply security.oidc.early_reg_group so early registration is possible")
		}
	}
}

func validateDuesConfiguration(errs url.Values, c DuesConfig) {
	earliest, err := time.Parse(IsoDateFormat, c.EarliestDueDate)
	if err != nil {
		errs.Add("dues.earliest_due_date", "invalid date format, use ISO date as in "+IsoDateFormat)
	}

	latest, err := time.Parse(IsoDateFormat, c.LatestDueDate)
	if err != nil {
		errs.Add("dues.latest_due_date", "invalid date format, use ISO date as in "+IsoDateFormat)
	}

	if c.DueDays <= 0 {
		errs.Add("dues.due_days", "must be positive integer")
	}

	if latest.Before(earliest) {
		errs.Add("dues.latest_due_date", "must be no earlier than dues.earliest_due_date")
	}
}

const publicUrlPattern = "^https?://"
const downstreamPattern = "^(|https?://.*[^/])$"

func validateServiceConfiguration(errs url.Values, c ServiceConfig) {
	if validation.ViolatesPattern(publicUrlPattern, c.RegsysPublicUrl) {
		errs.Add("service.regsys_public_url", "public url must start with http:// or https://, this should be the outward facing main page of the regsys, used in emails")
	}
	if validation.ViolatesPattern(downstreamPattern, c.PaymentService) {
		errs.Add("service.payment_service", "base url must be empty (enables in-memory simulator) or start with http:// or https:// and may not end in a /")
	}
	if validation.ViolatesPattern(downstreamPattern, c.MailService) {
		errs.Add("service.mail_service", "base url must be empty (enables in-memory simulator) or start with http:// or https:// and may not end in a /")
	}
}

const addInfoAreaPattern = "^[a-z]+$"
const attendeePermissionPattern = "^[a-z]+$"

func validateAdditionalInfoConfiguration(errs url.Values, areas map[string]AddInfoConfig) {
	for area, config := range areas {
		if validation.ViolatesPattern(addInfoAreaPattern, area) {
			errs.Add("additional_info_areas", "keys for additional info fields must match [a-z]+, no other characters allowed")
		}
		if area == "overdue" {
			errs.Add("additional_info_areas.overdue", "this key is reserved for internal use by the admin frontend, you may not configure it")
		}
		for _, perm := range config.Permissions {
			if validation.ViolatesPattern(attendeePermissionPattern, perm) {
				errs.Add(fmt.Sprintf("additional_info_areas.%s.permissions", perm), "permissions for additional info access must match [a-z]+, no other characters allowed")
			}
		}
	}
}
