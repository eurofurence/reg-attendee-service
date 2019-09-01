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
	router := mux.NewRouter().StrictSlash(true)
	dispatcher(router)
	log.Fatal(http.ListenAndServe(config.ServerAddr(), router))
}

func dispatcher(router *mux.Router) {
	infoctl.Dispatcher(router.PathPrefix("/info").Subrouter())

	restRouter := router.PathPrefix("/api/rest").Subrouter()
	attendeectl.RestDispatcher(restRouter)
}
