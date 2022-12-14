package mailservice

import (
	"context"
	"errors"
)

type MailService interface {
	SendEmail(ctx context.Context, request MailSendDto) error
}

var (
	DownstreamError = errors.New("downstream unavailable - see log for details")
)

type MailSendDto struct {
	CommonID  string            `json:"cid"`
	Lang      string            `json:"lang"`
	To        []string          `json:"to"`
	Cc        []string          `json:"cc"`
	Bcc       []string          `json:"bcc"`
	Variables map[string]string `json:"variables"`
}
