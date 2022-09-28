package paymentservice

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

var activeInstance PaymentService

func Create() (err error) {
	if config.PaymentServiceBaseUrl() != "" {
		activeInstance, err = newClient()
		return err
	} else {
		aulogging.Logger.NoCtx().Warn().Printf("downstream.payment_service not configured. Using in-memory simulator for payment service (not useful for production!)")
		activeInstance = newMock()
		return nil
	}
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() PaymentService {
	return activeInstance
}
