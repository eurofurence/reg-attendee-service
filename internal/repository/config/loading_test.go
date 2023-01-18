package config

import (
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/eurofurence/reg-attendee-service/internal/repository/system"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseAndOverwriteConfigInvalidYamlSyntax(t *testing.T) {
	docs.Description("check that a yaml with a syntax error leads to a parse error")
	invalidYaml := `# invalid yaml due to indentation error
choices:
  flags:
      hc:
    anon:
`
	err := parseAndOverwriteConfig([]byte(invalidYaml))
	require.NotNil(t, err, "expected an error")
}

func TestParseAndOverwriteConfigUnexpectedFields(t *testing.T) {
	docs.Description("check that a yaml with unexpected fields leads to a parse error")
	invalidYaml := `# yaml with model mismatches
serval:
  port: 8088
cheetah:
  speed: '60 mph'
`
	err := parseAndOverwriteConfig([]byte(invalidYaml))
	require.NotNil(t, err, "expected an error")
}

func TestStartupLoadConfigurationNoFilename(t *testing.T) {
	docs.Description("check that exit occurs when no configuration filename set")
	system.TestingExitCounter = 0
	system.TestingMode = true
	LoadTestingConfigurationFromPathOrAbort("")
	require.Equal(t, 1, system.TestingExitCounter, "should have called system.Exit()")
}

func TestStartupLoadConfigurationFileNotFound(t *testing.T) {
	docs.Description("check that exit occurs when the configuration file cannot be found")
	system.TestingExitCounter = 0
	system.TestingMode = true
	LoadTestingConfigurationFromPathOrAbort("does-not-exist.yaml")
	require.Equal(t, 1, system.TestingExitCounter, "should have called system.Exit()")
}

func TestParseAndOverwriteConfigValidationErrors1(t *testing.T) {
	docs.Description("check that a yaml with validation errors leads to an error")
	wrongConfigYaml := `# yaml with validation errors
server:
  port: abcde
logging:
  severity: FELINE
database:
  use: the-oracle-of-delphi
choices:
  flags:
    'funny flag $ name':
  packages:
    'funny package name':
  options:
    'funny option name':
`
	err := parseAndOverwriteConfig([]byte(wrongConfigYaml))
	require.NotNil(t, err, "expected an error")
	require.Equal(t, err.Error(), "configuration validation error", "unexpected error message")
}

func TestParseAndOverwriteConfigValidationErrors2(t *testing.T) {
	docs.Description("check that a yaml with validation errors leads to an error")
	wrongConfigYaml := `# yaml with validation errors
server:
  port: abcde
logging:
  severity: FELINE
database:
  use: mysql
`
	err := parseAndOverwriteConfig([]byte(wrongConfigYaml))
	require.NotNil(t, err, "expected an error")
	require.Equal(t, err.Error(), "configuration validation error", "unexpected error message")
}

func TestParseAndOverwriteDefaults(t *testing.T) {
	docs.Description("check that a minimal yaml leads to all defaults being set")
	minimalYaml := `# yaml with minimal settings
service:
  regsys_public_url: 'http://localhost:10000/register'
security:
  fixed_token:
    api: 'fixed-testing-token-abc'
  oidc:
    admin_role: 'admin'
birthday:
  earliest: '1851-01-01'
  latest: '2048-01-01'
go_live:
  start_iso_datetime: '2019-11-28T20:00:00+01:00'
dues:
  earliest_due_date: '2023-01-01'
  latest_due_date: '2023-08-23' # usually last day of convention
`
	err := parseAndOverwriteConfig([]byte(minimalYaml))
	require.Nil(t, err, "expected no error")
	require.Equal(t, "8080", Configuration().Server.Port, "unexpected value for server.port")
	require.Equal(t, "INFO", Configuration().Logging.Severity, "unexpected value for logging.severity")
	require.Equal(t, Inmemory, Configuration().Database.Use, "unexpected value for database.use")
}
