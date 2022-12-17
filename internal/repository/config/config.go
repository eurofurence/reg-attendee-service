// configuration management using a yaml configuration file
// You must have called LoadConfiguration() or otherwise set up the configuration before you can use these.
package config

import (
	"crypto/rsa"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
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

func DatabaseUse() DatabaseType {
	return Configuration().Database.Use
}

func DatabaseMysqlConnectString() string {
	c := Configuration().Database
	return c.Username + ":" + c.Password + "@" +
		c.Database + "?" + strings.Join(c.Parameters, "&")
}

func MigrateDatabase() bool {
	return dbMigrate
}

func LoggingSeverity() string {
	return Configuration().Logging.Severity
}

func FixedApiToken() string {
	return Configuration().Security.Fixed.Api
}

func OidcTokenCookieName() string {
	return Configuration().Security.Oidc.TokenCookieName
}

func OidcKeySet() []*rsa.PublicKey {
	return parsedKeySet
}

func OidcAdminRole() string {
	return Configuration().Security.Oidc.AdminRole
}

func OidcEarlyRegRole() string {
	return Configuration().Security.Oidc.EarlyReg
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

func AllowedSpokenLanguages() []string {
	// TODO add to configuration
	return []string{"de_DE", "en_US"}
}

func AllowedRegistrationLanguages() []string {
	// TODO add to configuration
	return []string{"en_US"}
}

func DefaultRegistrationLanguage() string {
	// TODO add to configuration
	return "en_US"
}

func AllowedStatusValues() []status.Status {
	return []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Waiting, status.Cancelled, status.Deleted}
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

func EarlyRegistrationStartTime() time.Time {
	early := Configuration().GoLive.EarlyRegStartIsoDatetime
	if early != "" {
		t, _ := time.Parse(StartTimeFormat, Configuration().GoLive.EarlyRegStartIsoDatetime)
		return t
	} else {
		return RegistrationStartTime() // same as normal
	}
}

func IsCorsDisabled() bool {
	return Configuration().Security.Cors.DisableCors
}

func CorsAllowOrigin() string {
	return Configuration().Security.Cors.AllowOrigin
}

func RequireLoginForReg() bool {
	return Configuration().Security.RequireLogin
}

func PaymentServiceBaseUrl() string {
	return Configuration().Service.PaymentService
}

func MailServiceBaseUrl() string {
	return Configuration().Service.MailService
}
