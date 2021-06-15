package infoctl

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func Create(server chi.Router) {
	server.Get("/info/health", filterhelper.BuildUnauthenticatedNologgingHandler("800ms", healthHandler))
	server.Get("/info/timeout", filterhelper.BuildUnauthenticatedHandler("800ms", timeoutHandler))
}

func healthHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}

// TODO when request timeouts are implemented, move this to testing code

func timeoutHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	_, _ = fmt.Fprintf(w, "OK")
}
