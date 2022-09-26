package infoctl

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/filterhelper"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Create(server chi.Router) {
	server.Get("/", filterhelper.BuildUnauthenticatedHandler("800ms", healthHandler))
	server.Get("/info/health", filterhelper.BuildUnauthenticatedHandler("800ms", healthHandler))
}

func healthHandler(_ context.Context, w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}
