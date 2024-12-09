// configuration management using a yaml configuration file
// You must have called LoadConfiguration() or otherwise set up the configuration before you can use these.
package config

import (
	"crypto/rsa"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"sort"
	"strings"
	"time"
)

func UseEcsLogging() bool {
	return ecsLogging
}

func GeneratorGenerateCount() uint {
	return generateCount
}

func GeneratorParallel() uint {
	return parallel
}

func GeneratorBaseUrl() string {
	return baseUrl
}

func GeneratorCookieDomain() string {
	return cookieDomain
}

func GeneratorIdToken() string {
	return idToken
}

func GeneratorAccessToken() string {
	return accessToken
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

func OidcIdTokenCookieName() string {
	return Configuration().Security.Oidc.IdTokenCookieName
}

func OidcAccessTokenCookieName() string {
	return Configuration().Security.Oidc.AccessTokenCookieName
}

func OidcKeySet() []*rsa.PublicKey {
	return parsedKeySet
}

func OidcAdminGroup() string {
	return Configuration().Security.Oidc.AdminGroup
}

func OidcEarlyRegGroup() string {
	return Configuration().Security.Oidc.EarlyRegGroup
}

func OidcAllowedAudience() string {
	return Configuration().Security.Oidc.Audience
}

func OidcAllowedIssuer() string {
	return Configuration().Security.Oidc.Issuer
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

func AdditionalInfoFieldNames() []string {
	result := make([]string, 0)
	for k := range Configuration().AdditionalInfo {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// AllowedPermissions returns a sorted unique list of all permissions referenced in
// additional info configurations or configured to have access to the find api.
func AllowedPermissions() []string {
	resultMap := make(map[string]bool)
	for _, v := range Configuration().AdditionalInfo {
		for _, perm := range v.Permissions {
			resultMap[perm] = true
		}
	}

	for _, perm := range Configuration().Security.FindApiAccess.Permissions {
		resultMap[perm] = true
	}

	result := make([]string, 0)
	for k := range resultMap {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func AdditionalInfoConfiguration(fieldName string) AddInfoConfig {
	v, ok := Configuration().AdditionalInfo[fieldName]
	if ok {
		return v
	} else {
		return AddInfoConfig{
			SelfRead:    false,
			SelfWrite:   false,
			Permissions: []string{},
		}
	}
}

func PermissionsAllowingFindAttendees() []string {
	return Configuration().Security.FindApiAccess.Permissions
}

func AllowedTshirtSizes() []string {
	return Configuration().TShirtSizes
}

func AllowedCountries() []string {
	return Configuration().Countries
}

func AllowedSpokenLanguages() []string {
	return Configuration().SpokenLanguages
}

func AllowedRegistrationLanguages() []string {
	return Configuration().RegistrationLanguages
}

func DefaultRegistrationLanguage() string {
	// default set after load ensures always at least one entry
	return Configuration().RegistrationLanguages[0]
}

func AllowedStatusValues() []status.Status {
	return []status.Status{status.New, status.Approved, status.PartiallyPaid, status.Paid, status.CheckedIn, status.Waiting, status.Cancelled, status.Deleted}
}

func DefaultFlags() string {
	return defaultChoiceStr(Configuration().Choices.Flags)
}

func DefaultPackages() string {
	// this is ok because all our parsing implementations can deal with a simple comma separated list
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

func AnonymizeIdentity() bool {
	return Configuration().Security.AnonymizeIdentity
}

func PaymentServiceBaseUrl() string {
	return Configuration().Service.PaymentService
}

func MailServiceBaseUrl() string {
	return Configuration().Service.MailService
}

func AuthServiceBaseUrl() string {
	return Configuration().Service.AuthService
}

func DueDays() time.Duration {
	return time.Duration(Configuration().Dues.DueDays*24) * time.Hour
}

func EarliestDueDate() string {
	return Configuration().Dues.EarliestDueDate
}

func LatestDueDate() string {
	return Configuration().Dues.LatestDueDate
}

func Currency() string {
	return Configuration().Currency
}

func VatPercent() float64 {
	return Configuration().VatPercent
}

func RegsysPublicUrl() string {
	return Configuration().Service.RegsysPublicUrl
}
