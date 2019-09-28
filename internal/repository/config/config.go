// configuration management using a yaml configuration file
// You must have called LoadConfiguration() or otherwise set up the configuration before you can use these.
package config

import "strings"

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

func FixedToken() string {
	return Configuration().Security.Fixed.Token
}

func AllowedFlags() []string {
	return sortedKeys(&Configuration().Choices.Flags)
}

func AllowedPackages() []string {
	return sortedKeys(&Configuration().Choices.Packages)
}

func AllowedOptions() []string {
	return sortedKeys(&Configuration().Choices.Options)
}

func AllowedTshirtSizes() []string {
	return Configuration().TShirtSizes
}
