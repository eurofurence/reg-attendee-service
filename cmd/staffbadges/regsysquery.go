package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/admin"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/errorapi"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

type BadgeLookupResult struct {
	StaffBadges    map[uint]status.Status
	DirectorBadges map[uint]status.Status
}

func lookupBadgeNumbers(idpLookupResult IDPLookupResult, config Config) (BadgeLookupResult, error) {
	log.Println("querying regsys for badge numbers and status")

	log.Println("querying regsys for badge numbers and status for directors")
	directors, err := lookupBadgeNumbersAndStatus(idpLookupResult.DirectorIDs, config)
	if err != nil {
		return BadgeLookupResult{}, err
	}

	log.Println("querying regsys for badge numbers and status for staff")
	staff, err := lookupBadgeNumbersAndStatus(idpLookupResult.StaffIDs, config)
	if err != nil {
		return BadgeLookupResult{}, err
	}

	result := BadgeLookupResult{
		StaffBadges:    staff,
		DirectorBadges: directors,
	}

	return result, nil
}

type AttInfo struct {
	Nickname string
	Staff    bool
	Director bool
}

func listAttendees(config Config) (map[uint]AttInfo, error) {
	result := make(map[uint]AttInfo)

	log.Println("searching regsys for all attendees and their flags")

	reqBody := attendee.AttendeeSearchCriteria{
		MatchAny: []attendee.AttendeeSearchSingleCriterion{
			{
				Status: []status.Status{
					status.New,
					status.Approved,
					status.PartiallyPaid,
					status.Paid,
					status.CheckedIn,
					status.Waiting,
					status.Cancelled,
				},
			},
		},
		FillFields: []string{
			"id",
			"nickname",
			"flags",
		},
	}
	searchResult, err := findAttendees(reqBody, config.RegsysUrl, config.Auth, config.Jwt)
	if err != nil {
		return result, err
	}

	if len(searchResult.Attendees) < 100 {
		return result, errors.New("find attendees returned too few results, this cannot be")
	}

	for _, att := range searchResult.Attendees {
		if att.Nickname == nil || *att.Nickname == "" {
			return result, fmt.Errorf("no nickname for reg id %d -- aborting", att.Id)
		}
		result[att.Id] = AttInfo{
			Nickname: *att.Nickname,
			Staff:    slices.Contains(att.FlagsList, "staff"),
			Director: slices.Contains(att.FlagsList, "director"),
		}
	}

	return result, nil
}

func addAdminFlag(badgeNo uint, flag string, config Config) error {
	adminInfo, err := readAdminInfo(badgeNo, config.RegsysUrl, config.Auth, config.Jwt)
	if err != nil {
		return err
	}

	if !strings.Contains(adminInfo.Flags, flag) {
		if adminInfo.Flags == "" {
			adminInfo.Flags = flag
		} else {
			adminInfo.Flags = adminInfo.Flags + "," + flag
		}
	}

	err = updateAdminInfo(badgeNo, config.RegsysUrl, adminInfo, config.Auth, config.Jwt)
	if err != nil {
		return err
	}

	return nil
}

// --- internals ---

func lookupBadgeNumbersAndStatus(identities []string, config Config) (map[uint]status.Status, error) {
	result := make(map[uint]status.Status)

	for _, ident := range identities {
		badge, err := lookupBadgeNumber(ident, config.RegsysUrl, config.Auth, config.Jwt)
		if err != nil {
			return result, err
		}

		if badge > 0 {
			regStatus, err := lookupStatus(badge, config.RegsysUrl, config.Auth, config.Jwt)
			if err != nil {
				return result, err
			}

			result[badge] = regStatus
		}
	}

	return result, nil
}

func lookupStatus(badge uint, baseUrl string, token string, jwt string) (status.Status, error) {
	url := fmt.Sprintf("%s/api/rest/v1/attendees/%d/status", baseUrl, badge)
	httpStatus, body, err := regsysGetNoBody(url, token, jwt)
	if err != nil {
		return "", err
	}

	if httpStatus == 404 {
		log.Printf("no status for badge %d", badge)
		return "", fmt.Errorf("failed to read status after finding attendee for badge %d", badge)
	}

	if httpStatus != 200 {
		result := errorapi.ErrorDto{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", fmt.Errorf("error parsing body after non-200 error: " + err.Error())
		}

		return "", fmt.Errorf("status lookup failed with unexpected http status %d message %s requestid %s", httpStatus, result.Message, result.RequestId)
	}

	result := status.StatusDto{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing body: " + err.Error())
	}
	return result.Status, nil
}

