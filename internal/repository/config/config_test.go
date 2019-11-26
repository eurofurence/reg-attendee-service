package config

import (
	"github.com/eurofurence/reg-attendee-service/docs"
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

func TestMigrateDatabase(t *testing.T) {
	docs.Description("ensure migrate database flag is returned correctly")
	dbMigrate = true
	require.Equal(t, true, MigrateDatabase(), "unexpected return value")

	dbMigrate = false
	require.Equal(t, false, MigrateDatabase(), "unexpected return value")
}

func TestFixedTokenInvalidGroup(t *testing.T) {
	docs.Description("test the normally unreachable fixed token lookup for an invalid enum value")
	configurationData = &conf{Security: securityConfig{Fixed: fixedTokenConfig{Admin: "admin", User: "user", InitialReg: "reg"}}}

	token, err := FixedToken(-1)
	require.NotNil(t, err)
	require.Equal(t, "", token)
}
