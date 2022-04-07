package fallbackctl

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/web/filter/ctxvalues"
	"github.com/eurofurence/reg-attendee-service/web/filter/filterhelper"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Create(server chi.Router) {
	server.HandleFunc("/*", filterhelper.BuildUnauthenticatedHandler("1s", fallbackErrorHandler))
}

func fallbackErrorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	ctxvalues.SetHttpStatus(ctx, http.StatusNotFound)
}
