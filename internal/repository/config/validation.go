package config

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/web/util/validation"
	"net/url"
	"strings"
)

func setConfigurationDefaults(c *conf) {
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Logging.Severity == "" {
		c.Logging.Severity = "INFO"
	}
	if c.Database.Use == "" {
		c.Database.Use = "inmemory"
	}
}

const portPattern = "^[1-9][0-9]{0,4}$"

func validateServerConfiguration(errs url.Values, c serverConfig) {
	if validation.ViolatesPattern(portPattern, c.Port) {
		errs.Add("server.port", "must be a number between 1 and 65535")
	}
}

var allowedSeverities = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}

func validateLoggingConfiguration(errs url.Values, c loggingConfig) {
	if validation.NotInAllowedValues(allowedSeverities[:], c.Severity) {
		errs.Add("logging.severity", "must be one of DEBUG, INFO, WARN, ERROR")
	}
}

var allowedSecurity = [...]string{"fixed-token"}

func validateSecurityConfiguration(errs url.Values, c securityConfig) {
	if validation.NotInAllowedValues(allowedSecurity[:], c.Use) {
		errs.Add("security.use", "currently must be fixed-token")
	}
	validation.CheckLength(&errs, 16, 256, "security.fixed.admin", c.Fixed.Admin)
	validation.CheckLength(&errs, 16, 256, "security.fixed.user", c.Fixed.User)
	if c.Fixed.InitialReg != "" {
		validation.CheckLength(&errs, 16, 256, "security.fixed.reg", c.Fixed.InitialReg)
	}
}

var allowedDatabases = [...]string{"mysql", "inmemory"}

func validateDatabaseConfiguration(errs url.Values, c databaseConfig) {
	if validation.NotInAllowedValues(allowedDatabases[:], c.Use) {
		errs.Add("database.use", "must be one of mysql, inmemory")
	}
	if c.Use == "mysql" {
		validation.CheckLength(&errs, 1, 256, "database.mysql.username", c.Mysql.Username)
		validation.CheckLength(&errs, 1, 256, "database.mysql.password", c.Mysql.Password)
		validation.CheckLength(&errs, 1, 256, "database.mysql.database", c.Mysql.Database)
	}
}

const keyPattern = "^[a-zA-Z0-9_-]+$"

func validateFlagsConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.flags." + k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.flags." + k + ".description", v.Description)
		validation.CheckLength(&errs, 1, 256, "choices.flags." + k + ".help_url", v.HelpUrl)
		checkConstraints(errs, c, "choices.flags", k, v.Constraint, v.ConstraintMsg)
	}
}

func validatePackagesConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.packages." + k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.packages." + k + ".description", v.Description)
		validation.CheckLength(&errs, 1, 256, "choices.packages." + k + ".help_url", v.HelpUrl)
		checkConstraints(errs, c, "choices.packages", k, v.Constraint, v.ConstraintMsg)
	}
}

func validateOptionsConfiguration(errs url.Values, c map[string]ChoiceConfig) {
	for k, v := range c {
		if validation.ViolatesPattern(keyPattern, k) {
			errs.Add("choices.options." + k, "invalid key, must consist of a-z A-Z 0-9 - _ only")
		}
		validation.CheckLength(&errs, 1, 256, "choices.options." + k + ".description", v.Description)
		validation.CheckLength(&errs, 1, 256, "choices.options." + k + ".help_url", v.HelpUrl)
		checkConstraints(errs, c, "choices.options", k, v.Constraint, v.ConstraintMsg)
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
				errs.Add(keyPrefix + "." + key + ".constraint", "invalid key in constraint, references nonexistent entry")
			}
			if choiceKey == key {
				errs.Add(keyPrefix + "." + key + ".constraint", "invalid self referential constraint")
			}
			validation.CheckLength(&errs, 1, 256, keyPrefix + "." + key + ".constraint_msg", constraintMsg)
		}
	}
}
