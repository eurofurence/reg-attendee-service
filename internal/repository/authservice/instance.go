package authservice

import (
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

var activeInstance AuthService

func Create() (err error) {
	if config.AuthServiceBaseUrl() != "" {
		activeInstance, err = newClient()
		return err
	} else {
		aulogging.Logger.NoCtx().Warn().Printf("service.auth_service not configured. Will skip online userinfo checks (not useful for production!)")
		activeInstance = newMock()
		return nil
	}
}

func CreateMock() Mock {
	instance := newMock()
	activeInstance = instance
	return instance
}

func Get() AuthService {
	return activeInstance
}
