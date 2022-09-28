package mailservice

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

var activeInstance MailService

func Create() (err error) {
	if config.MailServiceBaseUrl() != "" {
		activeInstance, err = newClient()
		return err
	} else {
		aulogging.Logger.NoCtx().Warn().Printf("downstream.mail_service not configured. Using in-memory simulator for mail service (not useful for production!)")
		activeInstance = newMock()
		return nil
	}
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() MailService {
	return activeInstance
}
