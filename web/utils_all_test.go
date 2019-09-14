package web

import (
	"github.com/go-http-utils/headers"
	"io/ioutil"
	"log"
	"net/http"
	"rexis/rexis-go-attendee/web/util/media"
	"strings"
)

// placing these here because they are package global

type tstWebResponse struct {
	status int
	body string
	contentType string
	location string
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
		status: status,
		body:   string(body),
		contentType: ct,
		location: loc,
	}
}

func tstPerformGet(relativeUrlWithLeadingSlash string) tstWebResponse {
	response, err := http.Get(ts.URL + relativeUrlWithLeadingSlash)
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}

func tstPerformPut(relativeUrlWithLeadingSlash string, requestBody string) tstWebResponse {
	request, err := http.NewRequest(http.MethodPut, ts.URL + relativeUrlWithLeadingSlash, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set(headers.ContentType, media.ContentTypeApplicationJson)
	response, err := http.DefaultClient.Do(request)
	return tstWebResponseFromResponse(response)
}

func tstPerformPost(relativeUrlWithLeadingSlash string, requestBody string) tstWebResponse {
	response, err := http.Post(ts.URL + relativeUrlWithLeadingSlash, media.ContentTypeApplicationJson, strings.NewReader(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	return tstWebResponseFromResponse(response)
}
