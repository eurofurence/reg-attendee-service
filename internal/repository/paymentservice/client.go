package paymentservice

import (
	"context"
	"fmt"
	aurestbreaker "github.com/StephanHCB/go-autumn-restclient-circuitbreaker/implementation/breaker"
	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	aurestlogging "github.com/StephanHCB/go-autumn-restclient/implementation/requestlogging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"net/http"
	"time"
)

type Impl struct {
	client  aurestclientapi.Client
	baseUrl string
}

func requestManipulator(ctx context.Context, r *http.Request) {
	// TODO do we ever need to pass on the user token instead?
	r.Header.Add(media.HeaderXApiKey, config.FixedApiToken())
}

func newClient() (PaymentService, error) {
	httpClient, err := auresthttpclient.New(0, nil, requestManipulator)
	if err != nil {
		return nil, err
	}

	requestLoggingClient := aurestlogging.New(httpClient)

	circuitBreakerClient := aurestbreaker.New(requestLoggingClient,
		"payment-service-breaker",
		10,
		2*time.Minute,
		30*time.Second,
		15*time.Second,
	)

	return &Impl{
		client:  circuitBreakerClient,
		baseUrl: config.PaymentServiceBaseUrl(),
	}, nil
}

func errByStatus(err error, status int) error {
	if err != nil {
		return err
	}
	if status == http.StatusNotFound {
		return NoSuchDebitor404Error
	}
	if status >= 300 {
		return DownstreamError
	}
	return nil
}

func (i Impl) GetTransactions(ctx context.Context, debitorId uint) ([]Transaction, error) {
	url := fmt.Sprintf("%s/api/rest/v1/transactions?debitor_id=%d", i.baseUrl, debitorId)
	bodyDto := TransactionResponse{}
	response := aurestclientapi.ParsedResponse{
		Body: &bodyDto,
	}
	err := i.client.Perform(ctx, http.MethodGet, url, nil, &response)
	return bodyDto.Payload, errByStatus(err, response.Status)
}

func (i Impl) AddTransaction(ctx context.Context, transaction Transaction) error {
	url := fmt.Sprintf("%s/api/rest/v1/transactions", i.baseUrl)
	response := aurestclientapi.ParsedResponse{}
	err := i.client.Perform(ctx, http.MethodPost, url, transaction, &response)
	return errByStatus(err, response.Status)
}
