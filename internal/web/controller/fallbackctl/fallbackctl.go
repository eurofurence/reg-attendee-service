package fallbackctl

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Create(server chi.Router) {
	server.HandleFunc("/*", fallbackErrorHandler)
}

func fallbackErrorHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
