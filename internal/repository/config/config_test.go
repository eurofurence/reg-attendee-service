package config

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/docs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServerAddrWithAddressAndPort(t *testing.T) {
	docs.Description("ensure ServerAddr() returns the correct server address string, with address specified")
	configurationData = &conf{Logging: loggingConfig{Severity: "DEBUG"}, Server: serverConfig{
		Address: "localhost",
		Port:    "1234",
	}}
	require.Equal(t, "localhost:1234", ServerAddr(), "unexpected server address string")
}

func TestServerAddrWithOnlyPort(t *testing.T) {
	docs.Description("ensure ServerAddr() returns the correct server address string, with no address specified")
	configurationData = &conf{Logging: loggingConfig{Severity: "DEBUG"}, Server: serverConfig{
		Port:    "1234",
	}}
	require.Equal(t, ":1234", ServerAddr(), "unexpected server address string")
}

func TestDatabaseMysqlConnectString(t *testing.T) {
	docs.Description("ensure DatabaseMysqlConnectString() returns the correct mysql connect string")
	configurationData = &conf{
		Logging: loggingConfig{Severity: "DEBUG"},
		Database: databaseConfig{
			Use: "mysql",
			Mysql: mysqlConfig{
				Username: "demouser",
				Password: "demopw",
				Database: "tcp(localhost:3306)/dbname",
				Parameters: []string{
					"charset=utf8mb4",
					"timeout=30s"},
			},
		}}
	require.Equal(t, "demouser:demopw@tcp(localhost:3306)/dbname?charset=utf8mb4&timeout=30s", DatabaseMysqlConnectString(), "unexpected mysql db connection string")
}
