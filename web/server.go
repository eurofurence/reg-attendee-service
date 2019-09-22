package web

import (
	"github.com/gorilla/mux"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/controller/attendeectl"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/controller/fallbackctl"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/controller/infoctl"
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

	attendeectl.RestDispatcher(router.PathPrefix("/api/rest").Subrouter())

	fallbackctl.ErrorDispatcher(router)
}
