package fallbackctl

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxvalues"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/filterhelper"
	"net/http"
)

func ErrorDispatcher(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(filterhelper.BuildUnauthenticatedHandler("1s", fallbackErrorHandler))
}

func fallbackErrorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	ctxvalues.SetHttpStatus(ctx, http.StatusNotFound)
}
