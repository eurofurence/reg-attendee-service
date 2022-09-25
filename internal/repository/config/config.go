// configuration management using a yaml configuration file
// You must have called LoadConfiguration() or otherwise set up the configuration before you can use these.
package config

import (
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"strings"
	"time"
)

func UseEcsLogging() bool {
	return ecsLogging
}

func ServerAddr() string {
	c := Configuration()
	return fmt.Sprintf("%s:%s", c.Server.Address, c.Server.Port)
}

func ServerReadTimeout() time.Duration {
	return time.Second * time.Duration(Configuration().Server.ReadTimeout)
}

func ServerWriteTimeout() time.Duration {
	return time.Second * time.Duration(Configuration().Server.WriteTimeout)
}

func ServerIdleTimeout() time.Duration {
	return time.Second * time.Duration(Configuration().Server.IdleTimeout)
}

func DatabaseUse() string {
	return Configuration().Database.Use
}

func DatabaseMysqlConnectString() string {
	c := Configuration().Database.Mysql
	return c.Username + ":" + c.Password + "@" +
		c.Database + "?" + strings.Join(c.Parameters, "&")
}

func MigrateDatabase() bool {
	return dbMigrate
}

func LoggingSeverity() string {
	return Configuration().Logging.Severity
}

type FixedTokenEnum int

const (
	TokenForAdmin              FixedTokenEnum = iota
	TokenForLoggedInUser       FixedTokenEnum = iota
	OptionalTokenForInitialReg FixedTokenEnum = iota
)

func FixedToken(forGroup FixedTokenEnum) (string, error) {
	tokens := Configuration().Security.Fixed
	switch forGroup {
	case TokenForAdmin:
		return tokens.Admin, nil
	case TokenForLoggedInUser:
		return tokens.User, nil
	case OptionalTokenForInitialReg:
		return tokens.InitialReg, nil
	default:
		aulogging.Logger.NoCtx().Error().Printf("invalid argument to config.FixedToken: %v, this is an error in your code! Find it and fix it. Returning invalid token!", forGroup)
		return "", errors.New("invalid token group argument")
	}
}

func OptionalInitialRegTokenConfigured() bool {
	return Configuration().Security.Fixed.InitialReg != ""
}

func AllAvailableFixedTokenGroups() []FixedTokenEnum {
	if OptionalInitialRegTokenConfigured() {
		return []FixedTokenEnum{TokenForAdmin, TokenForLoggedInUser, OptionalTokenForInitialReg}
	} else {
		return []FixedTokenEnum{TokenForAdmin, TokenForLoggedInUser}
	}
}

func AllowedFlagsNoAdmin() []string {
	flags := Configuration().Choices.Flags
	result := make([]string, 0)
	for _, k := range sortedKeys(flags) {
		if !flags[k].AdminOnly {
			result = append(result, k)
		}
	}
	return result
}

func AllowedFlagsAdminOnly() []string {
	flags := Configuration().Choices.Flags
	result := make([]string, 0)
	for _, k := range sortedKeys(flags) {
		if flags[k].AdminOnly {
			result = append(result, k)
		}
	}
	return result
}

func AllowedPackages() []string {
	return sortedKeys(Configuration().Choices.Packages)
}

func AllowedOptions() []string {
	return sortedKeys(Configuration().Choices.Options)
}

func AllowedTshirtSizes() []string {
	return Configuration().TShirtSizes
}

func AllowedCountries() []string {
	return Configuration().Countries
}

func AllowedStatusValues() []string {
	return []string{"new", "approved", "partially paid", "paid", "checked in", "cancelled", "deleted"}
}

func DefaultFlags() string {
	return defaultChoiceStr(Configuration().Choices.Flags)
}

func DefaultPackages() string {
	return defaultChoiceStr(Configuration().Choices.Packages)
}

func DefaultOptions() string {
	return defaultChoiceStr(Configuration().Choices.Options)
}

func defaultChoiceStr(choiceConf map[string]ChoiceConfig) string {
	a := sortedKeys(choiceConf)

	b := a[:0]
	for _, x := range a {
		if choiceConf[x].Default {
			b = append(b, x)
		}
	}

	return strings.Join(b, ",")
}

func FlagsConfigNoAdmin() map[string]ChoiceConfig {
	result := make(map[string]ChoiceConfig)
	for k, v := range Configuration().Choices.Flags {
		if !v.AdminOnly {
			result[k] = v
		}
	}
	return result
}

func FlagsConfigAdminOnly() map[string]ChoiceConfig {
	result := make(map[string]ChoiceConfig)
	for k, v := range Configuration().Choices.Flags {
		if v.AdminOnly {
			result[k] = v
		}
	}
	return result
}

func PackagesConfig() map[string]ChoiceConfig {
	return Configuration().Choices.Packages
}

func OptionsConfig() map[string]ChoiceConfig {
	return Configuration().Choices.Options
}

func EarliestBirthday() string {
	return Configuration().Birthday.Earliest
}

func LatestBirthday() string {
	return Configuration().Birthday.Latest
}

func RegistrationStartTime() time.Time {
	t, _ := time.Parse(StartTimeFormat, Configuration().GoLive.StartIsoDatetime)
	return t
}

func IsCorsDisabled() bool {
	return Configuration().Security.DisableCors
}
