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

func AllowedFlags() []string {
	return sortedKeys(&Configuration().Choices.Flags)
}

func AllowedPackages() []string {
	return sortedKeys(&Configuration().Choices.Packages)
}

func AllowedOptions() []string {
	return sortedKeys(&Configuration().Choices.Options)
}

// TODO implement actually reading the data

func AllowedTshirtSizes() []string {
	return []string{"XS", "S", "M", "L", "XL", "XXL", "XXXL", "XXXXL"}
}