package web

import (
	"github.com/eurofurence/reg-attendee-service/web/controller/countdownctl"
	"github.com/gorilla/mux"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/web/controller/attendeectl"
	"github.com/eurofurence/reg-attendee-service/web/controller/fallbackctl"
	"github.com/eurofurence/reg-attendee-service/web/controller/infoctl"
	"net/http"
)

func StartWebserverAndNeverReturn() {
	logging.NoCtx().Info("Building routers...")
	router := CreateRouter();
	logging.NoCtx().Info("Listening...")
	logging.NoCtx().Fatal(http.ListenAndServe(config.ServerAddr(), router))
}

// you can use this from tests
func CreateRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	dispatcher(router)
	return router
}

func dispatcher(router *mux.Router) {
	infoctl.Dispatcher(router.PathPrefix("/info").Subrouter())

	apiSubrouter := router.PathPrefix("/api/rest").Subrouter()
	attendeectl.RestDispatcher(apiSubrouter)
	countdownctl.RestDispatcher(apiSubrouter)

	fallbackctl.ErrorDispatcher(router)
}
