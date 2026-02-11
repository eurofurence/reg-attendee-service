package mailservice

import (
	"context"
	"fmt"
	"net/http"
	"time"

	aurestbreaker "github.com/StephanHCB/go-autumn-restclient-circuitbreaker/implementation/breaker"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/telemetry"
	"github.com/eurofurence/reg-attendee-service/internal/web/middleware"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
)

type Impl struct {
	client  aurestclientapi.Client
	baseUrl string
}

func requestManipulator(ctx context.Context, r *http.Request) {
	r.Header.Add(media.HeaderXApiKey, config.FixedApiToken())
	r.Header.Add(middleware.TraceIdHeader, ctxvalues.RequestId(ctx))
}

func newClient() (MailService, error) {
	httpClient := telemetry.NewHttpClient(requestManipulator)

	requestLoggingClient := aurestlogging.New(httpClient)

	circuitBreakerClient := aurestbreaker.New(requestLoggingClient,
		"mail-service-breaker",
		10,
		2*time.Minute,
		30*time.Second,
		15*time.Second,
	)

	return &Impl{
		client:  circuitBreakerClient,
		baseUrl: config.MailServiceBaseUrl(),
	}, nil
}

func errByStatus(err error, status int) error {
	if err != nil {
		return err
	}
	if status >= 300 {
		return DownstreamError
	}
	return nil
}

func (i Impl) SendEmail(ctx context.Context, request MailSendDto) error {
	url := fmt.Sprintf("%s/api/v1/mail", i.baseUrl)
	response := aurestclientapi.ParsedResponse{}
	err := i.client.Perform(ctx, http.MethodPost, url, request, &response)
	return errByStatus(err, response.Status)
}
