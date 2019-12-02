// configuration management using a yaml configuration file
// You must have called LoadConfiguration() or otherwise set up the configuration before you can use these.
package config

import (
	"errors"
	"log"
	"strings"
	"time"
)

func ServerAddr() string {
	c := Configuration();
	return c.Server.Address + ":" + c.Server.Port
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
		log.Printf("[00000000] ERROR invalid argument to config.FixedToken: %v, this is an error in your code! Find it and fix it. Returning invalid token!", forGroup)
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

func AllowedFlags() []string {
	return sortedKeys(Configuration().Choices.Flags)
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

func FlagsConfig() map[string]ChoiceConfig {
	return Configuration().Choices.Flags
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
	t, _ := time.Parse(startTimeFormat, Configuration().GoLive.StartIsoDatetime)
	return t
}
