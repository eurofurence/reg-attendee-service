package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func tstNoToken() string {
	return ""
}

func tstValidUserToken(t *testing.T, id string) string {
	// TODO - actually distinguish user ids
	token, err := config.FixedToken(config.TokenForLoggedInUser)
	require.Nil(t, err)
	return token
}

func tstValidAdminToken(t *testing.T) string {
	token, err := config.FixedToken(config.TokenForAdmin)
	require.Nil(t, err)
	return token
}

func tstValidStaffToken(t *testing.T, id string) string {
	// TODO - actually distinguish user ids
	token, err := config.FixedToken(config.OptionalTokenForInitialReg)
	require.Nil(t, err)
	require.NotEqual(t, "", token)
	return token
}

func tstValidStaffOrEmptyToken(t *testing.T) string {
	// TODO - this will need to be changed when full security is available
	token, err := config.FixedToken(config.OptionalTokenForInitialReg)
	require.Nil(t, err)
	return token // may be "" if not in staff reg config
}
