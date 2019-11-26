package infoctl

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
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

// TODO when request timeouts are implemented, move this to testing code

func timeoutHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	_, _ = fmt.Fprintf(w, "OK")
}
