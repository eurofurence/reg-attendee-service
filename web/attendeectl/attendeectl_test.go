package attendeectl

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"rexis/rexis-go-attendee/api/v1/attendee"
	"rexis/rexis-go-attendee/docs"
	"strings"
	"testing"
)

// see setup and mock for attendee service in config_test.go

func TestParseBodyToAttendeeDtoParseError(t *testing.T) {
	docs.Description("a parse failure of the json body should lead to the correct 'bad request' error response")
	w := httptest.NewRecorder()
	r := tstMockPostRequest("{[[garbage}")

	_, err := parseBodyToAttendeeDto(w, r)
	require.NotNil(t, err, "expected an error return code so controller will bail out")
	tstRequireErrorResponse(t, w, http.StatusBadRequest, "attendee.parse.error")
}

func TestGetAttendeeHandlerInvalidIdUnset(t *testing.T) {
	docs.Description("a parse failure of the id should lead to the correct 'bad request' error response")
	w := httptest.NewRecorder()
	r := tstMockGetRequest("")

	getAttendeeHandler(w, r)
	tstRequireErrorResponse(t, w, http.StatusBadRequest, "attendee.id.invalid")
}

func TestNewAttendeeHandlerWriteError(t *testing.T) {
	docs.Description("a write error should lead to the correct 'internal server error' error response")
	w := httptest.NewRecorder()
	r := tstMockPostRequest(tstRenderJson(tstCreateValidAttendee()))

	newAttendeeHandler(w, r)
	tstRequireErrorResponse(t, w, http.StatusInternalServerError, "attendee.write.error")
}

// helper functions

func tstMockPostRequest(body string) *http.Request {
	r, err := http.NewRequest(http.MethodPost, "/unused/url", strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func tstMockGetRequest(urlParamExtension string) *http.Request {
	r, err := http.NewRequest(http.MethodGet, "/unused/url?" + urlParamExtension, nil)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func tstReadBodyFromResponse(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

func tstParseJson(body string, dto interface{}) {
	err := json.Unmarshal([]byte(body), dto)
	if err != nil {
		log.Fatal(err)
	}
}

func tstRenderJson(dto interface{}) string {
	representationBytes, err := json.Marshal(dto)
	if err != nil {
		log.Fatal(err)
	}
	return string(representationBytes)
}

func tstRequireErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMsg string) {
	response := w.Result()
	require.Equal(t, expectedStatus, response.StatusCode, "unexpected response status")
	actualResponseDto := &attendee.ErrorDto{}
	tstParseJson(tstReadBodyFromResponse(response), actualResponseDto)
	require.Equal(t, expectedMsg, actualResponseDto.Message, "unexpected response contents")
}
