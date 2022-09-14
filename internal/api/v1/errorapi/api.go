package errorapi

import "net/url"

type ErrorDto struct {
	Timestamp string     `json:"timestamp"`
	RequestId string     `json:"requestid"`
	Message   string     `json:"message"`
	Details   url.Values `json:"details"`
}
