package web

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

// placing these here because they are package global

func tstAssertStringEqual(t *testing.T, message string, expected string, actual string) {
	if (expected != actual) {
		t.Errorf("Health endpoint did not report OK.\nExpected:\n%v\nActual:\n%v\n", expected, actual)
	}
}

func tstPerformGetReturnBody(relativeUrlWithLeadingSlash string) string {
	res, err := http.Get(ts.URL + relativeUrlWithLeadingSlash)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}
