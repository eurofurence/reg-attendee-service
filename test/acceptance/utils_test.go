package acceptance

import (
	"encoding/json"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

// placing these here because they are package global

type tstWebResponse struct {
	status      int
	body        string
	contentType string
	location    string
}

func tstWebResponseFromResponse(response *http.Response) tstWebResponse {
	status := response.StatusCode
	ct := ""
	if val, ok := response.Header[headers.ContentType]; ok {
		ct = val[0]
	}
	loc := ""
	if val, ok := response.Header[headers.Location]; ok {
		loc = val[0]
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponse{
		status:      status,
		body:        string(body),
		contentType: ct,
		location:    loc,
	}
}

func tstPerformGet(relativeUrlWithLeadingSlash string, bearerToken string) tstWebResponse {
	request, err := http.NewRequest(http.MethodGet, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		log.Fatal(err)
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPut(relativeUrlWithLeadingSlash string, requestBody string, bearerToken string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPut, ts.URL+relativeUrlWithLeadingSlash, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPost(relativeUrlWithLeadingSlash string, requestBody string, bearerToken string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPost, ts.URL+relativeUrlWithLeadingSlash, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstBuildValidAttendee(testcase string) attendee.AttendeeDto {
	timer := time.Now().UnixNano()
	return attendee.AttendeeDto{
		Nickname:     "BlackCheetah",
		FirstName:    "Hans",
		LastName:     "Mustermann",
		Street:       "Teststraße 24",
		Zip:          "12345",
		City:         "Berlin",
		Country:      "DE",
		CountryBadge: "DE",
		State:        "Sachsen",
		Email:        testcase + fmt.Sprint(timer) + "-jsquirrel_github_9a6d@packetloss.de",
		Phone:        "+49-30-123",
		Telegram:     "@ihopethisuserdoesnotexist",
		Birthday:     "1998-11-23",
		Gender:       "other",
		Pronouns:     "he/him",
		Flags:        "anon,hc",
		Packages:     "room-none,attendance,stage,sponsor2",
		Options:      "music,suit",
		TshirtSize:   "XXL",
	}
}

func tstBuildValidBanRule(testcase string) bans.BanRule {
	return bans.BanRule{
		Reason:          testcase,
		NamePattern:     "^name.*" + testcase,
		NicknamePattern: "^nickname.*" + testcase,
		EmailPattern:    "^email.*" + testcase,
	}
}

func tstRegisterAttendee(t *testing.T, testcase string) (location string, dtoWithId attendee.AttendeeDto) {
	return tstRegisterAttendeeWithToken(t, testcase, tstValidStaffToken(t, "1"))
}

func tstRegisterAttendeeWithToken(t *testing.T, testcase string, token string) (location string, dtoWithId attendee.AttendeeDto) {
	dto := tstBuildValidAttendee(testcase)
	creationResponse := tstPerformPost("/api/rest/v1/attendees", tstRenderJson(dto), token)
	require.Equal(t, http.StatusCreated, creationResponse.status, "unexpected http response status")

	rereadResponse := tstPerformGet(creationResponse.location, token)
	require.Equal(t, http.StatusOK, rereadResponse.status, "unexpected http response status")
	tstParseJson(rereadResponse.body, &dtoWithId)

	return creationResponse.location, dtoWithId
}

func tstRenderJson(v interface{}) string {
	representationBytes, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return string(representationBytes)
}

// tip: dto := &attendee.AttendeeDto{}
func tstParseJson(body string, dto interface{}) {
	err := json.Unmarshal([]byte(body), dto)
	if err != nil {
		log.Fatal(err)
	}
}

func tstValidAttendeeDues(amount int64, comment string) paymentservice.Transaction {
	return paymentservice.Transaction{
		ID:        "",
		DebitorID: 1,
		Type:      paymentservice.Due,
		Method:    paymentservice.Internal,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: amount,
			VatRate:   19.0,
		},
		Comment:       comment,
		Status:        paymentservice.Valid,
		EffectiveDate: "",          // TODO
		DueDate:       time.Time{}, // TODO
		Deletion:      nil,         // TODO
	}
}

func tstNewStatusMail(testcase string, newStatus string) mailservice.TemplateRequestDto {
	return mailservice.TemplateRequestDto{
		Name: "new-status-" + newStatus,
		Variables: map[string]string{
			"nickname": "BlackCheetah",
		},
		Email: testcase,
	}
}
