package main

import (
	"log"
	"net/http"
	"rexis/rexis-go-attendee/web/attendeectl"
	"rexis/rexis-go-attendee/web/infoctl"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	dispatcher(router)
	log.Fatal(http.ListenAndServe(":9091", router))
}

func dispatcher(router *mux.Router) {
	infoctl.Dispatcher(router.PathPrefix("/info").Subrouter())

	restRouter := router.PathPrefix("/api/rest").Subrouter()
	attendeectl.RestDispatcher(restRouter)
}
