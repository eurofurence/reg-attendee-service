package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/api/v1/errorapi"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func tstRequireErrorResponse(t *testing.T, response tstWebResponse, expectedStatus int, expectedMessage string, expectedDetails string) {
	require.Equal(t, expectedStatus, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, expectedMessage, errorDto.Message, "unexpected error code")
	if expectedDetails != "" {
		require.EqualValues(t, url.Values{"details": []string{expectedDetails}}, errorDto.Details, "unexpected error details")
	}
}
