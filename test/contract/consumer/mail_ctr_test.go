package consumer

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/pact-foundation/pact-go/dsl"
	"log"
	"testing"
)

// contract test consumer side for mail service

func TestMailServiceConsumer(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "MailService",
		Provider: "PaymentService",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// build models
	var (
		sendRequest = mailservice.MailSendDto{
			CommonID: "change-status-cancelled",
			Lang:     "de-DE",
			To:       []string{"person-1-to@not.a.real.domain.com", "person-2-to@not.a.real.domain.com"},
			Cc:       []string{"person-1-cc@not.a.real.domain.com", "person-2-cc@not.a.real.domain.com"},
			Bcc:      []string{"person-1-bcc@not.a.real.domain.com", "person-2-bcc@not.a.real.domain.com"},
			Variables: map[string]string{
				"nickname": "someone",
			},
		}
		requestId = "abcd1234"
	)

	// test case (consumer side)
	var test = func() (err error) {
		config.Configuration().Service.MailService = fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		if err := mailservice.Create(); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctx = ctxvalues.CreateContextWithValueMap(ctx)
		ctxvalues.SetRequestId(ctx, requestId)

		err = mailservice.Get().SendEmail(ctx, sendRequest)
		if err != nil {
			return err
		}

		return nil
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		// this is the identifier of the state handler that will be called on the other side
		Given("The standard templates are present in the database").
		UponReceiving("A request to notify a German speaking attendee of their cancellation").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/api/v1/mail"),
			Headers: dsl.MapMatcher{
				"Content-Type": dsl.String(media.ContentTypeApplicationJson),
				"X-Request-Id": dsl.String(requestId),
				"X-Api-Key":    dsl.String("api-token-for-testing-must-be-pretty-long"),
			},
			Body: sendRequest,
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	// Run the test, verify it did what we expected and capture the contract (writes a test log to logs/pact.log)
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

	// now write out the contract json (by default it goes to subdirectory pacts)
	if err := pact.WritePact(); err != nil {
		log.Fatalf("Error on pact write: %v", err)
	}
}
