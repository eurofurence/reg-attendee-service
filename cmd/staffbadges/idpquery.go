package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type IDPLookupResult struct {
	StaffIDs    []string
	DirectorIDs []string
}

func lookupUserIDs(config Config) (IDPLookupResult, error) {
	result := IDPLookupResult{}

	staffIDs, err := lookupUserIDsForGroup(config.IDPUrl, config.StaffGroupID, config.Token)
	if err != nil {
		return result, err
	}

	directorIDs, err := lookupUserIDsForGroup(config.IDPUrl, config.DirectorsGroupID, config.Token)
	if err != nil {
		return result, err
	}

	// remove staffIDs that are also directors
	result.DirectorIDs = make([]string, 0)
	for k, _ := range directorIDs {
		if _, ok := staffIDs[k]; ok {
			log.Printf("removing staff %s who is also director\n", k)
			delete(staffIDs, k)
		}
		result.DirectorIDs = append(result.DirectorIDs, k)
	}

	result.StaffIDs = make([]string, 0)
	for k, _ := range staffIDs {
		result.StaffIDs = append(result.StaffIDs, k)
	}

	log.Printf("we now have %d directors and %d staff", len(result.DirectorIDs), len(result.StaffIDs))
	return result, nil
}

// --- internals ---

type groupUsersResponse struct {
	Data []struct {
		GroupId string `json:"group_id"`
		UserId  string `json:"user_id"`
		Level   string `json:"level"`
	} `json:"data"`
	Links struct {
		First string      `json:"first"`
		Last  interface{} `json:"last"`
		Prev  interface{} `json:"prev"`
		Next  string      `json:"next"` // if null, no next page
	} `json:"links"`
	Meta struct {
		CurrentPage int    `json:"current_page"`
		From        int    `json:"from"`
		Path        string `json:"path"`
		PerPage     int    `json:"per_page"`
		To          int    `json:"to"`
	} `json:"meta"`
}

func requestGroupUsers(url string, token string) (groupUsersResponse, error) {
	result := groupUsersResponse{}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("error creating request: %s", err.Error())
	}
	request.Header.Add("Authorization", "Bearer "+token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return result, fmt.Errorf("error making request: %s", err.Error())
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("error reading body: %s", err.Error())
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("error parsing body: %s", err.Error())
	}

	return result, nil
}

func lookupUserIDsForGroup(baseUrl string, grpId string, token string) (map[string]struct{}, error) {
	log.Println("querying idp for group " + grpId)

	result := make(map[string]struct{})

	url := fmt.Sprintf("%s/api/v1/groups/%s/users?page=1", baseUrl, grpId)
	response, err := requestGroupUsers(url, token)
	if err != nil {
		return result, err
	}
	for _, entry := range response.Data {
		result[entry.UserId] = struct{}{}
	}
	for response.Links.Next != "" {
		response, err = requestGroupUsers(response.Links.Next, token)
		if err != nil {
			return result, err
		}
		for _, entry := range response.Data {
			result[entry.UserId] = struct{}{}
		}
	}

	log.Printf("successfully read %d member ids in group %s\n", len(result), grpId)
	return result, nil
}
