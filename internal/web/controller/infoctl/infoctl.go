package infoctl

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func Create(server chi.Router) {
	server.Get("/", filter.WithTimeout(800*time.Millisecond, healthHandler))
	server.Get("/info/health", filter.WithTimeout(800*time.Millisecond, healthHandler))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "OK")
}
