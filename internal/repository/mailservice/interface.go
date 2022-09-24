package mailservice

import (
	"context"
	"errors"
)

type MailService interface {
	SendEmail(ctx context.Context, request TemplateRequestDto) error
}

var (
	DownstreamError = errors.New("downstream unavailable - see log for details")
)

type TemplateRequestDto struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	Email     string            `json:"email"`
}
