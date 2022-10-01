package config

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
		Port: "1234",
	}}
	require.Equal(t, ":1234", ServerAddr(), "unexpected server address string")
}

func TestServerTimeouts(t *testing.T) {
	docs.Description("ensure ServerRead/Write/IdleTimout() return the correct timeouts")
	configurationData = &conf{Logging: loggingConfig{Severity: "DEBUG"}, Server: serverConfig{
		Address:      "localhost",
		Port:         "1234",
		ReadTimeout:  13,
		WriteTimeout: 17,
		IdleTimeout:  23,
	}}
	require.Equal(t, 13*time.Second, ServerReadTimeout())
	require.Equal(t, 17*time.Second, ServerWriteTimeout())
	require.Equal(t, 23*time.Second, ServerIdleTimeout())
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
