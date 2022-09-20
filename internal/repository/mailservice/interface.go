package mailservice

import "context"

type MailService interface {
	SendEmail(ctx context.Context, request TemplateRequestDto) error
}

type TemplateRequestDto struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	Email     string            `json:"email"`
}
