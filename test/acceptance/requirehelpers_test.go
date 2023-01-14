package acceptance

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/errorapi"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func tstRequireErrorResponse(t *testing.T, response tstWebResponse, expectedStatus int, expectedMessage string, expectedDetails interface{}) {
	require.Equal(t, expectedStatus, response.status, "unexpected http response status")
	errorDto := errorapi.ErrorDto{}
	tstParseJson(response.body, &errorDto)
	require.Equal(t, expectedMessage, errorDto.Message, "unexpected error code")
	expectedDetailsStr, ok := expectedDetails.(string)
	if ok && expectedDetailsStr != "" {
		require.EqualValues(t, url.Values{"details": []string{expectedDetailsStr}}, errorDto.Details, "unexpected error details")
	}
	expectedDetailsUrlValues, ok := expectedDetails.(url.Values)
	if ok {
		require.EqualValues(t, expectedDetailsUrlValues, errorDto.Details, "unexpected error details")
	}
}

func tstRequireTransactions(t *testing.T, expectedTransactions []paymentservice.Transaction) {
	require.Equal(t, len(expectedTransactions), len(paymentMock.Recording()))
	for i, expected := range expectedTransactions {
		actual := paymentMock.Recording()[i]
		require.EqualValues(t, expected, actual)
	}
}

func tstRequireMailRequests(t *testing.T, expectedMailRequests []mailservice.MailSendDto) {
	require.Equal(t, len(expectedMailRequests), len(mailMock.Recording()))
	for i, expected := range expectedMailRequests {
		actual := mailMock.Recording()[i]
		require.Equal(t, len(expected.To), len(actual.To))
		for i := range expected.To {
			require.Contains(t, actual.To[i], expected.To[i])
		}
		actual.To = expected.To
		require.Equal(t, len(expected.Variables), len(actual.Variables))
		require.EqualValues(t, expected, actual)
	}
}
