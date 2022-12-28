package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/media"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

// contract test consumer side for payment service

func TestPaymentServiceConsumer(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "AttendeeService",
		Provider: "PaymentService",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// build models
	var (
		time1DuesBooked     = time.Date(2022, 12, 21, 19, 58, 10, 0, time.UTC)
		time2PaymentCreated = time.Date(2022, 12, 27, 20, 12, 31, 0, time.UTC)
		time2PaymentPaid    = time.Date(2022, 12, 27, 22, 0, 0, 0, time.UTC)

		duesTransaction = paymentservice.Transaction{
			DebitorID:             1,
			TransactionIdentifier: "EF2022-000001-1221-195810-1234",
			TransactionType:       "due",
			Method:                "internal",
			Amount: paymentservice.Amount{
				Currency:  "EUR",
				GrossCent: 19700,
				VatRate:   19.0,
			},
			Comment:       "dues adjustment due to change in status or selected packages",
			Status:        "valid",
			EffectiveDate: "2022-12-21",
			DueDate:       "2023-01-04",
			CreationDate:  time1DuesBooked,
		}
		paymentTransaction = paymentservice.Transaction{
			DebitorID:             1,
			TransactionIdentifier: "EF2022-000001-1227-201231-2345",
			TransactionType:       "payment",
			Method:                "credit",
			Amount: paymentservice.Amount{
				Currency:  "EUR",
				GrossCent: 19700,
				VatRate:   19.0,
			},
			Comment:       "credit card payment received via adapter",
			Status:        "pending",
			EffectiveDate: "2022-12-27",
			DueDate:       "2022-12-27",
			CreationDate:  time2PaymentCreated,
			StatusHistory: []paymentservice.StatusHistory{
				{
					Status:     "tentative",
					Comment:    "credit card payment link created",
					ChangedBy:  "api",
					ChangeDate: time2PaymentPaid,
				},
			},
		}
		duesTransaction2 = paymentservice.Transaction{
			DebitorID:             1,
			TransactionIdentifier: "EF2022-000001-1222-195816-8888",
			TransactionType:       "due",
			Method:                "internal",
			Amount: paymentservice.Amount{
				Currency:  "EUR",
				GrossCent: 500,
				VatRate:   7.0,
			},
			Comment:       "dues adjustment due to change in status or selected packages",
			Status:        "valid",
			EffectiveDate: "2022-12-22",
			DueDate:       "2023-01-05",
			CreationDate:  time1DuesBooked,
		}
	)

	// test case (consumer side)
	var test = func() (err error) {
		config.Configuration().Service.PaymentService = fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		if err := paymentservice.Create(); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		transactions, err := paymentservice.Get().GetTransactions(ctx, 1)
		if err != nil {
			return err
		}

		if len(transactions) != 2 {
			return errors.New("got wrong length of response")
		}
		if !assert.Equal(t, duesTransaction, transactions[0]) {
			return errors.New("first transaction wrong")
		}
		if !assert.Equal(t, paymentTransaction, transactions[1]) {
			return errors.New("second transaction wrong")
		}

		err = paymentservice.Get().AddTransaction(ctx, duesTransaction2)
		if err != nil {
			return err
		}

		return nil
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		// this is the identifier of the state handler that will be called on the other side
		Given("Attendee 1 exists with one dues transaction and a matching pending payment").
		UponReceiving("A request to get their transactions").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/api/rest/v1/transactions"),
			Query:  dsl.MapMatcher{"debitor_id": dsl.String("1")},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String(media.ContentTypeApplicationJson)},
			Body: paymentservice.TransactionResponse{
				Payload: []paymentservice.Transaction{
					duesTransaction,
					paymentTransaction,
				},
			},
		})

	pact.
		AddInteraction().
		Given("Attendee 1 exists in any state").
		UponReceiving("A request to create an additional dues transaction").
		WithRequest(dsl.Request{
			Method:  "POST",
			Path:    dsl.String("/api/rest/v1/transactions"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String(media.ContentTypeApplicationJson)},
			// slight deviation, here we specify a transaction identifier, so we can ensure this is the one that's returned as location
			Body: duesTransaction2,
		}).
		WillRespondWith(dsl.Response{
			Status:  201,
			Headers: dsl.MapMatcher{"Location": dsl.String("EF2022-000001-1222-195816-8888")},
			Body:    duesTransaction2,
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