func lookupBadgeNumber(identity string, baseUrl string, token string, jwt string) (uint, error) {
	url := fmt.Sprintf("%s/api/rest/v1/attendees/identity/%s", baseUrl, identity)
	httpStatus, body, err := regsysGetNoBody(url, token, jwt)
	if err != nil {
		return 0, err
	}

	if httpStatus == 404 {
		log.Printf("no registration for identity %s", identity)
		return 0, nil
	}

	if httpStatus != 200 {
		result := errorapi.ErrorDto{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return 0, fmt.Errorf("error parsing body after non-200 error: " + err.Error())
		}

		return 0, fmt.Errorf("badge lookup failed with unexpected http status %d message %s requestid %s", httpStatus, result.Message, result.RequestId)
	}

	result := attendee.AttendeeIdList{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, fmt.Errorf("error parsing body: " + err.Error())
	}

	if len(result.Ids) == 0 {
		log.Printf("no registration for identity %s", identity)
		return 0, nil
	}
	if len(result.Ids) > 1 {
		log.Printf("WARNING, multiple registrations for identity %s, badge numbers %v -- skipping", identity, result.Ids)
		return 0, nil
	}

	return result.Ids[0], nil
}

func findAttendees(reqBody attendee.AttendeeSearchCriteria, baseUrl string, token string, jwt string) (attendee.AttendeeSearchResultList, error) {
	result := attendee.AttendeeSearchResultList{}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return result, err
	}

	url := fmt.Sprintf("%s/api/rest/v1/attendees/find", baseUrl)
	httpStatus, responseBody, err := regsysPost(url, reqBodyBytes, token, jwt)
	if err != nil {
		return result, err
	}

	if httpStatus != 200 {
		errResult := errorapi.ErrorDto{}
		err = json.Unmarshal(responseBody, &errResult)
		if err != nil {
			return result, fmt.Errorf("error parsing body after non-200 error: " + err.Error())
		}

		return result, fmt.Errorf("badge lookup failed with unexpected http status %d message %s requestid %s", httpStatus, errResult.Message, errResult.RequestId)
	}

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return result, fmt.Errorf("error parsing body after 200: " + err.Error())
	}

	return result, nil
}

func readAdminInfo(badge uint, baseUrl string, token string, jwt string) (admin.AdminInfoDto, error) {
	result := admin.AdminInfoDto{}

	url := fmt.Sprintf("%s/api/rest/v1/attendees/%d/admin", baseUrl, badge)
	httpStatus, body, err := regsysGetNoBody(url, token, jwt)
	if err != nil {
		return result, err
	}

	if httpStatus == 404 {
		log.Printf("no admin info for badge %d", badge)
		return result, fmt.Errorf("failed to read admin info after finding attendee for badge %d", badge)
	}

	if httpStatus != 200 {
		errResult := errorapi.ErrorDto{}
		err = json.Unmarshal(body, &errResult)
		if err != nil {
			return result, fmt.Errorf("error parsing body after non-200 error: " + err.Error())
		}

		return result, fmt.Errorf("admin info lookup failed with unexpected http status %d message %s requestid %s", httpStatus, errResult.Message, errResult.RequestId)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("error parsing body: " + err.Error())
	}
	return result, nil
}

func updateAdminInfo(badge uint, baseUrl string, info admin.AdminInfoDto, token string, jwt string) error {
	url := fmt.Sprintf("%s/api/rest/v1/attendees/%d/admin", baseUrl, badge)

	requestBody, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to json encode admin info: %s", err.Error())
	}

	httpStatus, responseBody, err := regsysPut(url, requestBody, token, jwt)
	if err != nil {
		return err
	}

	if httpStatus != 204 {
		errResult := errorapi.ErrorDto{}
		err = json.Unmarshal(responseBody, &errResult)
		if err != nil {
			return fmt.Errorf("error parsing body after non-204 error: " + err.Error())
		}

		return fmt.Errorf("admin info lookup failed with unexpected http status %d message %s requestid %s", httpStatus, errResult.Message, errResult.RequestId)
	}

	return nil
}

// --- low level internal ---

func regsysGetNoBody(url string, token string, jwt string) (int, []byte, error) {
	return regsysRequest(http.MethodGet, url, nil, token, jwt)
}

func regsysPost(url string, requestBody []byte, token string, jwt string) (int, []byte, error) {
	return regsysRequest(http.MethodPost, url, bytes.NewReader(requestBody), token, jwt)
}

func regsysPut(url string, requestBody []byte, token string, jwt string) (int, []byte, error) {
	return regsysRequest(http.MethodPut, url, bytes.NewReader(requestBody), token, jwt)
}

func regsysRequest(method string, url string, requestBody io.Reader, token string, jwt string) (int, []byte, error) {
	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("error creating request: " + err.Error())
	}
	request.AddCookie(&http.Cookie{
		Name:     "JWT",
		Value:    jwt,
		Domain:   "localhost",
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	request.AddCookie(&http.Cookie{
		Name:     "AUTH",
		Value:    token,
		Domain:   "localhost",
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	request.Header.Add("X-Admin-Request", "available")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("error making request: " + err.Error())
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, []byte{}, fmt.Errorf("error reading body: " + err.Error())
	}

	return response.StatusCode, body, nil
}
