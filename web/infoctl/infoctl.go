package infoctl

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func Dispatcher(router *mux.Router) {
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
}

// TODO this part is only a very preliminary example

func healthHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}
