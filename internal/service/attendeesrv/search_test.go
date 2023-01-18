package attendeesrv

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChecksum(t *testing.T) {
	require.Equal(t, "Y", calculateChecksum(4))
	require.Equal(t, "P", calculateChecksum(38))
	require.Equal(t, "B", calculateChecksum(422))
	require.Equal(t, "F", calculateChecksum(4194))
	require.Equal(t, "X", calculateChecksum(88210))
	require.Equal(t, "W", calculateChecksum(987666))
}
