package attendeesrv

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestRandomString(t *testing.T) {
	normalsReplacer := regexp.MustCompile(`[A-Za-z0-9]`)
	for n := 0; n < 1000; n++ {
		str := randomString(3, 80, 2)
		failMsg := fmt.Sprintf("failed for %s", str)
		require.True(t, len([]rune(str)) >= 3, failMsg)
		require.True(t, len([]rune(str)) <= 80, failMsg)
		require.True(t, len([]rune(normalsReplacer.ReplaceAllString(str, ""))) <= 2, failMsg)
	}
}
