package infoctl

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/filterhelper"
	"net/http"
)

func Dispatcher(router *mux.Router) {
	router.HandleFunc("/health", filterhelper.BuildUnauthenticatedNologgingHandler("800ms", healthHandler)).Methods(http.MethodGet)
}

func healthHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}
