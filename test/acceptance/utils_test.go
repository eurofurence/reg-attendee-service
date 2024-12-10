package acceptance

import (
	"encoding/json"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/bans"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
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

func tstAddAuth(request *http.Request, token string) {
	if token == tstValidApiToken() || token == tstInvalidApiToken() {
		request.Header.Set(media.HeaderXApiKey, token)
	} else if token == valid_JWT_is_staff_sub202 {
		// small trick: we derive the access token from the JWT token (for tests only)
		request.Header.Set(headers.Authorization, "Bearer access"+token)
	} else if strings.HasPrefix(token, "access") {
		request.Header.Set(headers.Authorization, "Bearer "+token)
	} else if token != "" {
		request.AddCookie(&http.Cookie{
			Name:     "JWT",
			Value:    token,
			Domain:   "localhost",
			Expires:  time.Now().Add(10 * time.Minute),
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		request.AddCookie(&http.Cookie{
			Name:     "AUTH",
			Value:    "access" + token,
			Domain:   "localhost",
			Expires:  time.Now().Add(10 * time.Minute),
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		if token == valid_JWT_is_admin_sub1234567890 {
			request.Header.Set("X-Admin-Request", "available")
		}
	}
}

func tstPerformGet(relativeUrlWithLeadingSlash string, token string) tstWebResponse {
	request, err := http.NewRequest(http.MethodGet, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		log.Fatal(err)
	}
	tstAddAuth(request, token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPut(relativeUrlWithLeadingSlash string, requestBody string, token string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPut, ts.URL+relativeUrlWithLeadingSlash, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	tstAddAuth(request, token)
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPost(relativeUrlWithLeadingSlash string, requestBody string, token string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPost, ts.URL+relativeUrlWithLeadingSlash, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	tstAddAuth(request, token)
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPostNoBody(relativeUrlWithLeadingSlash string, token string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPost, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		log.Fatal(err)
	}
	tstAddAuth(request, token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformDelete(relativeUrlWithLeadingSlash string, token string) tstWebResponse {
	request, err := http.NewRequest(http.MethodDelete, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		log.Fatal(err)
	}
	tstAddAuth(request, token)
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstBuildValidAttendee(testcase string) attendee.AttendeeDto {
	packages := "attendance,room-none,sponsor2,stage"
	return attendee.AttendeeDto{
		Nickname:             "BlackCheetah",
		FirstName:            "Hans",
		LastName:             "Mustermann",
		Street:               "Teststraße 24",
		Zip:                  testcase + "12345",
		City:                 "Berlin",
		Country:              "DE",
		State:                "Sachsen",
		Email:                "jsquirrel_github_9a6d@packetloss.de",
		Phone:                "+49-30-123",
		Telegram:             "@ihopethisuserdoesnotexist",
		Birthday:             "1998-11-23",
		Gender:               "other",
		Pronouns:             "he/him",
		SpokenLanguages:      "de,en",
		RegistrationLanguage: "en-US",
		Flags:                "anon,hc,terms-accepted",
		Packages:             packages, // ignored because PackagesList is set and takes precedence
		PackagesList:         tstPackagesListFromPackages(packages),
		Options:              "music,suit",
		TshirtSize:           "XXL",
	}
}

func tstOverridePackages(att *attendee.AttendeeDto, packages string) {
	att.Packages = packages
	att.PackagesList = tstPackagesListFromPackages(packages)
}

func tstAddPackages(att *attendee.AttendeeDto, additionalCommaSeparated string) {
	att.Packages = att.Packages + "," + additionalCommaSeparated
	att.PackagesList = tstPackagesListFromPackages(att.Packages)
	// now ensure packages is still sorted
	att.Packages = tstPackagesFromPackagesList(att.PackagesList)
}

func tstPackagesListFromPackages(commaSeparated string) []attendee.PackageState {
	counts := make(map[string]int)
	names := strings.Split(commaSeparated, ",")
	for _, name := range names {
		if name != "" {
			oldCount, _ := counts[name]
			counts[name] = oldCount + 1
		}
	}

	result := make([]attendee.PackageState, 0)
	for name, count := range counts {
		result = append(result, attendee.PackageState{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func tstPackagesFromPackagesList(asList []attendee.PackageState) string {
	resultList := make([]string, 0)
	for _, entry := range asList {
		if entry.Count == 0 {
			entry.Count = 1
		}
		for i := 0; i < entry.Count; i++ {
			resultList = append(resultList, entry.Name)
		}
	}
	sort.Strings(resultList)
	return strings.Join(resultList, ",")
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
	return tstRegisterAttendeeWithToken(t, testcase, tstValidStaffToken(t, 1))
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
	reason := tstGuessDuesReason(amount, comment)
	return tstValidAttendeeDuesWithReason(amount, comment, reason)
}

func tstGuessDuesReason(amount int64, comment string) string {
	// try to guess package list and manual dues based on amount (since we keep re-using the same package combinations with unique prices)
	// for the few test cases where this doesn't work, must specify the reason yourself
	pkgList := ""
	switch amount {
	default:
		pkgList = `{"name":"attendance","count":1},{"name":"room-none","count":1},{"name":"sponsor2","count":1},{"name":"stage","count":1}`
	}

	reason := ""
	if comment == "dues adjustment due to change in status or selected packages" ||
		comment == "void unpaid dues on cancel" ||
		comment == "admin info change" ||
		comment == "remove dues balance - status changed to deleted" ||
		comment == "remove dues balance - status changed to new" ||
		comment == "remove dues balance - status changed to waiting" {
		// normal package change or cancellation/deletion
		switch amount {
		case 13500:
			reason = fmt.Sprintf(`{"packages_list":[%s],"manual_dues":{"admin":{"amount":-12000,"description":"we owe you this from last year"}}}`, pkgList)
		case 33500:
			reason = fmt.Sprintf(`{"packages_list":[%s],"manual_dues":{"admin":{"amount":8000,"description":"you still need to pay for last year"}}}`, pkgList)
		default:
			reason = fmt.Sprintf(`{"packages_list":[%s],"manual_dues":{}}`, pkgList)
		}
	} else {
		// assume manual admin dues with provided comment
		reason = fmt.Sprintf(`{"packages_list":[%s],"manual_dues":{"admin":{"amount":%d,"description":"%s"}}}`, pkgList, amount, comment)
	}
	return reason
}

func tstValidAttendeeDuesWithReason(amount int64, comment string, reason string) paymentservice.Transaction {
	return paymentservice.Transaction{
		TransactionIdentifier: "",
		DebitorID:             1,
		TransactionType:       paymentservice.Due,
		Method:                paymentservice.Internal,
		Amount: paymentservice.Amount{
			Currency:  "EUR",
			GrossCent: amount,
			VatRate:   19.0,
		},
		Comment:       comment,
		Status:        paymentservice.Valid,
		EffectiveDate: "2022-12-08",
		DueDate:       "2022-12-22",
		StatusHistory: nil, // TODO
		Reason:        reason,
	}
}

func tstNewStatusMail(testcase string, newStatus status.Status) mailservice.MailSendDto {
	result := mailservice.MailSendDto{
		CommonID: "change-status-" + string(newStatus),
		Lang:     "en-US",
		To:       []string{"jsquirrel_github_9a6d@packetloss.de"},
		Variables: map[string]string{
			"badge_number":               "1",
			"badge_number_with_checksum": "1C",
			"nickname":                   "BlackCheetah",
			"email":                      "jsquirrel_github_9a6d@packetloss.de",
			"reason":                     "",
			"remaining_dues":             "EUR 0.00",
			"total_dues":                 "EUR 255.00",
			"pending_payments":           "EUR 0.00",
			"due_date":                   "",
			"regsys_url":                 "http://localhost:10000/register",

			// room group variables
			"room_group_member":       "TODO room group member nickname",
			"room_group_member_email": "TODO room group member email",
			"room_group_name":         "TODO room group name",
			"room_group_owner":        "TODO room group owner nickname",
			"room_group_owner_email":  "TODO room group owner email",

			// other stuff
			"confirm_link": "TODO confirmation link",
			"new_email":    "TODO email change new email",
		},
	}
	if newStatus == status.New {
		result.Variables["total_dues"] = "EUR 0.00"
	}
	if newStatus == status.Approved {
		result.Variables["remaining_dues"] = "EUR 255.00"
		result.Variables["due_date"] = "22.12.2022"
	}
	if newStatus == status.PartiallyPaid {
		result.Variables["remaining_dues"] = "EUR 100.00"
		result.Variables["due_date"] = "22.12.2022"
	}
	if newStatus == status.Paid {
		result.Variables["remaining_dues"] = "EUR 0.00"
	}
	if newStatus == status.Cancelled {
		result.Variables["total_dues"] = "EUR 0.00"
		result.Variables["remaining_dues"] = "EUR 0.00"
		result.Variables["reason"] = testcase
	}
	if newStatus == status.Waiting {
		result.Variables["total_dues"] = "EUR 0.00"
	}
	return result
}

func tstGuestMail(testcase string) mailservice.MailSendDto {
	result := tstNewStatusMail(testcase, "paid")
	result.CommonID = "guest"
	result.Variables["total_dues"] = "EUR 0.00"
	return result
}

func tstNewStatusMailWithAmounts(testcase string, newStatus status.Status, remaining float64, total float64) mailservice.MailSendDto {
	result := tstNewStatusMail(testcase, newStatus)
	result.Variables["remaining_dues"] = fmt.Sprintf("EUR %0.2f", remaining)
	result.Variables["total_dues"] = fmt.Sprintf("EUR %0.2f", total)
	if remaining > 0 {
		result.Variables["due_date"] = "22.12.2022"
	}
	return result
}

func tstNewCancelMail(testcase string, reason string, alreadyPaid float64) mailservice.MailSendDto {
	result := tstNewStatusMailWithAmounts(testcase, status.Cancelled, 0, alreadyPaid)
	result.Variables["reason"] = reason
	return result
}
