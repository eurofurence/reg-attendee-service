package infoctl

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/filterhelper"
	"net/http"
	"time"
)

func Dispatcher(router *mux.Router) {
	router.HandleFunc("/health", filterhelper.BuildUnauthenticatedNologgingHandler("800ms", healthHandler)).Methods(http.MethodGet)
	router.HandleFunc("/timeout", filterhelper.BuildUnauthenticatedHandler("800ms", timeoutHandler)).Methods(http.MethodGet)
}

func healthHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}

func timeoutHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	_, _ = fmt.Fprintf(w, "OK")
}
