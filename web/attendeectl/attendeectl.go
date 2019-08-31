package attendeectl

import (
	"encoding/json"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"net/http"
)

const contentTypeApplicationJson = "application/json"

func RestDispatcher(router *mux.Router) {
	router.HandleFunc("/v1/attendees", newAttendeeHandler).Methods(http.MethodPut)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", getAttendeeHandler).Methods(http.MethodGet)
	router.HandleFunc("/v1/attendees/{id:[1-9][0-9]*}", updateAttendeeHandler).Methods(http.MethodPost)
}

// TODO this part is only a very preliminary example

func newAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headers.Location, r.RequestURI + "/1")
	w.WriteHeader(http.StatusCreated)
}

func getAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	m := map[string]string {
		"id": vars["id"],
	}
	w.Header().Add(headers.ContentType, contentTypeApplicationJson)
	_ = json.NewEncoder(w).Encode(m)
}

func updateAttendeeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(headers.Location, r.RequestURI)
}
