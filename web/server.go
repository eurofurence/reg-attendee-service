package web

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/web/attendeectl"
	"rexis/rexis-go-attendee/web/infoctl"
)

func StartWebserverAndNeverReturn() {
	log.Print("Building routers...")
	router := CreateRouter();
	log.Print("Listening...")
	log.Fatal(http.ListenAndServe(config.ServerAddr(), router))
}

// you can use this from tests
func CreateRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	dispatcher(router)
	return router
}

func dispatcher(router *mux.Router) {
	infoctl.Dispatcher(router.PathPrefix("/info").Subrouter())

	restRouter := router.PathPrefix("/api/rest").Subrouter()
	attendeectl.RestDispatcher(restRouter)
}
