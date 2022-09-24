package web

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/adminctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/attendeectl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/countdownctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/fallbackctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/infoctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/statusctl"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Create() chi.Router {
	logging.NoCtx().Info("Building routers...")
	server := chi.NewRouter()
	countdownctl.Create(server)
	attendeectl.Create(server)
	adminctl.Create(server)
	statusctl.Create(server)
	infoctl.Create(server)
	fallbackctl.Create(server)
	return server
}

func StartWebserverAndNeverReturn(server chi.Router) {
	address := config.ServerAddr()
	logging.NoCtx().Info("Listening on " + address)
	err := http.ListenAndServe(address, server)
	if err != nil {
		logging.NoCtx().Error(err)
	}
}