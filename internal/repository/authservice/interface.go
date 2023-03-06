package authservice

import (
	"context"
	"errors"
)

const TraceIdHeader = "X-Request-Id" // repeat for circular ref avoidance

type AuthService interface {
	IsEnabled() bool

	UserInfo(ctx context.Context) (UserInfoResponse, error)
}

var (
	UnauthorizedError = errors.New("got unauthorized from userinfo endpoint")
	DownstreamError   = errors.New("downstream unavailable - see log for details")
)

type UserInfoResponse struct {
	Audiences     []string `json:"audiences"`
	Subject       string   `json:"subject"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
}
